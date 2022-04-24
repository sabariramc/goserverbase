package mongo

import (
	"context"

	"github.com/sabariramc/goserverbase/constant"
	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/log/logwriter"
	"github.com/sabariramc/goserverbase/utils/testutils"
)

var MongoTestConfig *testutils.TestConfig
var MongoTestLogger *log.Logger

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../../.env")
	MongoTestConfig = testutils.NewConfig()
	hostParams := log.HostParams{
		Version:     MongoTestConfig.Logger.Version,
		Host:        MongoTestConfig.App.Host,
		ServiceName: MongoTestConfig.App.ServiceName,
	}
	consoleLogWriter := logwriter.NewConsoleWriter(hostParams)
	graylog, err := logwriter.NewGraylogUDP(hostParams, consoleLogWriter, *MongoTestConfig.Logger.GrayLog)
	if err != nil {
		panic(err)
	}
	lmux := log.NewSequenctialLogMultipluxer(consoleLogWriter, graylog)
	MongoTestLogger = log.NewLogger(context.TODO(), MongoTestConfig.Logger, lmux, consoleLogWriter, "MongoTest", "test")
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), constant.CorrelationContextKey, log.GetDefaultCorrelationParams(MongoTestConfig.App.ServiceName))
	return ctx
}
