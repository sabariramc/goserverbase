package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/sabariramc/goserverbase/v2/log"
)

func (h *HttpServer) GetCorrelationParams(r *http.Request) *log.CorrelationParam {
	correlationId := r.Header.Get("x-correlation-id")
	if correlationId == "" {
		return log.GetDefaultCorrelationParams(h.c.ServiceName)
	}
	return &log.CorrelationParam{
		CorrelationId: correlationId,
		ScenarioId:    r.Header.Get("x-scenario-id"),
		ScenarioName:  r.Header.Get("x-scenario-name"),
		SessionId:     r.Header.Get("x-session-id"),
	}
}

func (h *HttpServer) GetCustomerId(r *http.Request) *log.CustomerIdentifier {
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

func (h *HttpServer) PrintRequest(ctx context.Context, r *http.Request) {
	header := r.Header
	popList := make(map[string][]string)
	for _, key := range h.c.Log.AuthHeaderKeyList {
		val := header.Values(key)
		if len(val) != 0 {
			popList[key] = val
			header.Set(key, "---redacted---")
		}
	}
	req := h.ExtractRequestMetadata(r)
	h.Log.Info(ctx, "Request", req)
	if h.c.Log.ContentLength >= r.ContentLength {
		h.Log.Debug(ctx, "Request-Body", func() string {
			return h.GetRequestBody(r)
		})
	} else if h.c.Log.ContentLength < r.ContentLength {
		h.Log.Notice(ctx, "Request-Body", "Content length is too big to print check server log configuration")
	}
	for key, value := range popList {
		header.Del(key)
		for _, v := range value {
			header.Add(key, v)
		}
	}
}

func (h *HttpServer) CopyRequestBody(r *http.Request) []byte {
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

func (h *HttpServer) ExtractRequestMetadata(r *http.Request) map[string]any {
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

func (h *HttpServer) GetRequestBody(r *http.Request) string {
	ctx := r.Context()
	ctxBody := ctx.Value(ContextKeyRequestBody)
	if ctxBody != nil {
		body, ok := ctxBody.(*string)
		if ok {
			if *body == "" {
				val := string(h.CopyRequestBody(r))
				*body = val
			}
			return *body
		}
	}
	val := string(h.CopyRequestBody(r))
	return val
}
