package httputil_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/log/logwriter"
	"github.com/sabariramc/goserverbase/v5/testutils"
	"github.com/sabariramc/goserverbase/v5/utils/httputil"
	"gotest.tools/assert"
)

var HttpUtilTestConfig *testutils.TestConfig
var HttpUtilTestLogger log.Log

func init() {
	testutils.LoadEnv("../../.env")
	testutils.Initialize()

	HttpUtilTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter()
	lMux := log.NewDefaultLogMux(consoleLogWriter)
	HttpUtilTestLogger = log.NewLogger(context.TODO(), HttpUtilTestConfig.Logger, "AWSTest", lMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(HttpUtilTestConfig.App.ServiceName))
	return ctx
}

const URL = "http://localhost:64000/service/v1/echo/a/b"
const RetryURL = "http://localhost:64000/service/v1/echo/error/b"
const ErrURL = "http://localhost:80/service/v1/echo/error/b"

func TestHttpUtilGet(t *testing.T) {
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger, nil)
	data := make(map[string]any)
	res, err := client.Get(GetCorrelationContext(), URL, nil, &data, nil)
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, 200)
}

func TestHttpUtilGetWithBody(t *testing.T) {
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger, nil)
	data := make(map[string]any)
	body := map[string]any{
		"tag": "Test",
	}
	res, err := client.Get(GetCorrelationContext(), URL, &body, &data, map[string]string{"Content-Type": "application/json"})
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, 200)
	assert.DeepEqual(t, body, data["body"])
	assert.DeepEqual(t, http.MethodGet, data["method"])
}

func TestHttpUtilUnwrapError(t *testing.T) {
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger, nil)
	data := 0
	body := map[string]string{
		"tag": "Test",
	}
	_, err := client.Post(GetCorrelationContext(), URL, &body, &data, map[string]string{"Content-Type": "application/json"})
	if errors.Is(err, httputil.ErrResponseUnmarshal) {
		return
	}
	t.Fail()
}

func TestHttpUtilPost(t *testing.T) {
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger, nil)
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
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger, nil)
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
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger, nil)
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
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger, nil)
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
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger, nil)
	data := make(map[string]any)
	body := map[string]any{
		"tag": "Test",
	}
	res, _ := client.Call(GetCorrelationContext(), "POST", RetryURL, &body, &data, nil)
	assert.Equal(t, res.StatusCode, 500)
}

func TestHttpUtilRetryError(t *testing.T) {
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger, nil)
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
	client := httputil.New(HttpUtilTestLogger, nil, &http.Client{Transport: ht}, 4, time.Second*1, time.Second*5)
	var wg sync.WaitGroup
	body, _ := json.Marshal(map[string]string{"fasdfsda": "fasdfas", "fasdfas": "fasdfas"})
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := GetCorrelationContext()
			for i := 0; i < 100; i++ {
				res, _ := client.Post(ctx, "https://localhost:60006/service/v1/test/all", bytes.NewBuffer(body), nil, map[string]string{
					"Content-Type": "application/json",
				})
				assert.Assert(t, res.ProtoAtLeast(2, 0))
			}
		}()
	}
	wg.Wait()
}

func TestH2CClient(t *testing.T) {
	client := httputil.NewH2CClient(HttpUtilTestLogger, nil, 4, time.Second, 4*time.Second)
	var wg sync.WaitGroup
	ctx := GetCorrelationContext()
	body, _ := json.Marshal(map[string]any{"sabariram": 10})
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				res, err := client.Post(ctx, "https://localhost:60007/service/v1/test/req", body, nil, map[string]string{"Content-Type": "application/json"})
				assert.NilError(t, err)
				assert.Assert(t, res.StatusCode < 300)
			}
		}()
	}
	wg.Wait()
}
