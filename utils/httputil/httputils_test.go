package httputil_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/log/logwriter"
	"github.com/sabariramc/goserverbase/v4/utils/httputil"
	"github.com/sabariramc/goserverbase/v4/utils/testutils"
	"gotest.tools/assert"
)

var HttpUtilTestConfig *testutils.TestConfig
var HttpUtilTestLogger *log.Logger

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

const URL = "http://localhost:64000/service/a/b"
const RetryURL = "http://localhost:64000/service/error/b"
const ErrURL = "http://localhost:80/service/error/b"

func TestHttpUtilGet(t *testing.T) {
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger)
	data := make(map[string]any)
	res, err := client.Get(GetCorrelationContext(), URL, nil, &data, nil)
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, 200)
}

func TestHttpUtilGetWithBody(t *testing.T) {
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger)
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
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger)
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
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger)
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
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger)
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
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger)
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
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger)
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
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger)
	data := make(map[string]any)
	body := map[string]any{
		"tag": "Test",
	}
	res, _ := client.Call(GetCorrelationContext(), "POST", RetryURL, &body, &data, nil)
	assert.Equal(t, res.StatusCode, 500)
}

func TestHttpUtilRetryError(t *testing.T) {
	client := httputil.NewDefaultHTTPClient(HttpUtilTestLogger)
	data := make(map[string]any)
	body := map[string]any{
		"tag": "Test",
	}
	res, err := client.Call(GetCorrelationContext(), "POST", ErrURL, &body, &data, nil)
	if err == nil || res != nil {
		t.Fail()
	}
}
