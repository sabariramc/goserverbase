package errors_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/testutils"
	"github.com/sabariramc/goserverbase/v6/correlation"
)

var TestConfig *testutils.TestConfig
var TestLogger log.Log

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../.env")
	TestConfig = testutils.NewConfig()
	TestLogger = log.New(log.WithServiceName("ErrorTest"))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, correlation.GetDefaultCorrelationParam(TestConfig.App.ServiceName))
	return ctx
}
