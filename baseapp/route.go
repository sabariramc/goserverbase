package baseapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
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

func (b *BaseApp) registerResource(prefix string, router *mux.Router, route *APIResource) {
	for path, resource := range route.SubResource {
		b.registerResource(fmt.Sprintf("%v%v", prefix, path), router, resource)
	}
	for method, handler := range route.Handlers {
		router.HandleFunc(prefix, handler.Func).Methods(method)
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
	b.router.NotFoundHandler = NotFound()
	b.router.MethodNotAllowedHandler = MethodNotAllowed()
	if b.c.AppConfig.Debug {
		err := b.router.Walk(b.GorillaWalkFn(ctx))
		if err != nil {
			b.log.Emergency(ctx, "Error at BaseApp.RegisterRoutes Walk", nil, err)
		}
	}
}

func (b *BaseApp) GorillaWalkFn(ctx context.Context) mux.WalkFunc {
	return func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		b.log.Debug(ctx, path, "")
		if err != nil {
			b.log.Emergency(ctx, "Error at BaseApp.gorillaWalkFn Walk", nil, err)
		}
		return nil
	}
}
