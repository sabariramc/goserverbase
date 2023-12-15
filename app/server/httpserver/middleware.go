package httpserver

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func (h *HttpServer) SetContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		corr := h.GetCorrelationParams(r)
		id := h.GetCustomerId(r)
		ctx := h.GetContextWithCorrelation(r.Context(), corr)
		ctx = h.GetContextWithCustomerId(ctx, id)
		c.Request = r.WithContext(ctx)
		if span, ok := tracer.SpanFromContext(r.Context()); ok {
			span.SetTag("correlationId", corr.CorrelationId)
			defer span.Finish()
		}
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
		span, spanOk := tracer.SpanFromContext(r.Context())
		defer func() {
			if spanOk {
				span.SetTag(ext.HTTPCode, strconv.Itoa(loggingW.status))
			}
		}()
		ctx := r.Context()
		r = r.WithContext(ctx)
		h.PrintRequest(r)
		c.Writer = loggingW
		c.Request = r
		c.Next()
		if loggingW.status > 299 {
			h.log.Error(r.Context(), "Response", map[string]any{"statusCode": loggingW.status, "headers": loggingW.Header(), "body": loggingW.body})
		}
	}
}

func (h *HttpServer) HandleExceptionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w, r := c.Writer, c.Request
		req := h.ExtractRequestMetadata(r)
		req["Body"], _ = h.CopyRequestBody(r)
		span, spanOk := tracer.SpanFromContext(r.Context())
		defer func() {
			if rec := recover(); rec != nil {
				statusCode, body := h.PanicRecovery(r.Context(), rec, req)
				h.WriteJSONWithStatusCode(r.Context(), w, statusCode, body)
				if spanOk {
					err, errOk := rec.(error)
					if !errOk {
						err = fmt.Errorf("panic during execution")
					}
					span.SetTag(ext.Error, err)
				}
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
			if spanOk && statusCode > 299 {
				span.SetTag(ext.Error, handlerError)
			}

		}
	}
}
