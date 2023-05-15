package httpserver

import (
	"context"
	"net/http"
	"time"
)

func (b *HttpServer) RequestTimerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()
		next.ServeHTTP(w, r)
		b.Log.Info(r.Context(), "Request processing time in ms", time.Since(st).Milliseconds())
	})

}

func (b *HttpServer) SetContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := b.GetContextWithCorrelation(r.Context(), b.GetCorrelationParams(r))
		ctx = b.GetContextWithCustomerId(ctx, b.GetCustomerId(r))
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

type loggingResponseWriter struct {
	status int
	body   string
	http.ResponseWriter
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingResponseWriter) Write(body []byte) (int, error) {
	w.body = string(body)
	return w.ResponseWriter.Write(body)
}

func (b *HttpServer) LogRequestResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.PrintRequest(r.Context(), r)
		loggingRW := &loggingResponseWriter{
			ResponseWriter: w,
		}
		next.ServeHTTP(loggingRW, r)
		if loggingRW.status < 500 {
			b.Log.Info(r.Context(), "Response", map[string]any{"statusCode": loggingRW.status, "headers": loggingRW.Header()})
			b.Log.Debug(r.Context(), "Response-Body", loggingRW.body)
		} else {
			b.Log.Error(r.Context(), "Response", map[string]any{"statusCode": loggingRW.status, "headers": loggingRW.Header()})
			b.Log.Error(r.Context(), "Response-Body", loggingRW.body)
		}

	})
}

func (b *HttpServer) HandleExceptionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				statusCode, body := b.PanicRecovery(r.Context(), rec, nil)
				b.WriteJsonWithStatusCode(r.Context(), w, statusCode, body)
			}
		}()
		body := b.CopyRequestBody(r.Context(), r)
		ctx := r.Context()
		var handlerError error
		ctx = context.WithValue(ctx, ContextKeyError, func(err error) { handlerError = err })
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
		if handlerError != nil {
			reqMeta := b.ExtractRequestMetadata(r)
			reqMeta["Body"] = string(body)
			statusCode, body := b.ProcessError(ctx, "", handlerError, reqMeta)
			b.WriteJsonWithStatusCode(r.Context(), w, statusCode, body)
		}
	})
}
