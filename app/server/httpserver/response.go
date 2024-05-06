package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/sabariramc/goserverbase/v5/log"
)

type loggingResponseWriter struct {
	status  int
	log     log.Log
	reqBody *[]byte
	ctx     context.Context
	gin.ResponseWriter
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

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

func (h *HTTPServer) WriteJSON(ctx context.Context, w http.ResponseWriter, responseBody any) {
	h.WriteJSONWithStatusCode(ctx, w, http.StatusOK, responseBody)
}

func (h *HTTPServer) WriteResponseWithStatusCode(ctx context.Context, w http.ResponseWriter, statusCode int, contentType string, responseBody []byte) {
	w.Header().Set(HttpHeaderContentType, contentType)
	w.WriteHeader(statusCode)
	w.Write(responseBody)
}

func (h *HTTPServer) WriteResponse(ctx context.Context, w http.ResponseWriter, contentType string, responseBody []byte) {
	h.WriteResponseWithStatusCode(ctx, w, http.StatusOK, contentType, responseBody)
}

func (h *HTTPServer) WriteErrorResponse(ctx context.Context, w http.ResponseWriter, err error, includeStackTrace bool) {
	stackTrace := ""
	if includeStackTrace {
		stack := []byte{}
		runtime.Stack(stack, false)
		stackTrace = string(stack)
	}
	statusCode, body := h.ProcessError(ctx, stackTrace, err)
	span, spanOk := h.GetSpanFromContext(ctx)
	if spanOk {
		span.SetError(err, stackTrace)
	}
	h.WriteJSONWithStatusCode(ctx, w, statusCode, body)
}
