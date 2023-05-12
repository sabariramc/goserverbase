package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

func (b *HttpServer) RequestTimerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()
		next.ServeHTTP(w, r)
		b.log.Info(r.Context(), "Request processing time in ms", time.Since(st).Milliseconds())
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
			b.log.Info(r.Context(), "Response", map[string]any{"statusCode": loggingRW.status, "headers": loggingRW.Header()})
			b.log.Debug(r.Context(), "Response-Body", loggingRW.body)
		} else {
			b.log.Error(r.Context(), "Response", map[string]any{"statusCode": loggingRW.status, "headers": loggingRW.Header()})
			b.log.Error(r.Context(), "Response-Body", loggingRW.body)
		}

	})
}

func (b *HttpServer) SendErrorResponse(ctx context.Context, w http.ResponseWriter, stackTrace string, err error) {
	statusCode, body := b.ProcessError(ctx, stackTrace, err)
	WriteJsonWithStatusCode(w, statusCode, body)
}

func (b *HttpServer) HandleExceptionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rec := recover(); rec != nil {
				stackTrace := string(debug.Stack())
				b.log.Error(ctx, "Recovered - Panic", rec)
				b.log.Error(ctx, "Recovered - StackTrace", stackTrace)
				err, ok := rec.(error)
				if !ok {
					blob, _ := json.Marshal(rec)
					err = fmt.Errorf("non error panic: %v", string(blob))
				}
				b.SendErrorResponse(ctx, w, stackTrace, err)
			}
		}()
		var handlerError error
		ctx = context.WithValue(ctx, ContextKeyError, func(err error) { handlerError = err })
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
		if handlerError != nil {
			b.SendErrorResponse(ctx, w, "", handlerError)
		}
	})
}
