package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type loggingResponseWriter struct {
	status int
	body   string
	gin.ResponseWriter
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingResponseWriter) Write(body []byte) (int, error) {
	w.body = string(body)
	return w.ResponseWriter.Write(body)
}

func (h *HttpServer) WriteJSONWithStatusCode(ctx context.Context, w http.ResponseWriter, statusCode int, responseBody any) {
	var err error
	blob, ok := responseBody.([]byte)
	if !ok {
		blob, err = json.Marshal(responseBody)
		if err != nil {
			h.log.Emergency(ctx, "Error in response json marshall", fmt.Errorf("HttpServer.WriteJsonWithStatusCode: error marshalling response: %w", err), responseBody)
		}
	}
	w.Header().Set(HttpHeaderContentType, HttpContentTypeJSON)
	w.WriteHeader(statusCode)
	w.Write(blob)
}

func (h *HttpServer) WriteJson(ctx context.Context, w http.ResponseWriter, responseBody any) {
	h.WriteJSONWithStatusCode(ctx, w, http.StatusOK, responseBody)
}

func (h *HttpServer) WriteResponseWithStatusCode(ctx context.Context, w http.ResponseWriter, statusCode int, contentType string, responseBody []byte) {
	w.Header().Set(HttpHeaderContentType, contentType)
	w.WriteHeader(statusCode)
	w.Write(responseBody)
}

func (h *HttpServer) WriteResponse(ctx context.Context, w http.ResponseWriter, statusCode int, contentType string, responseBody []byte) {
	h.WriteResponseWithStatusCode(ctx, w, http.StatusOK, contentType, responseBody)
}
