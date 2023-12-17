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

func (h *HTTPServer) SetContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		corr := h.GetCorrelationParams(r)
		ctx := h.GetContextWithCorrelation(r.Context(), corr)
		ctx = h.GetContextWithCustomerId(ctx, h.GetCustomerID(r))
		c.Request = r.WithContext(ctx)
		if span, ok := tracer.SpanFromContext(r.Context()); ok {
			span.SetTag("correlationId", corr.CorrelationId)
			defer span.Finish()
		}
		c.Next()
	}
}

func (h *HTTPServer) RequestTimerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		st := time.Now()
		c.Next()
		h.log.Info(r.Context(), "Request processing time in ms", time.Since(st).Milliseconds())
	}
}

func (h *HTTPServer) LogRequestResponseMiddleware() gin.HandlerFunc {
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
		var bodyBlob *[]byte
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyRequestBody, &bodyBlob)
		r = r.WithContext(ctx)
		req := h.GetMaskedRequestMeta(r)
		c.Writer = loggingW
		c.Request = r
		body, _ := h.CopyRequestBody(r)
		bodyBlob = &body
		h.log.Info(ctx, "RequestMeta", req)
		c.Next()
		res := map[string]any{"statusCode": loggingW.status, "headers": loggingW.Header()}
		if loggingW.status > 299 {
			req["Body"] = string(body)
			res["Body"] = loggingW.body
			h.log.Error(ctx, "Request", req)
			h.log.Error(ctx, "Response", res)
		} else {
			h.log.Info(ctx, "ResponseMeta", res)
		}
	}
}

func (h *HTTPServer) HandleExceptionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w, r := c.Writer, c.Request
		span, spanOk := tracer.SpanFromContext(r.Context())
		defer func() {
			if rec := recover(); rec != nil {
				statusCode, body := h.PanicRecovery(r.Context(), rec)
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
			statusCode, body := h.ProcessError(ctx, stackTrace, handlerError)
			h.WriteJSONWithStatusCode(r.Context(), w, statusCode, body)
			if spanOk && statusCode > 299 {
				span.SetTag(ext.Error, handlerError)
			}

		}
	}
}
