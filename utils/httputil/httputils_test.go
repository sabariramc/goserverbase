package httputil_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/testutils"
	"github.com/sabariramc/goserverbase/v6/utils/httputil"
	"gotest.tools/assert"
)

var HTTPUtilTestLogger log.Log

const ContentTypeHeader = "Content-Type"
const MIMEJSON = "application/json"
const ServiceName = "HTTPClientTest"

func init() {
	testutils.LoadEnv("../../.env")
	testutils.Initialize()
	HTTPUtilTestLogger = log.New(log.WithServiceName(ServiceName))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, correlation.NewCorrelationParam(ServiceName))
	return ctx
}

const URL = "http://localhost:64000/service/v1/echo/a/b"
const RetryURL = "http://localhost:64000/service/v1/echo/error/b"
const ErrURL = "http://localhost:80/service/v1/echo/error/b"

func ExampleHTTPClient() {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	data := make(map[string]any)
	res, err := client.Get(GetCorrelationContext(), URL, nil, &data, nil)
	fmt.Println(err)
	fmt.Println(res.StatusCode)
	//Output:
	//<nil>
	//200
}

func ExampleHTTPClient_responsebody() {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	response := map[string]any{} // object to decode response body
	body := map[string]string{
		"tag": "Test",
	}
	//URL is a echo  endpoint that returns the whole request as response body
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, &correlation.CorrelationParam{
		CorrelationID: ServiceName,
	})
	res, err := client.Post(ctx, URL, &body, &response, map[string]string{ContentTypeHeader: MIMEJSON})
	fmt.Println(err)
	fmt.Println(res.StatusCode)
	fmt.Printf("%+v", response)
	//Output:
	//<nil>
	//200
	//map[body:map[tag:Test] headers:map[accept-encoding:gzip connection:close content-length:15 content-type:application/json host:backend user-agent:Go-http-client/1.1 x-correlation-id:HTTPClientTest] method:POST pathParams:map[path_1:a path_2:b] url:http://backend/service/v1/echo/a/b]
}

func ExampleHTTPClient_errorresponsebody() {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	response := 0
	body := map[string]string{
		"tag": "Test",
	}
	_, err := client.Post(GetCorrelationContext(), URL, &body, &response, map[string]string{ContentTypeHeader: MIMEJSON})
	fmt.Println(errors.Is(err, httputil.ErrResponseUnmarshal))
	//Output:
	//true
}

func ExampleHTTPClient_retry() {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	data := make(map[string]any)
	body := map[string]any{
		"tag": "Test",
	}
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, &correlation.CorrelationParam{
		CorrelationID: ServiceName,
	})
	//RetryURL is a  endpoint that returns the always returns 500
	res, err := client.Call(ctx, "POST", RetryURL, &body, &data, nil)
	fmt.Println(err)
	fmt.Println(res.StatusCode)
	//Output:
	//[2024-05-21T14:33:04.793+05:30] [INFO] [HTTPClientTest] [default] [log] [Request] [map[string]interface {}] [{
	//     "headers": {
	//         "X-Correlation-Id": [
	//             "HTTPClientTest"
	//         ]
	//     },
	//     "method": "POST",
	//     "url": "http://localhost:64000/service/v1/echo/error/b"
	// }]
	// [2024-05-21T14:33:04.801+05:30] [NOTICE] [HTTPClientTest] [default] [log] [request failed with status code 500 retry 1 of 4 in 10ms] [string] [{"url": "http://backend/service/v1/echo/error/b", "headers": {"host": "backend", "connection": "close", "content-length": "15", "user-agent": "Go-http-client/1.1", "x-correlation-id": "HTTPClientTest", "accept-encoding": "gzip"}, "method": "POST", "body": {"tag": "Test"}, "pathParams": {"path_1": "error", "path_2": "b"}}]
	// [2024-05-21T14:33:04.829+05:30] [NOTICE] [HTTPClientTest] [default] [log] [request failed with status code 500 retry 2 of 4 in 20ms] [string] [{"url": "http://backend/service/v1/echo/error/b", "headers": {"host": "backend", "connection": "close", "content-length": "15", "user-agent": "Go-http-client/1.1", "x-correlation-id": "HTTPClientTest", "accept-encoding": "gzip"}, "method": "POST", "body": {"tag": "Test"}, "pathParams": {"path_1": "error", "path_2": "b"}}]
	// [2024-05-21T14:33:04.868+05:30] [NOTICE] [HTTPClientTest] [default] [log] [request failed with status code 500 retry 3 of 4 in 40ms] [string] [{"url": "http://backend/service/v1/echo/error/b", "headers": {"host": "backend", "connection": "close", "content-length": "15", "user-agent": "Go-http-client/1.1", "x-correlation-id": "HTTPClientTest", "accept-encoding": "gzip"}, "method": "POST", "body": {"tag": "Test"}, "pathParams": {"path_1": "error", "path_2": "b"}}]
	// [2024-05-21T14:33:04.930+05:30] [NOTICE] [HTTPClientTest] [default] [log] [request failed with status code 500 retry 4 of 4 in 80ms] [string] [{"url": "http://backend/service/v1/echo/error/b", "headers": {"host": "backend", "connection": "close", "content-length": "15", "user-agent": "Go-http-client/1.1", "x-correlation-id": "HTTPClientTest", "accept-encoding": "gzip"}, "method": "POST", "body": {"tag": "Test"}, "pathParams": {"path_1": "error", "path_2": "b"}}]
	// [2024-05-21T14:33:05.031+05:30] [ERROR] [HTTPClientTest] [default] [log] [Response] [map[string]interface {}] [{
	//     "headers": {
	//         "Connection": [
	//             "keep-alive"
	//         ],
	//         "Content-Length": [
	//             "322"
	//         ],
	//         "Date": [
	//             "Tue, 21 May 2024 09:03:05 GMT"
	//         ],
	//         "Server": [
	//             "nginx/1.26.0"
	//         ]
	//     },
	//     "statusCode": 500
	// }]
	// HttpClient.Call: non 2xx status
	// 500
}

func TestHttpUtilGet(t *testing.T) {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	data := make(map[string]any)
	res, err := client.Get(GetCorrelationContext(), URL, nil, &data, nil)
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, 200)
}

func TestHttpUtilGetWithBody(t *testing.T) {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	data := make(map[string]any)
	body := map[string]any{
		"tag": "Test",
	}
	res, err := client.Get(GetCorrelationContext(), URL, &body, &data, map[string]string{ContentTypeHeader: MIMEJSON})
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, 200)
	assert.DeepEqual(t, body, data["body"])
	assert.DeepEqual(t, http.MethodGet, data["method"])
}

func TestHttpUtilUnwrapError(t *testing.T) {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	data := 0
	body := map[string]string{
		"tag": "Test",
	}
	_, err := client.Post(GetCorrelationContext(), URL, &body, &data, map[string]string{ContentTypeHeader: MIMEJSON})
	if errors.Is(err, httputil.ErrResponseUnmarshal) {
		return
	}
	t.Fail()
}

func TestHttpUtilPost(t *testing.T) {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	data := make(map[string]any)
	body := map[string]any{
		"tag": "Test",
	}
	res, err := client.Post(GetCorrelationContext(), URL, &body, &data, nil)
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, 200)
	assert.DeepEqual(t, body, data["body"])
	assert.DeepEqual(t, http.MethodPost, data["method"])
}

func TestHttpUtilPatch(t *testing.T) {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	data := make(map[string]any)
	body := map[string]any{
		"tag": "Test",
	}
	res, err := client.Patch(GetCorrelationContext(), URL, &body, &data, nil)
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, 200)
	assert.DeepEqual(t, body, data["body"])
	assert.DeepEqual(t, http.MethodPatch, data["method"])
}

func TestHttpUtilPut(t *testing.T) {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	data := make(map[string]any)
	body := map[string]any{
		"tag": "Test",
	}
	res, err := client.Put(GetCorrelationContext(), URL, &body, &data, nil)
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, 200)
	assert.DeepEqual(t, body, data["body"])
	assert.DeepEqual(t, http.MethodPut, data["method"])
}

func TestHttpUtilDelete(t *testing.T) {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	data := make(map[string]any)
	body := map[string]any{
		"tag": "Test",
	}
	res, err := client.Delete(GetCorrelationContext(), URL, &body, &data, nil)
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, 200)
	assert.DeepEqual(t, body, data["body"])
	assert.DeepEqual(t, http.MethodDelete, data["method"])
}

func TestHttpUtilRetry(t *testing.T) {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	data := make(map[string]any)
	body := map[string]any{
		"tag": "Test",
	}
	res, _ := client.Call(GetCorrelationContext(), "POST", RetryURL, &body, &data, nil)
	assert.Equal(t, res.StatusCode, 500)
}

func TestHttpUtilRetryError(t *testing.T) {
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger))
	data := make(map[string]any)
	body := map[string]any{
		"tag": "Test",
	}
	res, err := client.Call(GetCorrelationContext(), "POST", ErrURL, &body, &data, nil)
	if err == nil || res != nil {
		t.Fail()
	}
}

func TestConnectionReuse(t *testing.T) {
	ht := http.DefaultTransport.(*http.Transport).Clone()
	ht.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := httputil.New(httputil.WithLog(HTTPUtilTestLogger), httputil.WithHTTPClient(&http.Client{Transport: ht}))
	var wg sync.WaitGroup
	body, _ := json.Marshal(map[string]string{"fasdfsda": "fasdfas", "fasdfas": "fasdfas"})
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := GetCorrelationContext()
			for i := 0; i < 100; i++ {
				res, _ := client.Post(ctx, "https://localhost:60006/service/v1/test/all", bytes.NewBuffer(body), nil, map[string]string{
					ContentTypeHeader: MIMEJSON,
				})
				assert.Assert(t, res.ProtoAtLeast(2, 0))
			}
		}()
	}
	wg.Wait()
}

func TestH2CClient(t *testing.T) {
	client := httputil.NewH2CClient(httputil.WithLog(HTTPUtilTestLogger))
	var wg sync.WaitGroup
	ctx := GetCorrelationContext()
	body, _ := json.Marshal(map[string]any{"sabariram": 10})
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				res, err := client.Get(ctx, "https://localhost:60007/meta/health", body, nil, map[string]string{ContentTypeHeader: MIMEJSON})
				assert.NilError(t, err)
				assert.Assert(t, res.StatusCode < 300)
			}
		}()
	}
	wg.Wait()
}
