package baseapp

import (
	"context"
	"encoding/json"
	e "errors"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/sabariramc/goserverbase/errors"
	"github.com/sabariramc/goserverbase/log"
)

func (b *BaseApp) RequestTimerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()
		next.ServeHTTP(w, r)
		b.log.Info(r.Context(), "Request processing time in ms", time.Since(st).Milliseconds())
	})

}

func (b *BaseApp) SetContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := b.GetCorrelationContext(r.Context(), b.GetHttpCorrelationParams(r))
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

func (b *BaseApp) LogRequestResponseMiddleware(next http.Handler) http.Handler {
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

type ErrorRecorder func(err error)

func (b *BaseApp) HandleExceptionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		ctx := r.Context()
		var body, errorData, customerIdentifier interface{}
		var errorCode string
		defer func() {
			if rec := recover(); rec != nil {
				stackTrace := string(debug.Stack())
				b.log.Error(ctx, "Recovered in Responder - Error", rec)
				b.log.Error(ctx, "Recovered in Responder - StackTrace", stackTrace)
				statusCode = http.StatusInternalServerError
				customerIdentifier = log.GetCustomerIdentifier(ctx)
				err := rec.(error)
				body = map[string]string{"error": "Internal error occurred, if persist contact technical team"}
				notify := false
				var customError *errors.CustomError
				var httpErr *errors.HTTPError
				if e.As(err, &httpErr) {
					statusCode = httpErr.ErrorStatusCode
					notify = httpErr.Notify
					body = httpErr.GetErrorResponse()
					errorCode = httpErr.ErrorCode
					errorData = httpErr.ErrorData

				} else if e.As(err, &customError) {
					statusCode = http.StatusInternalServerError
					notify = customError.Notify
					body = customError.GetErrorResponse()
					errorData = customError.ErrorData
				} else {
					statusCode = http.StatusInternalServerError
					customError = errors.NewCustomError("UNKNOWN", "Unknown error", err, nil, true)
					body = customError.GetErrorResponse()
					err = customError
				}
				if notify && b.errorNotifier != nil {
					b.errorNotifier.Send(ctx, errorCode, err, stackTrace, errorData, customerIdentifier)
				}
				w.Header().Set(HttpHeaderContentType, HttpContentTypeJSON)
				w.WriteHeader(statusCode)
				res, _ := json.Marshal(body)
				w.Write([]byte(res))
			}
		}()
		var handlerError error
		ctx = context.WithValue(ctx, ContextKeyError, func(err error) { handlerError = err })
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
		if handlerError != nil {
			panic(handlerError)
		}
	})
}
