package csfle_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/log/logwriter"
	"github.com/sabariramc/goserverbase/v4/utils/testutils"
)

var MongoTestConfig *testutils.TestConfig
var MongoTestLogger *log.Logger

func init() {
	testutils.LoadEnv("../../../.env")
	testutils.Initialize()
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
