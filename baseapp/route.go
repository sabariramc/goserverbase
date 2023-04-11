package baseapp

import (
	"context"
	"net/http"

	"github.com/sabariramc/goserverbase/v2/errors"
	chitrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go-chi/chi.v5"
)

type APIDocumentation struct {
	Server []DocumentServer
	Routes APIRoute
}

type DocumentServer struct {
	Tag     string
	BaseURL string
}

type APIRoute map[string]map[string]*APIHandler

type Response struct {
	StatusCode        int
	StatusDescription string
	Response          interface{}
}

type APIHandler struct {
	Func            http.HandlerFunc `json:"-"`
	Description     string
	Tags            []string
	Payload         interface{}
	SuccessResponse []Response
	FailureResponse []Response
}

func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(HttpHeaderContentType, HttpContentTypeJSON)
		w.WriteHeader(http.StatusNotFound)
		body := errors.NewHTTPClientError(http.StatusNotFound, "URL_NOT_FOUND", "Invalid path", nil, map[string]string{
			"path": r.URL.Path,
		}).Error()
		w.Write([]byte(body))
	}
}

func MethodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(HttpHeaderContentType, HttpContentTypeJSON)
		w.WriteHeader(http.StatusMethodNotAllowed)
		body := errors.NewHTTPClientError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Invalid method", nil, map[string]string{
			"path":   r.URL.Path,
			"method": r.Method,
		}).Error()
		w.Write([]byte(body))
	}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(204)
}

func (b *BaseApp) SetupRouter(ctx context.Context) {
	b.handler.Use(chitrace.Middleware())
	b.handler.Use(b.SetContextMiddleware, b.RequestTimerMiddleware, b.LogRequestResponseMiddleware, b.HandleExceptionMiddleware)
	b.handler.NotFound(NotFound())
	b.handler.MethodNotAllowed(MethodNotAllowed())
	b.handler.Get("/meta/health", HealthCheck)
}
