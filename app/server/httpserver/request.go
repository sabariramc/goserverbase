package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sabariramc/goserverbase/v6/correlation"
)

// extractKeyValue extracts the values of specified keys from the request headers.
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

// GetCorrelationParams extracts correlation parameters from the request headers.
// If the correlation ID is missing, it generates a new one using the service name.
func (h *HTTPServer) GetCorrelationParams(r *http.Request) *correlation.CorrelationParam {
	keyList := []string{"x-correlation-id", "x-scenario-id", "x-scenario-name", "x-session-id"}
	headers := extractKeyValue(r, keyList)
	cr := &correlation.CorrelationParam{}
	cr.LoadFromHeader(headers)
	if cr.CorrelationID == "" {
		return correlation.NewCorrelationParam(h.c.ServiceName)
	}
	return cr
}

// GetCustomerID extracts user identifiers from the request headers.
func (h *HTTPServer) GetCustomerID(r *http.Request) *correlation.UserIdentifier {
	keyList := []string{"x-user-id", "x-user-client-id", "x-entity-id"}
	headers := extractKeyValue(r, keyList)
	id := &correlation.UserIdentifier{}
	id.LoadFromHeader(headers)
	return id
}

// GetMaskedRequestMeta extracts request metadata and masks sensitive headers specified in the configuration.
func (h *HTTPServer) GetMaskedRequestMeta(r *http.Request) map[string]any {
	header := r.Header
	popList := make(map[string][]string)
	for _, key := range h.c.Mask.HeaderKeyList {
		val := header.Values(key)
		if len(val) != 0 {
			popList[key] = val
			header.Set(key, "---redacted---")
		}
	}
	req := h.ExtractRequestMetadata(r)
	for key, value := range popList {
		header.Del(key)
		for _, v := range value {
			header.Add(key, v)
		}
	}
	return req
}

// CopyRequestBody copies the request body and returns it as a byte slice.
func (h *HTTPServer) CopyRequestBody(r *http.Request) ([]byte, error) {
	blobBody, err := h.GetRequestBody(r)
	if err != nil {
		return blobBody, fmt.Errorf("HttpServer.CopyRequestBody: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(blobBody))
	return blobBody, nil
}

// ExtractRequestMetadata extracts metadata from the HTTP request.
func (h *HTTPServer) ExtractRequestMetadata(r *http.Request) map[string]any {
	res := map[string]any{
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

// GetRequestBody reads and returns the request body as a byte slice.
func (h *HTTPServer) GetRequestBody(r *http.Request) ([]byte, error) {
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

// LoadRequestJSONBody reads the request body and unmarshals it into the provided interface.
func (h *HTTPServer) LoadRequestJSONBody(r *http.Request, body any) error {
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
