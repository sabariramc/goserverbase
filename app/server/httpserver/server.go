package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	log     *log.Logger
	c       *HttpServerConfig
}

func New(appConfig HttpServerConfig, loggerConfig log.Config, lMux log.LogMux, errorNotifier errors.ErrorNotifier, auditLogger log.AuditLogWriter) *HttpServer {
	b := baseapp.New(*appConfig.ServerConfig, loggerConfig, lMux, errorNotifier, auditLogger)
	h := &HttpServer{
		BaseApp: b,
		handler: chi.NewRouter(),
		docMeta: APIDocumentation{
			Server: make([]DocumentServer, 0),
			Routes: make(APIRoute, 0),
		},
		log: b.GetLogger(),
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
	b.log.Notice(context.TODO(), fmt.Sprintf("Server starting at %v", b.GetPort()), nil)
	err := http.ListenAndServe(b.GetPort(), b)
	b.log.Emergency(context.Background(), "Server crashed", nil, err)
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

func WriteJsonWithStatusCode(w http.ResponseWriter, statusCode int, responseBody any) {
	var err error
	blob, ok := responseBody.([]byte)
	if !ok {
		blob, err = json.Marshal(responseBody)
		if err != nil {
			panic(fmt.Errorf("response marshal error: %w", err))
		}
	}
	w.Header().Add(HttpHeaderContentType, HttpContentTypeJSON)
	w.WriteHeader(statusCode)
	w.Write(blob)
}

func WriteJson(w http.ResponseWriter, responseBody any) {
	WriteJsonWithStatusCode(w, http.StatusOK, responseBody)
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
	for _, key := range b.c.AuthHeaderKeyList {
		val := h.Values(key)
		if len(val) != 0 {
			popList[key] = val
			h.Set(key, "---redacted---")
		}
	}
	b.log.Info(ctx, "Request", map[string]interface{}{
		"Method":        r.Method,
		"Header":        h,
		"URL":           r.URL,
		"Proto":         r.Proto,
		"ContentLength": r.ContentLength,
		"Host":          r.Host,
		"RemoteAddr":    r.RemoteAddr,
		"RequestURI":    r.RequestURI,
	})
	if r.ContentLength > 0 {
		body := r.Body
		defer body.Close()
		blobBody, _ := io.ReadAll(body)
		data := make(map[string]any)
		json.Unmarshal(blobBody, &data)
		r.Body = io.NopCloser(bytes.NewReader(blobBody))
		b.log.Debug(ctx, "Request Body", data)
	}
	for key, value := range popList {
		h.Del(key)
		for _, v := range value {
			h.Add(key, v)
		}
	}
}

func (b *HttpServer) GetPort() string {
	return fmt.Sprintf("%v:%v", b.c.Host, b.c.Port)
}

func (b *HttpServer) SetHandlerError(ctx context.Context, err error) {
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
