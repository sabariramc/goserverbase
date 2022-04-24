package mysql_test

import (
	"context"

	"github.com/sabariramc/goserverbase/constant"
	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/log/logwriter"
	"github.com/sabariramc/goserverbase/utils/testutils"
)

var MysqlTestConfig *testutils.TestConfig
var MysqlTestLogger *log.Logger

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../../.env")
	MysqlTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter(log.HostParams{
		Version:     MysqlTestConfig.Logger.Version,
		Host:        MysqlTestConfig.App.Host,
		ServiceName: MysqlTestConfig.App.ServiceName,
	})
	hostParams := log.HostParams{
		Version:     MysqlTestConfig.Logger.Version,
		Host:        MysqlTestConfig.App.Host,
		ServiceName: MysqlTestConfig.App.ServiceName,
	}
	graylog, err := logwriter.NewGraylogUDP(hostParams, consoleLogWriter, *MysqlTestConfig.Logger.GrayLog)
	if err != nil {
		panic(err)
	}
	lmux := log.NewSequenctialLogMultipluxer(consoleLogWriter, graylog)
	MysqlTestLogger = log.NewLogger(context.TODO(), MysqlTestConfig.Logger, lmux, consoleLogWriter, "Mysql Test", "test")
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), constant.CorrelationContextKey, log.GetDefaultCorrelationParams(MysqlTestConfig.App.ServiceName))
	return ctx
}
