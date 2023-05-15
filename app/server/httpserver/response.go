package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

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

func (h *HttpServer) WriteResponseWithStatusCode(ctx context.Context, w http.ResponseWriter, statusCode int, contentType string, responseBody any) {
	var blob []byte
	var ok bool
	var err error
	if blob, ok = responseBody.([]byte); !ok {
		blob, err = h.GetBytes(responseBody)
		if err != nil {
			h.Log.Emergency(ctx, "response encoding error", responseBody, err)
		}
	}
	w.Header().Set(HttpHeaderContentType, contentType)
	w.WriteHeader(statusCode)
	w.Write(blob)
}

func (h *HttpServer) WriteResponse(ctx context.Context, w http.ResponseWriter, statusCode int, contentType string, responseBody any) {
	h.WriteResponseWithStatusCode(ctx, w, http.StatusOK, contentType, responseBody)
}
