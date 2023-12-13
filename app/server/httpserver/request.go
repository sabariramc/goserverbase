package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

func (h *HttpServer) PrintRequest(r *http.Request) {
	ctx := r.Context()
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

func (h *HttpServer) CopyRequestBody(r *http.Request) ([]byte, error) {
	blobBody, err := h.GetRequestBody(r)
	if err != nil {
		return blobBody, fmt.Errorf("HttpServer.CopyRequestBody: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(blobBody))
	return blobBody, nil
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

func (h *HttpServer) GetRequestBody(r *http.Request) ([]byte, error) {
	if r.ContentLength <= 0 {
		return nil, nil
	}
	body := r.Body
	defer body.Close()
	blobBody, err := io.ReadAll(body)
	if err != nil {
		err = fmt.Errorf("HttpServer.GetRequestBody: error reading request body: %w", err)
	}
	return blobBody, err
}

func (h *HttpServer) LoadRequestJSONBody(r *http.Request, body any) error {
	blobBody, err := h.GetRequestBody(r)
	if err != nil {
		return fmt.Errorf("HttpServer.LoadJSONBody: %w", err)
	}
	err = json.Unmarshal(blobBody, body)
	if err != nil {
		err = fmt.Errorf("HttpServer.LoadJSONBody: error loading request body to object: %w", err)
	}
	return err
}
