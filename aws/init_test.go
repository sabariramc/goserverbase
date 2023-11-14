package aws_test

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/sabariramc/goserverbase/v4/aws"
	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/log/logwriter"
	"github.com/sabariramc/goserverbase/v4/utils/testutils"
)

var AWSTestConfig *testutils.TestConfig
var AWSTestLogger *log.Logger

func init() {
	testutils.LoadEnv("../.env")
	testutils.Initialize()
	os.RemoveAll("./testdata/result")
	err := os.Mkdir("./testdata/result", 0755)
	if err != nil {
		panic(err)
	}
	AWSTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter(log.HostParams{
		Version:     AWSTestConfig.Logger.Version,
		Host:        AWSTestConfig.Http.Host,
		ServiceName: AWSTestConfig.App.ServiceName,
	})
	lMux := log.NewDefaultLogMux(consoleLogWriter)
	AWSTestLogger = log.NewLogger(context.TODO(), AWSTestConfig.Logger, "AWSTest", lMux, nil)
	defaultConfig, _ := config.LoadDefaultConfig(context.TODO())
	aws.SetDefaultAWSConfig(defaultConfig)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(AWSTestConfig.App.ServiceName))
	return ctx
}
