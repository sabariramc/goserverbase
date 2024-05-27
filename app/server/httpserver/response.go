// Package httpserver provides utilities for managing an HTTP server, including logging and response handling.
package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sabariramc/goserverbase/v6/log"
)

// loggingResponseWriter is a custom response writer that logs responses and request bodies.
type loggingResponseWriter struct {
	status             int             // HTTP status code
	log                log.Log         // Logger instance
	reqBody            *[]byte         // Pointer to the request body
	ctx                context.Context // Context for logging
	gin.ResponseWriter                 // Embedded Gin response writer
}

// WriteHeader logs the status code and writes the header to the response.
func (w *loggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Write logs the response body and writes it to the response.
func (w *loggingResponseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	ctx := w.ctx
	res := map[string]any{"statusCode": w.status, "headers": w.Header()}
	if w.status <= 299 {
		w.log.Info(ctx, "Response", res)
	} else {
		res["body"] = string(body)
		w.log.Error(ctx, "Request Body", string(*w.reqBody))
		w.log.Error(ctx, "Response", res)
	}
	return w.ResponseWriter.Write(body)
}

// WriteJSONWithStatusCode writes a JSON response with the specified status code.
func (h *HTTPServer) WriteJSONWithStatusCode(ctx context.Context, w http.ResponseWriter, statusCode int, responseBody any) {
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

// WriteJSON writes a JSON response with a status code of 200 OK.
func (h *HTTPServer) WriteJSON(ctx context.Context, w http.ResponseWriter, responseBody any) {
	h.WriteJSONWithStatusCode(ctx, w, http.StatusOK, responseBody)
}

// WriteResponseWithStatusCode writes a response with the specified status code and content type.
func (h *HTTPServer) WriteResponseWithStatusCode(ctx context.Context, w http.ResponseWriter, statusCode int, contentType string, responseBody []byte) {
	w.Header().Set(HttpHeaderContentType, contentType)
	w.WriteHeader(statusCode)
	w.Write(responseBody)
}

// WriteResponse writes a response with a status code of 200 OK and the specified content type.
func (h *HTTPServer) WriteResponse(ctx context.Context, w http.ResponseWriter, contentType string, responseBody []byte) {
	h.WriteResponseWithStatusCode(ctx, w, http.StatusOK, contentType, responseBody)
}

// WriteErrorResponse writes an error response, logging the error and stack trace.
func (h *HTTPServer) WriteErrorResponse(ctx context.Context, w http.ResponseWriter, err error, stackTrace string) {
	statusCode, body := h.ProcessError(ctx, stackTrace, err)
	span, spanOk := h.GetSpanFromContext(ctx)
	if spanOk {
		span.SetError(err, stackTrace)
	}
	h.WriteJSONWithStatusCode(ctx, w, statusCode, body)
}
