package httpserver

import (
	"context"
	"net/http"
	"time"
)

func (h *HttpServer) SetContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := h.GetContextWithCorrelation(r.Context(), h.GetCorrelationParams(r))
		ctx = h.GetContextWithCustomerId(ctx, h.GetCustomerId(r))
		r = r.WithContext(ctx)
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
		r = h.SetRequestBodyInContext(r)
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
		defer func() {
			if rec := recover(); rec != nil {
				statusCode, body := h.PanicRecovery(r.Context(), rec, req)
				h.WriteJsonWithStatusCode(r.Context(), w, statusCode, body)
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
		}
	})
}
