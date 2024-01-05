package kafka_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/log/logwriter"
	"github.com/sabariramc/goserverbase/v4/testutils"
)

var KafkaTestConfig *testutils.TestConfig
var KafkaTestLogger *log.Logger

func init() {
	testutils.LoadEnv("../.env")
	testutils.Initialize()

	KafkaTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter()
	lMux := log.NewDefaultLogMux(consoleLogWriter)
	KafkaTestLogger = log.NewLogger(context.TODO(), KafkaTestConfig.Logger, "KafkaTest", lMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(KafkaTestConfig.App.ServiceName))
	return ctx
}
