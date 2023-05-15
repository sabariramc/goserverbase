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
