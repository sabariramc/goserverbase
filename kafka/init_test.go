package kafka_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/testutils"
	"github.com/sabariramc/goserverbase/v6/trace"
)

var KafkaTestConfig *testutils.TestConfig
var KafkaTestLogger log.Log

func init() {
	testutils.LoadEnv("../.env")
	testutils.Initialize()
	KafkaTestConfig = testutils.NewConfig()
	KafkaTestLogger = log.New(log.WithServiceName("KafkaTest"))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), trace.ContextKeyCorrelation, trace.GetDefaultCorrelationParam(KafkaTestConfig.App.ServiceName))
	return ctx
}
