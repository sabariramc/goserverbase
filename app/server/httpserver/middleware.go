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
		h.log.Info(r.Context(), "Request processing time in ms", time.Since(st).Milliseconds())
	}
}

func (h *HttpServer) LogRequestResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w, r := c.Writer, c.Request
		loggingW := &loggingResponseWriter{
			ResponseWriter: w,
		}
		ctx := r.Context()
		r = r.WithContext(ctx)
		h.PrintRequest(r)
		c.Writer = loggingW
		c.Request = r
		c.Next()
		if loggingW.status > 299 {
			h.log.Error(r.Context(), "Response", map[string]any{"statusCode": loggingW.status, "headers": loggingW.Header()})
			h.log.Error(r.Context(), "Response-Body", loggingW.body)
		}

	}
}

func (h *HttpServer) HandleExceptionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w, r := c.Writer, c.Request
		req := h.ExtractRequestMetadata(r)
		req["Body"], _ = h.CopyRequestBody(r)
		defer func() {
			if rec := recover(); rec != nil {
				statusCode, body := h.PanicRecovery(r.Context(), rec, req)
				h.WriteJSONWithStatusCode(r.Context(), w, statusCode, body)
			}
		}()
		ctx := r.Context()
		var handlerError error
		var stackTrace string
		ctx = context.WithValue(ctx, ContextKeyHandlerError, func(err error) { handlerError = err })
		ctx = context.WithValue(ctx, ContextKeyHandlerErrorStackTrace, func(st string) { stackTrace = st })
		c.Request = r.WithContext(ctx)
		c.Next()
		if handlerError != nil {
			if blob, ok := req["Body"].([]byte); ok {
				req["Body"] = string(blob)
			}
			statusCode, body := h.ProcessError(ctx, stackTrace, handlerError, req)
			h.WriteJSONWithStatusCode(r.Context(), w, statusCode, body)
		}
	}
}
