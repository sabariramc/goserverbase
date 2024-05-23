package kafka_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/testutils"
)

var KafkaTestConfig *testutils.TestConfig
var KafkaTestLogger log.Log

const ServiceName = "KafkaTest"

func init() {
	testutils.LoadEnv("../.env")
	KafkaTestConfig = testutils.NewConfig()
	KafkaTestLogger = log.New(log.WithServiceName(ServiceName))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, correlation.NewCorrelationParam(ServiceName))
	return ctx
}
