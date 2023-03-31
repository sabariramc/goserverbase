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
		body := errors.NewHTTPClientError(http.StatusNotFound, "Invalid path", map[string]string{
			"path": r.URL.Path,
		}, nil).Error()
		w.Write([]byte(body))
	}
}

func MethodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(HttpHeaderContentType, HttpContentTypeJSON)
		w.WriteHeader(http.StatusMethodNotAllowed)
		body := errors.NewHTTPClientError(http.StatusMethodNotAllowed, "Invalid method", map[string]string{
			"path":   r.URL.Path,
			"method": r.Method,
		}, nil).Error()
		w.Write([]byte(body))
	}
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
	b.router.HandlerFunc(method, path, handler)
}

func (b *BaseApp) RegisterDefaultRoutes(ctx context.Context) {
	b.router.NotFound = NotFound()
	b.router.MethodNotAllowed = MethodNotAllowed()
}

func GetPathParams(ctx context.Context, log *log.Logger, r *http.Request) httprouter.Params {
	pp := r.Context().Value(httprouter.ParamsKey)
	pathParams, ok := pp.(httprouter.Params)
	if !ok {
		err := errors.NewHTTPServerError(http.StatusInternalServerError, "Invalid path params processing", map[string]interface{}{
			"pathParams": pp,
		}, nil)
		log.Emergency(ctx, "Invalid path params processing", err, err)
	}
	return pathParams
}
