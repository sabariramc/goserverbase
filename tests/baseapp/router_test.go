package tests

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"gotest.tools/assert"
	"sabariram.com/goserverbase/baseapp"
)

func Func1(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
	w.WriteHeader(200)
}

func Func2(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("World"))
	w.WriteHeader(200)
}

var route = &baseapp.APIRoute{
	"/tenant": &baseapp.APIResource{
		Handlers: map[string]*baseapp.APIHandler{
			http.MethodGet: {
				Func: Func1,
			},
		},
		SubResource: map[string]*baseapp.APIResource{
			"/{tenantId:tenant_[0-9a-zA-Z]{13}}": {
				Handlers: map[string]*baseapp.APIHandler{
					http.MethodGet: {
						Func: Func2,
					},
				},
			},
		},
	},
}

func TestRouter(t *testing.T) {
	srv := &baseapp.BaseApp{}
	srv.SetConfig(baseapp.ServerConfig{
		LoggerConfig: ServerTestConfig.Logger,
		AppConfig:    ServerTestConfig.App,
	})
	srv.SetLogger(ServerTestLogger)
	srv.SetRouter(mux.NewRouter().StrictSlash(true))
	srv.RegisterRoutes(context.TODO(), route)
	req := httptest.NewRequest(http.MethodGet, "/tenant", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, string(blob), "Hello")
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	req = httptest.NewRequest(http.MethodGet, "/tenant/search", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, w.Result().StatusCode, http.StatusNotFound)
	res := string(blob)
	assert.Equal(t, res, "{\"errorData\":{\"path\":\"/tenant/search\"},\"errorMessage\":\"Invalid path\",\"errorCode\":\"NOT_FOUND\"}")
	req = httptest.NewRequest(http.MethodGet, "/tenant/tenant_ABC4567890abc", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	assert.Equal(t, string(blob), "World")

}
