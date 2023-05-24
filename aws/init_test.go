package aws_test

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sabariramc/goserverbase/v2/aws"
	"github.com/sabariramc/goserverbase/v2/log"
	"github.com/sabariramc/goserverbase/v2/log/logwriter"
	"github.com/sabariramc/goserverbase/v2/utils/testutils"
)

var AWSTestConfig *testutils.TestConfig
var AWSTestLogger *log.Logger

func init() {
	testutils.LoadEnv("../.env")
	testutils.Initialize()

	AWSTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter(log.HostParams{
		Version:     AWSTestConfig.Logger.Version,
		Host:        AWSTestConfig.Http.Host,
		ServiceName: AWSTestConfig.App.ServiceName,
	})
	lMux := log.NewDefaultLogMux(consoleLogWriter)
	AWSTestLogger = log.NewLogger(context.TODO(), AWSTestConfig.Logger, "AWSTest", lMux, nil)
	aws.SetDefaultAWSSession(session.Must(session.NewSession()))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(AWSTestConfig.App.ServiceName))
	return ctx
}
