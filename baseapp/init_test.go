package baseapp_test

import (
	"context"

	"github.com/sabariramc/goserverbase/constant"
	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/log/logwriter"
	"github.com/sabariramc/goserverbase/utils/testutils"
)

var ServerTestConfig *testutils.TestConfig
var ServerTestLogger *log.Logger
var ServerTestLMux log.LogMux
var ServerTestAuditLogger log.AuditLogWriter

func init() {
	testutils.Initialize()
	ServerTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter(log.HostParams{
		Version:     ServerTestConfig.Logger.Version,
		Host:        ServerTestConfig.App.Host,
		ServiceName: ServerTestConfig.App.ServiceName,
	})
	ServerTestAuditLogger = consoleLogWriter
	ServerTestLMux = log.NewDefaultLogMux(consoleLogWriter)
	ServerTestLogger = log.NewLogger(context.TODO(), ServerTestConfig.Logger, ServerTestLMux, consoleLogWriter, "BaseTest")
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), constant.CorrelationContextKey, log.GetDefaultCorrelationParams(ServerTestConfig.App.ServiceName))
	return ctx
}
