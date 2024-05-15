package httpserver_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	server "github.com/sabariramc/goserverbase/v6/app/server/httpserver/test"
	"gotest.tools/assert"
)

func TestRouter(t *testing.T) {
	srv := server.NewServer(nil)
	req := httptest.NewRequest(http.MethodGet, "/service/echo", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	assert.Equal(t, string(blob), `{"body":"","headers":{},"pathParams":"","queryParams":{}}`)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	req = httptest.NewRequest(http.MethodGet, "/service/echos/search", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = io.ReadAll(w.Body)
	assert.Equal(t, w.Result().StatusCode, http.StatusNotFound)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse := map[string]any{"errorDescription": map[string]any{"path": "/service/echos/search"}, "errorMessage": "Invalid path", "errorCode": "URL_NOT_FOUND"}
	assert.DeepEqual(t, res, expectedResponse)
	req = httptest.NewRequest(http.MethodGet, "/service/echo/tenant_ABC4567890abc", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = io.ReadAll(w.Body)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	assert.Equal(t, string(blob), `{"body":"","headers":{},"pathParams":"/tenant_ABC4567890abc","queryParams":{}}`)
	req = httptest.NewRequest(http.MethodPost, "/service/echo/tenant_ABC4567890abc", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = io.ReadAll(w.Body)
	res = make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse = map[string]any{"errorDescription": map[string]any{"method": "POST", "path": "/service/echo/tenant_ABC4567890abc"}, "errorMessage": "Invalid method", "errorCode": "METHOD_NOT_ALLOWED"}
	assert.Equal(t, w.Result().StatusCode, http.StatusMethodNotAllowed)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterCustomError(t *testing.T) {
	srv := server.NewServer(nil)
	req := httptest.NewRequest(http.MethodGet, "/service/error/error500", nil)
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
	srv := server.NewServer(nil)
	req := httptest.NewRequest(http.MethodGet, "/service/error/errorWithPanic", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse := map[string]any{"errorDescription": nil, "errorMessage": "display this", "errorCode": "hello.new.custom.error"}
	assert.Equal(t, w.Result().StatusCode, http.StatusServiceUnavailable)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterPanic2(t *testing.T) {
	srv := server.NewServer(nil)
	req := httptest.NewRequest(http.MethodGet, "/service/error/panic", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse := map[string]any{"errorDescription": map[string]any{"error": "Internal error occurred, if persist contact technical team"}, "errorMessage": "Unknown error", "errorCode": "com.base.internalServerError"}
	assert.Equal(t, w.Result().StatusCode, http.StatusInternalServerError)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterClientError(t *testing.T) {
	srv := server.NewServer(nil)
	req := httptest.NewRequest(http.MethodGet, "/service/error/errorUnauthorized", nil)
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
	srv := server.NewServer(nil)
	req := httptest.NewRequest(http.MethodGet, "/meta/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assert.Equal(t, w.Result().StatusCode, 204)
}

func TestPost(t *testing.T) {
	srv := server.NewServer(nil)
	payload, _ := json.Marshal(map[string]string{"fasdfas": "FASDFASf"})
	buff := bytes.NewBuffer(payload)
	req := httptest.NewRequest(http.MethodPost, "/service/echo", buff)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	expectedResponse := map[string]any{
		"body":        `{"fasdfas":"FASDFASf"}`,
		"headers":     map[string]any{},
		"pathParams":  "",
		"queryParams": map[string]any{},
	}
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestIntegration(t *testing.T) {
	srv := server.NewServer(nil)
	payload, _ := json.Marshal(map[string]string{"fasdfas": "FASDFASf"})
	buff := bytes.NewBuffer(payload)
	req := httptest.NewRequest(http.MethodPost, "/service/test/all", buff)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assert.Equal(t, w.Result().StatusCode, http.StatusNoContent)
}

const (
	start = 1 // actual = start  * goprocs
	end   = 8 // actual = end    * goprocs
	step  = 1
)

var goprocs = runtime.GOMAXPROCS(0) // 8

func TestBencRoute(t *testing.T) {
	srv := server.NewServer(nil)
	payload, _ := json.Marshal(map[string]string{"fasdfas": "FASDFASf"})
	buff := bytes.NewBuffer(payload)
	req := httptest.NewRequest(http.MethodGet, "/meta/bench", buff)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assert.Equal(t, w.Result().StatusCode, http.StatusNoContent)
}

func BenchmarkRoutes(b *testing.B) {
	srv := server.NewServer(nil)
	payload, _ := json.Marshal(map[string]string{"fasdfas": "FASDFASf"})
	buff := bytes.NewBuffer(payload)
	for i := start; i < end; i += step {
		b.Run(fmt.Sprintf("goroutines-%d", i*goprocs), func(b *testing.B) {
			b.SetParallelism(i)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					req := httptest.NewRequest(http.MethodGet, "/meta/bench", buff)
					w := httptest.NewRecorder()
					srv.ServeHTTP(w, req)
					assert.Equal(b, w.Result().StatusCode, http.StatusNoContent)
				}
			})
		})
	}
}
