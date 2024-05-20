package csfle_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/trace"
	"github.com/sabariramc/goserverbase/v6/testutils"
)

var MongoTestConfig *testutils.TestConfig
var MongoTestLogger log.Log

const ServiceName = "MongoDBCSFLETest"

func init() {
	testutils.LoadEnv("../../../.env")
	testutils.Initialize()
	MongoTestConfig = testutils.NewConfig()
	MongoTestLogger = log.New(log.WithServiceName("MongoDBCSFLETest"))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), trace.ContextKeyCorrelation, trace.GetDefaultCorrelationParam(ServiceName))
	return ctx
}
