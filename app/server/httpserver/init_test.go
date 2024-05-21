package httpserver_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/testutils"
	"github.com/sabariramc/goserverbase/v6/correlation"
)

var ServerTestConfig *testutils.TestConfig
var ServerTestLogger log.Log
var ServerTestLMux log.Mux

const ServiceName = "BaseTest"

func init() {
	testutils.LoadEnv("../../../.env")
	testutils.Initialize()
	ServerTestConfig = testutils.NewConfig()
	ServerTestLogger = log.New(log.WithServiceName(ServiceName))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, correlation.GetDefaultCorrelationParam(ServiceName))
	return ctx
}
