package errors_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/log/logwriter"
	"github.com/sabariramc/goserverbase/v6/testutils"
)

var TestConfig *testutils.TestConfig
var TestLogger log.Log

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../.env")
	TestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter()
	lMux := log.NewDefaultLogMux(consoleLogWriter)
	TestLogger = log.New(context.TODO(), TestConfig.Logger, "ErrorTest", lMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(TestConfig.App.ServiceName))
	return ctx
}
