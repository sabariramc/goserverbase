package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	baseapp "github.com/sabariramc/goserverbase/v2/app"
	"github.com/sabariramc/goserverbase/v2/errors"
	"github.com/sabariramc/goserverbase/v2/log"
	"github.com/sabariramc/goserverbase/v2/utils"
)

type HttpServer struct {
	*baseapp.BaseApp
	handler *chi.Mux
	docMeta APIDocumentation
	Log     *log.Logger
	c       *HttpServerConfig
}

func New(appConfig HttpServerConfig, loggerConfig log.Config, lMux log.LogMux, errorNotifier errors.ErrorNotifier, auditLogger log.AuditLogWriter) *HttpServer {
	b := baseapp.New(*appConfig.ServerConfig, loggerConfig, lMux, errorNotifier, auditLogger)
	if appConfig.Log.ContentLength <= 0 {
		appConfig.Log.ContentLength = 1024
	}
	h := &HttpServer{
		BaseApp: b,
		handler: chi.NewRouter(),
		docMeta: APIDocumentation{
			Server: make([]DocumentServer, 0),
			Routes: make(APIRoute, 0),
		},
		Log: b.GetLogger(),
		c:   &appConfig,
	}
	ctx := b.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParams(appConfig.ServiceName))
	h.SetupRouter(ctx)
	return h
}

func (b *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.handler.ServeHTTP(w, r)
}

func (b *HttpServer) GetAPIDocument() APIDocumentation {
	return b.docMeta
}

func (b *HttpServer) GetRouter() *chi.Mux {
	return b.handler
}

func (b *HttpServer) StartHttpServer() {
	b.Log.Notice(context.TODO(), fmt.Sprintf("Server starting at %v", b.GetPort()), nil)
	err := http.ListenAndServe(b.GetPort(), b)
	b.Log.Emergency(context.Background(), "Server crashed", nil, err)
}

type Filter struct {
	PageNo int64  `json:"pageNo" schema:"pageNo"`
	Limit  int64  `json:"limit" schema:"limit"`
	SortBy string `json:"sortBy" schema:"sortBy"`
	Asc    *bool  `json:"asc" schema:"asc"`
}

func SetDefaultPagination(filter interface{}, defaultSortBy string) error {
	var defaultFilter Filter
	err := utils.StrictJsonTransformer(filter, &defaultFilter)
	if err != nil {
		return fmt.Errorf("baseapp.SetDefault : %w", err)
	}
	if defaultFilter.PageNo <= 0 {
		defaultFilter.PageNo = 1
	}
	if defaultFilter.Limit <= 0 {
		defaultFilter.Limit = 10
	}
	if defaultFilter.SortBy == "" {
		defaultFilter.SortBy = defaultSortBy
	}
	if defaultFilter.Asc == nil {
		v := true
		defaultFilter.Asc = &v
	}
	err = utils.StrictJsonTransformer(&defaultFilter, filter)
	if err != nil {
		return fmt.Errorf("app.SetDefault : %w", err)
	}
	return nil
}

func (h *HttpServer) WriteJsonWithStatusCode(ctx context.Context, w http.ResponseWriter, statusCode int, responseBody any) {
	var err error
	blob, ok := responseBody.([]byte)
	if !ok {
		blob, err = json.Marshal(responseBody)
		if err != nil {
			h.Log.Emergency(ctx, "Error in response json marshall", responseBody, fmt.Errorf("response marshal error: %w", err))
		}
	}
	w.Header().Set(HttpHeaderContentType, HttpContentTypeJSON)
	w.WriteHeader(statusCode)
	w.Write(blob)
}

func (h *HttpServer) WriteJson(ctx context.Context, w http.ResponseWriter, responseBody any) {
	h.WriteJsonWithStatusCode(ctx, w, http.StatusOK, responseBody)
}

func (b *HttpServer) GetCorrelationParams(r *http.Request) *log.CorrelationParam {
	correlationId := r.Header.Get("x-correlation-id")
	if correlationId == "" {
		return log.GetDefaultCorrelationParams(b.c.ServiceName)
	}
	return &log.CorrelationParam{
		CorrelationId: correlationId,
		ScenarioId:    r.Header.Get("x-scenario-id"),
		ScenarioName:  r.Header.Get("x-scenario-name"),
		SessionId:     r.Header.Get("x-session-id"),
	}
}

func (b *HttpServer) GetCustomerId(r *http.Request) *log.CustomerIdentifier {
	appUserId := r.Header.Get("x-app-user-id")
	if appUserId == "" {
		return &log.CustomerIdentifier{}
	}
	return &log.CustomerIdentifier{
		AppUserId:  appUserId,
		CustomerId: r.Header.Get("x-customer-id"),
		Id:         r.Header.Get("x-entity-id"),
	}
}

func (b *HttpServer) PrintRequest(ctx context.Context, r *http.Request) {
	h := r.Header
	popList := make(map[string][]string)
	for _, key := range b.c.Log.AuthHeaderKeyList {
		val := h.Values(key)
		if len(val) != 0 {
			popList[key] = val
			h.Set(key, "---redacted---")
		}
	}
	req := b.ExtractRequestMetadata(r)
	b.Log.Info(ctx, "Request", req)
	if b.c.Log.ContentLength >= r.ContentLength {
		body := b.CopyRequestBody(ctx, r)
		b.Log.Debug(ctx, "Request-Body", string(body))
	} else if b.c.Log.ContentLength < r.ContentLength {
		b.Log.Notice(ctx, "Request-Body", "Content length is too big to print check server log configuration")
	}
	for key, value := range popList {
		h.Del(key)
		for _, v := range value {
			h.Add(key, v)
		}
	}
}

func (b *HttpServer) CopyRequestBody(ctx context.Context, r *http.Request) []byte {
	if r.ContentLength <= 0 {
		return nil
	}
	body := r.Body
	defer body.Close()
	blobBody, _ := io.ReadAll(body)
	r.Body = io.NopCloser(bytes.NewReader(blobBody))
	contentType := r.Header.Get(HttpHeaderContentType)
	if strings.HasPrefix(contentType, HttpContentTypeJSON) {
		var prettyJSON bytes.Buffer
		err := json.Indent(&prettyJSON, blobBody, "", "\t")
		if err == nil {
			return prettyJSON.Bytes()
		}
	}
	return blobBody
}

func (b *HttpServer) ExtractRequestMetadata(r *http.Request) map[string]any {
	res := map[string]interface{}{
		"Method":        r.Method,
		"Header":        r.Header,
		"URL":           r.URL,
		"Proto":         r.Proto,
		"ContentLength": r.ContentLength,
		"Host":          r.Host,
		"RemoteAddr":    r.RemoteAddr,
		"RequestURI":    r.RequestURI,
	}
	return res
}

func (b *HttpServer) GetPort() string {
	return fmt.Sprintf("%v:%v", b.c.Host, b.c.Port)
}

func (b *HttpServer) SetContextError(ctx context.Context, err error) {
	iSetter := ctx.Value(ContextKeyError)
	if iSetter == nil {
		return
	}
	setter, ok := iSetter.(func(error))
	if !ok {
		panic(fmt.Errorf("context error handler corrupted, error to handle: %w", err))
	}
	setter(err)
}

func (b *HttpServer) AddServerHost(server DocumentServer) {
	b.docMeta.Server = append(b.docMeta.Server, server)
}
