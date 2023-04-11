package csfle_test

import (
	"context"

	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/log/logwriter"
	"github.com/sabariramc/goserverbase/utils/testutils"
)

var MongoTestConfig *testutils.TestConfig
var MongoTestLogger *log.Logger

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../../../.env")
	MongoTestConfig = testutils.NewConfig()
	hostParams := log.HostParams{
		Version:     MongoTestConfig.Logger.Version,
		Host:        MongoTestConfig.App.Host,
		ServiceName: MongoTestConfig.App.ServiceName,
	}
	consoleLogWriter := logwriter.NewConsoleWriter(hostParams)

	lMux := log.NewDefaultLogMux(consoleLogWriter)
	MongoTestLogger = log.NewLogger(context.TODO(), MongoTestConfig.Logger, "MongoTest", lMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParams(MongoTestConfig.App.ServiceName))
	return ctx
}
