package httpserver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
)

// SetCorrelationMiddleware returns a middleware that sets the correlation parameters and user identifier in the request context.
func (h *HTTPServer) SetCorrelationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		corr := h.GetCorrelationParams(r)
		identity := h.GetCustomerID(r)
		ctx := correlation.GetContextWithCorrelationParam(r.Context(), corr)
		ctx = correlation.GetContextWithUserIdentifier(ctx, identity)
		c.Request = r.WithContext(ctx)
		if h.tracer != nil {
			span, ok := h.tracer.GetSpanFromContext(ctx)
			if ok {
				span.SetAttribute("correlationId", corr.CorrelationID)
				defer span.Finish()
				data := identity.GetPayload()
				for key, value := range data {
					if value != "" {
						span.SetAttribute("user."+key, value)
					}
				}
			}
		}
		c.Next()
	}
}

// RequestTimerMiddleware returns a middleware that logs the request processing time.
func (h *HTTPServer) RequestTimerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		st := time.Now()
		c.Next()
		h.log.Info(r.Context(), "Request processing time in ms", time.Since(st).Milliseconds())
	}
}

// LogRequestResponseMiddleware returns a middleware that logs the request and response.
func (h *HTTPServer) LogRequestResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w, r := c.Writer, c.Request
		ctx := r.Context()
		r = r.WithContext(ctx)
		req := h.GetMaskedRequestMeta(r)
		c.Request = r
		body, _ := h.CopyRequestBody(r)
		logResWri := &loggingResponseWriter{
			ResponseWriter: w,
			ctx:            r.Context(),
			log:            h.log,
			reqBody:        &body,
		}
		c.Writer = logResWri
		h.log.Info(ctx, "Request", req)
		h.log.Debug(ctx, "Request Body", func() string { return string(body) })
		cs, spanOk := h.GetSpanFromContext(ctx)
		defer func() {
			if spanOk {
				cs.SetAttribute(span.HTTPStatusCode, strconv.Itoa(logResWri.status))
			}
		}()
		c.Next()
	}
}

// PanicHandleMiddleware returns a middleware that recovers from panics and logs the error and stack trace.
func (h *HTTPServer) PanicHandleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w, r := c.Writer, c.Request
		ctx := r.Context()
		span, spanOk := h.GetSpanFromContext(ctx)
		defer func() {
			if rec := recover(); rec != nil {
				stackTrace, err := h.PanicRecovery(ctx, rec)
				statusCode, body := h.ProcessError(ctx, stackTrace, err)
				h.WriteJSONWithStatusCode(r.Context(), w, statusCode, body)
				if spanOk {
					err, errOk := rec.(error)
					if !errOk {
						err = fmt.Errorf("panic during execution")
					}
					span.SetError(err, stackTrace)
				}
			}
		}()
		c.Next()
	}
}
