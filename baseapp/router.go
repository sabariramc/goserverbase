package baseapp

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type HttpRouter struct {
	*httprouter.Router
}

func NewRouter() HttpRouter {
	return HttpRouter{
		Router: httprouter.New(),
	}
}

func (r *HttpRouter) SetNotFound(handler http.HandlerFunc) {
	r.NotFound = handler
}

func (r *HttpRouter) SetMethodNotAllowed(handler http.HandlerFunc) {
	r.MethodNotAllowed = handler
}

func (r *HttpRouter) RegisterRoute(method, path string, handler http.HandlerFunc) {
	r.HandlerFunc(method, path, handler)
}
