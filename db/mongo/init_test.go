package mongo_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/testutils"
	"github.com/sabariramc/goserverbase/v6/trace"
)

var MongoTestConfig *testutils.TestConfig
var MongoTestLogger log.Log

const ServiceName = "MongoDBTest"

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../../.env")
	MongoTestConfig = testutils.NewConfig()
	MongoTestLogger = log.New(log.WithServiceName(ServiceName))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), trace.ContextKeyCorrelation, trace.GetDefaultCorrelationParam(ServiceName))
	return ctx
}
