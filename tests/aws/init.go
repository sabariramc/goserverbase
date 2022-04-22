package tests

import (
	"context"

	"github.com/sabariramc/goserverbase/constant"
	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/log/logwriter"
	"github.com/sabariramc/goserverbase/utils/testutils"
)

var AWSTestConfig *testutils.TestConfig
var AWSTestLogger *log.Logger

func init() {
	testutils.Initialize()
	AWSTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter(log.HostParams{
		Version:     AWSTestConfig.Logger.Version,
		Host:        AWSTestConfig.App.Host,
		ServiceName: AWSTestConfig.App.ServiceName,
	})
	lmux := log.NewSequenctialLogMultipluxer(consoleLogWriter)
	AWSTestLogger = log.NewLogger(context.TODO(), AWSTestConfig.Logger, lmux, consoleLogWriter)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), constant.CorrelationContextKey, log.GetDefaultCorrelationParams(AWSTestConfig.App.ServiceName))
	return ctx
}
