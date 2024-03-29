package aes_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/log/logwriter"
	"github.com/sabariramc/goserverbase/v5/testutils"
)

var ServerTestConfig *testutils.TestConfig
var ServerTestLogger log.Log
var ServerTestLMux log.LogMux

func init() {
	testutils.LoadEnv("../../.env")
	testutils.Initialize()
	ServerTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter()
	lmux := log.NewDefaultLogMux(consoleLogWriter)
	ServerTestLogger = log.NewLogger(context.TODO(), ServerTestConfig.Logger, "CRYPTOTEST", lmux, nil)
	ServerTestLMux = lmux
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(ServerTestConfig.App.ServiceName))
	return ctx
}
