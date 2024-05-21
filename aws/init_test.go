package aws_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/testutils"
)

var AWSTestConfig *testutils.TestConfig
var AWSTestLogger log.Log

const ServiceName = "AWSTest"

func init() {
	testutils.LoadEnv("../../../.env")
	AWSTestConfig = testutils.NewConfig()
	AWSTestLogger = log.New(log.WithServiceName(ServiceName))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, correlation.NewCorrelationParam(ServiceName))
	return ctx
}
