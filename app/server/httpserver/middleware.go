package httpserver

import (
	"context"
	"fmt"
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
			span.SetTag("http.headers.x-correlation-id", corr.CorrelationId)
			span.SetTag("http.headers.x-app-user-id", id.AppUserId)
			span.SetTag("http.headers.x-customer-id", id.CustomerId)
			span.SetTag("http.headers.x-entity-id", id.Id)
		}
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
		span, spanOk := tracer.SpanFromContext(r.Context())
		defer func() {
			if rec := recover(); rec != nil {
				statusCode, body := h.PanicRecovery(r.Context(), rec, req)
				h.WriteJsonWithStatusCode(r.Context(), w, statusCode, body)
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
		ctx = context.WithValue(ctx, ContextKeyHandlerError, func(err error) { handlerError = err })
		c.Request = r.WithContext(ctx)
		c.Next()
		if handlerError != nil {
			statusCode, body := h.ProcessError(ctx, "", handlerError, req)
			h.WriteJsonWithStatusCode(r.Context(), w, statusCode, body)
			if statusCode >= 500 {
				span.SetTag(ext.Error, handlerError)
			}
		}
	}
}
