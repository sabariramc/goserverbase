package kafka_test

import (
	"context"

	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/log/logwriter"
	"github.com/sabariramc/goserverbase/utils/testutils"
)

var KafkaTestConfig *testutils.TestConfig
var KafkaTestLogger *log.Logger

func init() {
	testutils.LoadEnv("../.env")
	testutils.Initialize()

	KafkaTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter(log.HostParams{
		Version:     KafkaTestConfig.Logger.Version,
		Host:        KafkaTestConfig.App.Host,
		ServiceName: KafkaTestConfig.App.ServiceName,
	})
	lmux := log.NewDefaultLogMux(consoleLogWriter)
	KafkaTestLogger = log.NewLogger(context.TODO(), KafkaTestConfig.Logger, "AWSTest", lmux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.CorrelationContextKey, log.GetDefaultCorrelationParams(KafkaTestConfig.App.ServiceName))
	return ctx
}
