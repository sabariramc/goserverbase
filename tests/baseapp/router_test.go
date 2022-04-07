package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"gotest.tools/assert"
	"sabariram.com/goserverbase/baseapp"
)

func Func1(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
	w.WriteHeader(200)
}

func Func2(w http.ResponseWriter, r *http.Request) {
	srv := &baseapp.BaseApp{}
	fmt.Println(srv.GetPathParams(r))
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
			"/:tenantId": {
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
	srv.SetRouter(httprouter.New())
	srv.RegisterRoutes(context.TODO(), route)
	req := httptest.NewRequest(http.MethodGet, "/tenant", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, string(blob), "Hello")
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	req = httptest.NewRequest(http.MethodGet, "/tenants/search", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, w.Result().StatusCode, http.StatusNotFound)
	res := string(blob)
	assert.Equal(t, res, "{\"errorData\":{\"path\":\"/tenants/search\"},\"errorMessage\":\"Invalid path\",\"errorCode\":\"NOT_FOUND\"}")
	req = httptest.NewRequest(http.MethodGet, "/tenant/tenant_ABC4567890abc", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	assert.Equal(t, string(blob), "World")
	req = httptest.NewRequest(http.MethodPost, "/tenant/tenant_ABC4567890abc", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = ioutil.ReadAll(w.Body)
	res = string(blob)
	assert.Equal(t, w.Result().StatusCode, http.StatusMethodNotAllowed)
	assert.Equal(t, res, "{\"errorData\":{\"method\":\"POST\",\"path\":\"/tenant/tenant_ABC4567890abc\"},\"errorMessage\":\"Invalid method\",\"errorCode\":\"INVALID_METHOD\"}")
}
