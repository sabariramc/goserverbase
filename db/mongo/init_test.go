package mongo_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/sabariramc/goserverbase/v3/log/logwriter"
	"github.com/sabariramc/goserverbase/v3/utils/testutils"
)

var MongoTestConfig *testutils.TestConfig
var MongoTestLogger *log.Logger

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../../.env")
	MongoTestConfig = testutils.NewConfig()
	hostParams := log.HostParams{
		Version:     MongoTestConfig.Logger.Version,
		Host:        MongoTestConfig.Http.Host,
		ServiceName: MongoTestConfig.App.ServiceName,
	}
	consoleLogWriter := logwriter.NewConsoleWriter(hostParams)
	lMux := log.NewDefaultLogMux(consoleLogWriter)
	MongoTestLogger = log.NewLogger(context.TODO(), MongoTestConfig.Logger, "MongoTest", lMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(MongoTestConfig.App.ServiceName))
	return ctx
}
