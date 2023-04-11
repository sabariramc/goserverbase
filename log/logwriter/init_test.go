package logwriter_test

import (
	"github.com/sabariramc/goserverbase/v2/log"
	"github.com/sabariramc/goserverbase/v2/utils/testutils"
)

var LoggerTestConfig *testutils.TestConfig
var KafkaTestLogger *log.Logger

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../../.env")
	LoggerTestConfig = testutils.NewConfig()

}
