package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sabariramc/goserverbase/v4/log"
)

func (h *HttpServer) GetCorrelationParams(r *http.Request) *log.CorrelationParam {
	correlationId := r.Header.Get("x-correlation-id")
	if correlationId == "" {
		return log.GetDefaultCorrelationParam(h.c.ServiceName)
	}
	return &log.CorrelationParam{
		CorrelationId: correlationId,
		ScenarioId:    r.Header.Get("x-scenario-id"),
		ScenarioName:  r.Header.Get("x-scenario-name"),
		SessionId:     r.Header.Get("x-session-id"),
	}
}

func (h *HttpServer) GetCustomerId(r *http.Request) *log.CustomerIdentifier {
	return &log.CustomerIdentifier{
		AppUserId:  r.Header.Get("x-app-user-id"),
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
	h.log.Info(ctx, "Request", req)
	if h.c.Log.ContentLength >= r.ContentLength {
		h.log.Debug(ctx, "Request-Body", func() string {
			return h.GetRequestBody(r)
		})
	} else if h.c.Log.ContentLength < r.ContentLength {
		h.log.Notice(ctx, "Request-Body", "Content length is too big to print check server log configuration")
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
	ctxBody := ctx.Value(ContextKeyRequestBodyString)
	if ctxBody != nil {
		body, ok := ctxBody.(*string)
		if !ok {
			h.log.Emergency(ctx, "Invalid type for ContextKeyRequestBodyString context variable", fmt.Errorf("HttpServer.GetRequestBody: invalid type for context body reference"), ctxBody)
		}
		if *body == "" {
			*body = string(h.GetBody(r))
		}
		return *body
	}
	val := string(h.CopyRequestBody(r))
	return val
}

func (h *HttpServer) GetBody(r *http.Request) []byte {
	ctx := r.Context()
	ctxBody := ctx.Value(ContextKeyRequestBodyRaw)
	if ctxBody != nil {
		body, ok := ctxBody.(**[]byte)
		if !ok {
			h.log.Emergency(ctx, "Invalid type for ContextKeyRequestBodyRaw context variable", fmt.Errorf("HttpServer.GetBody: invalid type for context body reference"), ctxBody)
		}
		if body == nil {
			return h.CopyRequestBody(r)
		}
		if *body == nil {
			data := h.CopyRequestBody(r)
			*body = &data
		}
		return **body
	}
	return h.CopyRequestBody(r)
}

func (h *HttpServer) GetJSONBody(r *http.Request, body any) error {
	blob := h.GetBody(r)
	if blob == nil {
		return fmt.Errorf("GetJSONBody.empty body")
	}
	return json.Unmarshal(blob, body)
}
