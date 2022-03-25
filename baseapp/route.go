package baseapp

import (
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
	for method, handler := range route.Handlers {
		router.Methods(method).HandlerFunc(handler.Func)
	}
	for path, resource := range route.SubResource {
		b.registerResource(router.PathPrefix(path).Subrouter(), resource)
	}
}

func (b *BaseApp) RegisterRoutes(route *APIRoute) {
	for path, resource := range *route {
		b.registerResource(b.router.PathPrefix(path).Subrouter(), resource)
	}
}
