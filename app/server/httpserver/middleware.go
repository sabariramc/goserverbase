package httpserver

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *HttpServer) SetContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		ctx := h.GetContextWithCorrelation(r.Context(), h.GetCorrelationParams(r))
		ctx = h.GetContextWithCustomerId(ctx, h.GetCustomerId(r))
		c.Request = r.WithContext(ctx)
		c.Next()
	}
}

func (h *HttpServer) RequestTimerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		st := time.Now()
		c.Next()
		h.Log.Info(r.Context(), "Request processing time in ms", time.Since(st).Milliseconds())
	}
}

func (h *HttpServer) LogRequestResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w, r := c.Writer, c.Request
		loggingW := &loggingResponseWriter{
			ResponseWriter: w,
		}
		var bodyBlob *[]byte
		var bodyStr string
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyRequestBodyRaw, &bodyBlob)
		ctx = context.WithValue(ctx, ContextKeyRequestBodyString, &bodyStr)
		r = r.WithContext(ctx)
		h.PrintRequest(r.Context(), r)
		c.Writer = loggingW
		c.Request = r
		c.Next()
		if loggingW.status < 500 {
			h.Log.Info(r.Context(), "Response", map[string]any{"statusCode": loggingW.status, "headers": loggingW.Header()})
			h.Log.Debug(r.Context(), "Response-Body", loggingW.body)
		} else {
			h.Log.Error(r.Context(), "Response", map[string]any{"statusCode": loggingW.status, "headers": loggingW.Header()})
			h.Log.Error(r.Context(), "Response-Body", loggingW.body)
		}

	}
}

func (h *HttpServer) HandleExceptionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w, r := c.Writer, c.Request
		req := h.ExtractRequestMetadata(r)
		req["Body"] = h.GetRequestBody(r)
		defer func() {
			if rec := recover(); rec != nil {
				statusCode, body := h.PanicRecovery(r.Context(), rec, req)
				h.WriteJsonWithStatusCode(r.Context(), w, statusCode, body)
			}
		}()
		ctx := r.Context()
		var handlerError error
		ctx = context.WithValue(ctx, ContextKeyHandlerError, func(err error) { handlerError = err })
		c.Request = r.WithContext(ctx)
		c.Next()
		if handlerError != nil {
			statusCode, body := h.ProcessError(ctx, "", handlerError, req)
			h.WriteJsonWithStatusCode(r.Context(), w, statusCode, body)
		}
	}
}
