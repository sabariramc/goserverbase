package aws_test

import (
	"context"
	"os"

	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/log/logwriter"
	"github.com/sabariramc/goserverbase/v5/testutils"
)

var AWSTestConfig *testutils.TestConfig
var AWSTestLogger log.Log

func init() {
	testutils.LoadEnv("../.env")
	testutils.Initialize()
	os.RemoveAll("./testdata/result")
	err := os.Mkdir("./testdata/result", 0755)
	if err != nil {
		panic(err)
	}
	AWSTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter()
	lMux := log.NewDefaultLogMux(consoleLogWriter)
	AWSTestLogger = log.NewLogger(context.TODO(), AWSTestConfig.Logger, "AWSTest", lMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(AWSTestConfig.App.ServiceName))
	return ctx
}
