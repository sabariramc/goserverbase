package httpserver_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sabariramc/goserverbase/v3/app/server/httpserver/test/server"
	"gotest.tools/assert"
)

func TestRouter(t *testing.T) {
	srv := server.NewServer()
	req := httptest.NewRequest(http.MethodGet, "/service/v1/tenant", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	assert.Equal(t, string(blob), "{\"body\":\"\"}")
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	req = httptest.NewRequest(http.MethodGet, "/service/v1/tenants/search", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = io.ReadAll(w.Body)
	assert.Equal(t, w.Result().StatusCode, http.StatusNotFound)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse := map[string]any{"errorDescription": map[string]any{"path": "/service/v1/tenants/search"}, "errorMessage": "Invalid path", "errorCode": "URL_NOT_FOUND"}
	assert.DeepEqual(t, res, expectedResponse)
	req = httptest.NewRequest(http.MethodGet, "/service/v1/tenant/tenant_ABC4567890abc", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = io.ReadAll(w.Body)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	assert.Equal(t, string(blob), "World")
	req = httptest.NewRequest(http.MethodPost, "/service/v1/tenant/tenant_ABC4567890abc", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = io.ReadAll(w.Body)
	res = make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse = map[string]any{"errorDescription": map[string]any{"method": "POST", "path": "/service/v1/tenant/tenant_ABC4567890abc"}, "errorMessage": "Invalid method", "errorCode": "METHOD_NOT_ALLOWED"}
	assert.Equal(t, w.Result().StatusCode, http.StatusMethodNotAllowed)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterCustomError(t *testing.T) {
	srv := server.NewServer()
	req := httptest.NewRequest(http.MethodGet, "/service/v1/error/error1", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	fmt.Println(string(blob))
	expectedResponse := map[string]any{"errorDescription": nil, "errorMessage": "display this", "errorCode": "hello.new.custom.error"}
	assert.Equal(t, w.Result().StatusCode, http.StatusInternalServerError)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterPanic(t *testing.T) {
	srv := server.NewServer()
	req := httptest.NewRequest(http.MethodGet, "/service/v1/error/error2", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse := map[string]any{"errorDescription": map[string]any{"error": "Internal error occurred, if persist contact technical team"}, "errorMessage": "Unknown error", "errorCode": "UNKNOWN"}
	assert.Equal(t, w.Result().StatusCode, http.StatusInternalServerError)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterClientError(t *testing.T) {
	srv := server.NewServer()
	req := httptest.NewRequest(http.MethodGet, "/service/v1/error/error3", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	fmt.Println(string(blob))
	expectedResponse := map[string]any{"errorDescription": nil, "errorMessage": "display this", "errorCode": "hello.new.custom.error"}
	assert.Equal(t, w.Result().StatusCode, 403)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterHealthCheck(t *testing.T) {
	srv := server.NewServer()
	req := httptest.NewRequest(http.MethodGet, "/meta/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assert.Equal(t, w.Result().StatusCode, 204)
}

func TestPost(t *testing.T) {
	srv := server.NewServer()
	payload, _ := json.Marshal(map[string]string{"fasdfas": "FASDFASf"})
	buff := bytes.NewBuffer(payload)
	req := httptest.NewRequest(http.MethodPost, "/service/v1/tenant", buff)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	expectedResponse := map[string]any{
		"body": "{\"fasdfas\":\"FASDFASf\"}",
	}
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	assert.DeepEqual(t, res, expectedResponse)
}
