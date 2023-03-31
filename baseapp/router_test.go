package baseapp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sabariramc/goserverbase/baseapp"
	"github.com/sabariramc/goserverbase/errors"
	"gotest.tools/assert"
)

type server struct {
	*baseapp.BaseApp
}

func (s *server) Func1(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
	w.WriteHeader(200)
}

func (s *server) Func2(w http.ResponseWriter, r *http.Request) {
	fmt.Println(baseapp.GetPathParams(r.Context(), ServerTestLogger, r))
	w.Write([]byte("World"))
	w.WriteHeader(200)
}

func (s *server) Func3(w http.ResponseWriter, r *http.Request) {
	s.SetHandlerError(r.Context(), errors.NewCustomError("hello.new.custom.error", "display this", map[string]any{"one": "two"}, nil, true))
}

func (s *server) Func4(w http.ResponseWriter, r *http.Request) {
	panic("random panic at Func4")
}

func (s *server) Func5(w http.ResponseWriter, r *http.Request) {
	s.SetHandlerError(r.Context(), errors.NewHTTPClientError(403, "hello.new.custom.error", "display this", map[string]any{"one": "two"}, nil))
}

func NewServer() *server {
	srv := &server{
		BaseApp: baseapp.New(*ServerTestConfig.App, *ServerTestConfig.Logger, ServerTestLMux, nil, nil),
	}
	srv.RegisterRoutes(context.TODO(), http.MethodGet, "/tenant", srv.Func1)
	srv.RegisterRoutes(context.TODO(), http.MethodGet, "/tenant/:tenantId", srv.Func2)
	srv.RegisterRoutes(context.TODO(), http.MethodGet, "/error/error1", srv.Func3)
	srv.RegisterRoutes(context.TODO(), http.MethodGet, "/error/error2", srv.Func4)
	srv.RegisterRoutes(context.TODO(), http.MethodGet, "/error/error3", srv.Func5)
	return srv
}

func TestRouter(t *testing.T) {
	srv := NewServer()
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
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse := map[string]any{"errorDescription": map[string]any{"path": "/tenants/search"}, "errorMessage": "Invalid path", "errorCode": "URL_NOT_FOUND"}
	assert.DeepEqual(t, res, expectedResponse)
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
	res = make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse = map[string]any{"errorDescription": map[string]any{"method": "POST", "path": "/tenant/tenant_ABC4567890abc"}, "errorMessage": "Invalid method", "errorCode": "METHOD_NOT_ALLOWED"}
	assert.Equal(t, w.Result().StatusCode, http.StatusMethodNotAllowed)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterCustomError(t *testing.T) {
	srv := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/error/error1", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := ioutil.ReadAll(w.Body)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	fmt.Println(string(blob))
	expectedResponse := map[string]any{"errorDescription": nil, "errorMessage": "display this", "errorCode": "hello.new.custom.error"}
	assert.Equal(t, w.Result().StatusCode, http.StatusInternalServerError)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterPanic(t *testing.T) {
	srv := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/error/error2", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := ioutil.ReadAll(w.Body)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	fmt.Println(string(blob))
	expectedResponse := map[string]any{"errorDescription": map[string]any{"error": "Internal error occurred, if persist contact technical team"}, "errorMessage": "Unknown error", "errorCode": "UNKNOWN"}
	assert.Equal(t, w.Result().StatusCode, http.StatusInternalServerError)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterClientError(t *testing.T) {
	srv := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/error/error3", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := ioutil.ReadAll(w.Body)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	fmt.Println(string(blob))
	expectedResponse := map[string]any{"errorDescription": nil, "errorMessage": "display this", "errorCode": "hello.new.custom.error"}
	assert.Equal(t, w.Result().StatusCode, 403)
	assert.DeepEqual(t, res, expectedResponse)
}
