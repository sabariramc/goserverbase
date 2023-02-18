package baseapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sabariramc/goserverbase/constant"
	"github.com/sabariramc/goserverbase/errors"
	"github.com/sabariramc/goserverbase/log"
)

type APIRoute map[string]*APIResource

type APIResource struct {
	Handlers    map[string]*APIHandler
	SubResource map[string]*APIResource
}

type APIHandler struct {
	Func            http.HandlerFunc `json:"-"`
	Description     string
	Payload         interface{}
	SuccessResponse interface{}
	FailureReaponse map[int][]interface{}
}

func (b *BaseApp) registerResource(prefix string, router *httprouter.Router, route *APIResource) {
	for path, resource := range route.SubResource {
		b.registerResource(fmt.Sprintf("%v%v", prefix, path), router, resource)
	}
	for method, handler := range route.Handlers {
		router.HandlerFunc(method, prefix, handler.Func)
	}
}

func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(constant.HeaderContentType, constant.ContentTypeJSON)
		w.WriteHeader(http.StatusNotFound)
		body := errors.NewHTTPClientError(http.StatusNotFound, "Invalid path", map[string]string{
			"path": r.URL.Path,
		}).Error()
		w.Write([]byte(body))
	}
}

func MethodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(constant.HeaderContentType, constant.ContentTypeJSON)
		w.WriteHeader(http.StatusMethodNotAllowed)
		body := errors.NewHTTPClientError(http.StatusMethodNotAllowed, "Invalid method", map[string]string{
			"path":   r.URL.Path,
			"method": r.Method,
		}).Error()
		w.Write([]byte(body))
	}
}

func (b *BaseApp) RegisterRoutes(ctx context.Context, route *APIRoute) {
	for path, resource := range *route {
		b.registerResource(path, b.router, resource)
	}
	b.router.NotFound = NotFound()
	b.router.MethodNotAllowed = MethodNotAllowed()
}

func GetPathParams(ctx context.Context, log *log.Logger, r *http.Request) httprouter.Params {
	pp := r.Context().Value(httprouter.ParamsKey)
	pathParmas, ok := pp.(httprouter.Params)
	if !ok {
		err := errors.NewHTTPServerError(http.StatusInternalServerError, "Invalid path params processing", map[string]interface{}{
			"pathParams": pp,
		})
		log.Emergency(ctx, "Invalid path params processing", err, err)
	}
	return pathParmas
}
