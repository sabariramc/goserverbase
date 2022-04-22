package tests

import (
	"context"
	"testing"

	"github.com/sabariramc/goserverbase/constant"
	"github.com/sabariramc/goserverbase/db/mysql"
	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/log/logwriter"
	"github.com/sabariramc/goserverbase/utils/testutils"
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
	MysqlTestLogger = log.NewLogger(context.TODO(), MysqlTestConfig.Logger, lmux, consoleLogWriter)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), constant.CorrelationContextKey, log.GetDefaultCorrelationParams(MysqlTestConfig.App.ServiceName))
	return ctx
}

func TestMysqlConnection(t *testing.T) {
	envVar := MysqlTestConfig.Mysql
	_ = mysql.NewConnection(envVar)
}
