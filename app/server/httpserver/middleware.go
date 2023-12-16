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
		req := h.GetMaskedRequestMeta(r)
		c.Writer = loggingW
		c.Request = r
		c.Next()
		if loggingW.status > 299 {
			h.log.Error(ctx, "Request", req)
			h.log.Error(ctx, "Response", map[string]any{"statusCode": loggingW.status, "headers": loggingW.Header(), "body": loggingW.body})
		} else {
			h.log.Info(ctx, "Request", req)
		}
	}
}

func (h *HttpServer) HandleExceptionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w, r := c.Writer, c.Request
		defer func() {
			if rec := recover(); rec != nil {
				statusCode, body := h.PanicRecovery(r.Context(), rec)
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
			statusCode, body := h.ProcessError(ctx, stackTrace, handlerError)
			h.WriteJSONWithStatusCode(r.Context(), w, statusCode, body)
		}
	}
}
