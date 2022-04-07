package baseapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"sabariram.com/goserverbase/constant"
	"sabariram.com/goserverbase/errors"
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
		body := errors.NewCustomError("NOT_FOUND", "Invalid path", map[string]string{
			"path": r.URL.Path,
		}).Error()
		w.Write([]byte(body))
	}
}

func MethodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(constant.HeaderContentType, constant.ContentTypeJSON)
		w.WriteHeader(http.StatusMethodNotAllowed)
		body := errors.NewCustomError("INVALID_METHOD", "Invalid method", map[string]string{
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

func GetPathParams(r *http.Request) httprouter.Params {
	pp := r.Context().Value(httprouter.ParamsKey)
	pathParmas, ok := pp.(httprouter.Params)
	if !ok {
		panic(errors.NewCustomError("INVALID_PATH_PARAM", "Invalid path params processing", map[string]interface{}{
			"pathParams": pp,
		}))
	}
	return pathParmas
}
