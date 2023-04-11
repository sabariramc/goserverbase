package baseapp

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sabariramc/goserverbase/errors"
	"github.com/sabariramc/goserverbase/log"
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

func (b *BaseApp) RegisterRoutes(ctx context.Context, method, path string, handler http.HandlerFunc) {
	b.RegisterRouteWithMetaData(ctx, method, path, handler, "", nil, nil, nil, nil)
}

func (b *BaseApp) RegisterRouteWithMetaData(ctx context.Context, method, path string, handler http.HandlerFunc, description string, tags []string, payload interface{}, successResponse, failureResponse []Response) {
	val, ok := b.docMeta.Routes[path]
	if !ok {
		val = map[string]*APIHandler{}
		b.docMeta.Routes[path] = val
	}
	val[method] = &APIHandler{
		Func:            handler,
		Description:     description,
		Tags:            tags,
		Payload:         payload,
		SuccessResponse: successResponse,
		FailureResponse: failureResponse,
	}
	b.handler.HandlerFunc(method, path, handler)
}

func (b *BaseApp) RegisterDefaultRoutes(ctx context.Context) {
	b.handler.NotFound = NotFound()
	b.handler.MethodNotAllowed = MethodNotAllowed()
	b.RegisterRoutes(ctx, http.MethodGet, "/meta/health", HealthCheck)
}

func GetPathParams(ctx context.Context, log *log.Logger, r *http.Request) httprouter.Params {
	pp := r.Context().Value(httprouter.ParamsKey)
	pathParams, ok := pp.(httprouter.Params)
	if !ok {
		err := errors.NewHTTPServerError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Invalid path params processing", nil, map[string]interface{}{
			"pathParams": pp,
		})
		log.Emergency(ctx, "Invalid path params processing", err, err)
	}
	return pathParams
}
