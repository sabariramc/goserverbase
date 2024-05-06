package httpserver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sabariramc/goserverbase/v5/instrumentation/span"
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
		cs, spanOk := h.GetSpanFromContext(ctx)
		defer func() {
			if spanOk {
				cs.SetAttribute(span.HTTPStatusCode, strconv.Itoa(logResWri.status))
			}
		}()
		c.Next()
	}
}

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
