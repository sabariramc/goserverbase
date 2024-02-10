package httpserver

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *HTTPServer) SetContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		corr := h.GetCorrelationParams(r)
		identity := h.GetCustomerID(r)
		ctx := h.GetContextWithCorrelation(r.Context(), corr)
		ctx = h.GetContextWithCustomerId(ctx, identity)
		c.Request = r.WithContext(ctx)
		if h.tracer != nil {
			span, ok := h.tracer.GetSpanFromContext(ctx)
			if ok {
				data := identity.GetPayload()
				for key, value := range data {
					if value != "" {
						span.SetTag("customer."+key, value)
					}
				}
			}
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
		logResWri := &loggingResponseWriter{
			ResponseWriter: w,
		}
		var bodyBlob *[]byte
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyRequestBody, &bodyBlob)
		r = r.WithContext(ctx)
		req := h.GetMaskedRequestMeta(r)
		c.Writer = logResWri
		c.Request = r
		body, _ := h.CopyRequestBody(r)
		bodyBlob = &body
		h.log.Info(ctx, "RequestMeta", req)
		c.Next()
		res := map[string]any{"statusCode": logResWri.status, "headers": logResWri.Header()}
		if logResWri.status > 299 {
			req["Body"] = string(body)
			res["Body"] = logResWri.body
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
		defer func() {
			if rec := recover(); rec != nil {
				stackTrace, err := h.PanicRecovery(r.Context(), rec)
				statusCode, body := h.ProcessError(r.Context(), stackTrace, err)
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
