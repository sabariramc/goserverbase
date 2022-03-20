package tests

import (
	"context"
	"testing"

	"sabariram.com/goserverbase/constant"
	"sabariram.com/goserverbase/db/mysql"
	"sabariram.com/goserverbase/log"
	"sabariram.com/goserverbase/log/logwriter"
	"sabariram.com/goserverbase/utils"
	"sabariram.com/goserverbase/utils/testutils"
)

var MysqlTestConfig *testutils.TestConfig
var MysqlTestLogger *log.Logger

func init() {
	testutils.Initialize()
	MysqlTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter(log.HostParams{
		Version:     MysqlTestConfig.Logger.Version,
		Host:        MysqlTestConfig.App.Host,
		ServiceName: MysqlTestConfig.App.ServiceName,
	})
	lmux := log.NewSequenctialLogMultipluxer(consoleLogWriter)
	MysqlTestLogger = log.NewLogger(context.TODO(), MysqlTestConfig.Logger, lmux, consoleLogWriter, utils.IST)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), constant.CorrelationContextKey, log.GetDefaultCorrelationParams(MysqlTestConfig.App.ServiceName))
	return ctx
}

func TestMysqlConnection(t *testing.T) {
	envVar := MysqlTestConfig.Mysql
	_ = mysql.NewConnection(envVar)
}
