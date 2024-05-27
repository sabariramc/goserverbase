package errors_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/testutils"
)

var TestLogger log.Log

const ServiceName = "ErrorTest"

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../.env")
	TestLogger = log.New(log.WithServiceName(ServiceName))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, correlation.NewCorrelationParam(ServiceName))
	return ctx
}
