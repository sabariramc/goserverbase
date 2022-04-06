package baseapp

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
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

func (b *BaseApp) registerResource(router *mux.Router, route *APIResource) {
	for path, resource := range route.SubResource {
		b.registerResource(router.PathPrefix(path).Subrouter(), resource)
	}
	for method, handler := range route.Handlers {
		router.Methods(method).HandlerFunc(handler.Func)
	}
}

func (b *BaseApp) RegisterRoutes(ctx context.Context, route *APIRoute) {
	for path, resource := range *route {
		b.registerResource(b.router.PathPrefix(path).Subrouter(), resource)
	}
	err := b.router.Walk(b.gorillaWalkFn(ctx))
	if err != nil {
		b.log.Emergency(ctx, "Error at BaseApp.RegisterRoutes Walk", nil, err)
	}
}

func (b *BaseApp) gorillaWalkFn(ctx context.Context) mux.WalkFunc {
	return func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		b.log.Debug(ctx, path, "")
		if err != nil {
			b.log.Emergency(ctx, "Error at BaseApp.gorillaWalkFn Walk", nil, err)
		}
		return nil
	}
}
