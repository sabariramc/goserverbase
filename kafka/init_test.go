package kafka_test

import (
	"context"

	"github.com/sabariramc/goserverbase/constant"
	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/log/logwriter"
	"github.com/sabariramc/goserverbase/utils"
	"github.com/sabariramc/goserverbase/utils/testutils"
)

var KafkaTestConfig *testutils.TestConfig
var KafkaTestLogger *log.Logger

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../.env")
	KafkaTestConfig = testutils.NewConfig()
	hostParams := log.HostParams{
		Version:     KafkaTestConfig.Logger.Version,
		Host:        utils.GetHostName(),
		ServiceName: KafkaTestConfig.Logger.ServiceName,
	}
	consoleLogWriter := logwriter.NewConsoleWriter(hostParams)
	graylog, err := logwriter.NewGraylogUDP(hostParams, consoleLogWriter, *KafkaTestConfig.Logger.GrayLog)
	if err != nil {
		panic(err)
	}
	lmux := log.NewSequenctialLogMultipluxer(consoleLogWriter, graylog)
	KafkaTestLogger = log.NewLogger(context.TODO(), KafkaTestConfig.Logger, lmux, consoleLogWriter, "KAFKA_TEST")
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), constant.CorrelationContextKey, log.GetDefaultCorrelationParams(KafkaTestConfig.Logger.ServiceName))
	return ctx
}
