package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func (h *HttpServer) SetContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corr := h.GetCorrelationParams(r)
		id := h.GetCustomerId(r)
		ctx := h.GetContextWithCorrelation(r.Context(), corr)
		ctx = h.GetContextWithCustomerId(ctx, id)
		r = r.WithContext(ctx)
		if span, ok := tracer.SpanFromContext(r.Context()); ok {
			span.SetTag("http.headers.x-correlation-id", corr.CorrelationId)
			span.SetTag("http.headers.x-app-user-id", id.AppUserId)
			span.SetTag("http.headers.x-customer-id", id.CustomerId)
			span.SetTag("http.headers.x-entity-id", id.Id)
		}
		next.ServeHTTP(w, r)
	})
}

func (h *HttpServer) RequestTimerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()
		next.ServeHTTP(w, r)
		h.Log.Info(r.Context(), "Request processing time in ms", time.Since(st).Milliseconds())
	})

}

func (h *HttpServer) LogRequestResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggingW := &loggingResponseWriter{
			ResponseWriter: w,
		}
		var body string
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyRequestBody, &body)
		r = r.WithContext(ctx)
		h.PrintRequest(r.Context(), r)
		next.ServeHTTP(loggingW, r)
		if loggingW.status < 500 {
			h.Log.Info(r.Context(), "Response", map[string]any{"statusCode": loggingW.status, "headers": loggingW.Header()})
			h.Log.Debug(r.Context(), "Response-Body", loggingW.body)
		} else {
			h.Log.Error(r.Context(), "Response", map[string]any{"statusCode": loggingW.status, "headers": loggingW.Header()})
			h.Log.Error(r.Context(), "Response-Body", loggingW.body)
		}

	})
}

func (h *HttpServer) HandleExceptionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := h.ExtractRequestMetadata(r)
		req["Body"] = h.GetRequestBody(r)
		span, spanOk := tracer.SpanFromContext(r.Context())
		defer func() {
			if rec := recover(); rec != nil {
				statusCode, body := h.PanicRecovery(r.Context(), rec, req)
				h.WriteJsonWithStatusCode(r.Context(), w, statusCode, body)
				if spanOk {
					err, errOk := rec.(error)
					span.Finish(func(cfg *ddtrace.FinishConfig) {
						if errOk {
							cfg.Error = err
						} else {
							cfg.Error = fmt.Errorf("panic during execution")
						}
						cfg.NoDebugStack = false
						cfg.StackFrames = 15
						cfg.SkipStackFrames = 1
					})
				}
			}
		}()
		ctx := r.Context()
		var handlerError error
		ctx = context.WithValue(ctx, ContextKeyHandlerError, func(err error) { handlerError = err })
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
		if handlerError != nil {
			statusCode, body := h.ProcessError(ctx, "", handlerError, req)
			h.WriteJsonWithStatusCode(r.Context(), w, statusCode, body)
			if statusCode >= 500 {
				span.SetTag(ext.Error, handlerError)
			}
		}
	})
}
