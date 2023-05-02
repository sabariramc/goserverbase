package httputil_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sabariramc/goserverbase/v2/log"
	"github.com/sabariramc/goserverbase/v2/log/logwriter"
	"github.com/sabariramc/goserverbase/v2/utils/httputil"
	"github.com/sabariramc/goserverbase/v2/utils/testutils"
	"gotest.tools/assert"
)

var HttpUtilTestConfig *testutils.TestConfig
var HttpUtilTestLogger *log.Logger

func init() {
	testutils.LoadEnv("../../.env")
	testutils.Initialize()

	HttpUtilTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter(log.HostParams{
		Version:     HttpUtilTestConfig.Logger.Version,
		Host:        HttpUtilTestConfig.App.Host,
		ServiceName: HttpUtilTestConfig.App.ServiceName,
	})
	lMux := log.NewDefaultLogMux(consoleLogWriter)
	HttpUtilTestLogger = log.NewLogger(context.TODO(), HttpUtilTestConfig.Logger, "AWSTest", lMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParams(HttpUtilTestConfig.App.ServiceName))
	return ctx
}

func TestHttpUtilGet(t *testing.T) {
	client := httputil.NewDefaultHttpClient(HttpUtilTestLogger)
	data := make(map[string]any)
	res, err := client.Do(GetCorrelationContext(), "GET", "https://d6o0fhi2nl.execute-api.ap-south-1.amazonaws.com/dev/echo/go/base", nil, &data, nil)
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, 200)
}

func TestHttpUtilUnwrapError(t *testing.T) {
	client := httputil.NewDefaultHttpClient(HttpUtilTestLogger)
	data := ""
	body := map[string]string{
		"tag": "Test",
	}
	res, err := client.Do(GetCorrelationContext(), "POST", "https://d6o0fhi2nl.execute-api.ap-south-1.amazonaws.com/dev/echo/go/base", &body, &data, nil)
	if !errors.Is(err, httputil.ErrResponseUnmarshal) {
		t.Fail()
	}
	assert.Equal(t, res.StatusCode, 200)
}

func TestHttpUtilPost(t *testing.T) {
	client := httputil.NewDefaultHttpClient(HttpUtilTestLogger)
	data := make(map[string]any)
	body := map[string]string{
		"tag": "Test",
	}
	res, err := client.Do(GetCorrelationContext(), "POST", "https://d6o0fhi2nl.execute-api.ap-south-1.amazonaws.com/dev/echo/go/base", &body, &data, nil)
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, 200)
}
