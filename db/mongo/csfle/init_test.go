package csfle_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/log/logwriter"
	"github.com/sabariramc/goserverbase/v6/testutils"
)

var MongoTestConfig *testutils.TestConfig
var MongoTestLogger log.Log

func init() {
	testutils.LoadEnv("../../../.env")
	testutils.Initialize()
	MongoTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter()
	lMux := log.NewDefaultLogMux(consoleLogWriter)
	MongoTestLogger = log.New(context.TODO(), MongoTestConfig.Logger, "MongoTest", lMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(MongoTestConfig.App.ServiceName))
	return ctx
}
