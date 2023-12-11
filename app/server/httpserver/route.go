package httpserver

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sabariramc/goserverbase/v4/errors"
	gintrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
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
		err := errors.NewHTTPClientError(http.StatusNotFound, "URL_NOT_FOUND", "Invalid path", nil, map[string]string{
			"path": r.URL.Path,
		}, nil)
		body, _ := err.GetErrorResponse()
		w.Write(body)
		if span, spanOk := tracer.SpanFromContext(r.Context()); spanOk {
			span.SetTag(ext.HTTPCode, strconv.Itoa(404))
			span.SetTag(ext.Error, err)
		}
	}
}

func MethodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(HttpHeaderContentType, HttpContentTypeJSON)
		w.WriteHeader(http.StatusMethodNotAllowed)
		err := errors.NewHTTPClientError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Invalid method", nil, map[string]string{
			"path":   r.URL.Path,
			"method": r.Method,
		}, nil)
		body, _ := err.GetErrorResponse()
		w.Write(body)
		if span, spanOk := tracer.SpanFromContext(r.Context()); spanOk {
			span.SetTag(ext.HTTPCode, strconv.Itoa(405))
			span.SetTag(ext.Error, err)
		}
	}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(204)
}

func (h *HttpServer) SetupRouter(ctx context.Context) {
	h.handler.NoRoute(gin.WrapF(NotFound()))
	h.handler.NoMethod(gin.WrapF(MethodNotAllowed()))
	h.handler.GET("/meta/health", gin.WrapF(HealthCheck))
	h.handler.HandleMethodNotAllowed = true
	h.handler.Use(gintrace.Middleware(h.c.ServiceName))
	h.handler.Use(h.SetContextMiddleware(), h.RequestTimerMiddleware(), h.LogRequestResponseMiddleware(), h.HandleExceptionMiddleware())
}
