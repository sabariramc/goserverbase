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

func extractKeyValue(r *http.Request, keyList []string) map[string]string {
	res := make(map[string]string, len(keyList))
	for _, key := range keyList {
		value := r.Header.Get(key)
		if value != "" {
			res[key] = value
		}
	}
	return res
}

func (h *HttpServer) GetCorrelationParams(r *http.Request) *log.CorrelationParam {
	keyList := []string{"x-correlation-id", "x-scenario-id", "x-scenario-name", "x-session-id"}
	headers := extractKeyValue(r, keyList)
	cr := &log.CorrelationParam{}
	cr.LoadFromHeader(headers)
	if cr.CorrelationId == "" {
		return log.GetDefaultCorrelationParam(h.c.ServiceName)
	}
	return cr
}

func (h *HttpServer) GetCustomerId(r *http.Request) *log.CustomerIdentifier {
	keyList := []string{"x-app-user-id", "x-customer-id", "x-entity-id"}
	headers := extractKeyValue(r, keyList)
	id := &log.CustomerIdentifier{}
	id.LoadFromHeader(headers)
	return id
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

func (h *HttpServer) GetTextBody(r *http.Request) string {
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
	val := string(h.GetBody(r))
	return val
}
