package kafka_test

import (
	"context"
	"runtime"
	"testing"

	pKafka "github.com/sabariramc/goserverbase/v5/kafka"
	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/log/logwriter"
	"github.com/sabariramc/goserverbase/v5/notifier/kafka"
	"github.com/sabariramc/goserverbase/v5/testutils"
	"gotest.tools/assert"
)

var TestConfig *testutils.TestConfig
var TestLogger log.Log

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../../.env")
	TestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter()
	lMux := log.NewDefaultLogMux(consoleLogWriter)
	TestLogger = log.New(context.TODO(), TestConfig.Logger, "KafkaTest", lMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(TestConfig.App.ServiceName))
	return ctx
}

func TestKafkaNotifier(t *testing.T) {
	p, _ := pKafka.NewProducer(context.TODO(), TestLogger, TestConfig.KafkaProducer, nil)
	notifier := kafka.New(context.TODO(), TestLogger, "Test", TestConfig.KafkaTestTopic, p)
	ctx := GetCorrelationContext()
	custID := "cust_test_id"
	errorCode := "com.testing.error"
	stackTraceBuff := []byte{}
	runtime.Stack(stackTraceBuff, false)
	stackTrace := string(stackTraceBuff)
	errorData := map[string]any{"check": "Testing error"}
	err := notifier.Notify4XX(log.GetContextWithUserIdentifier(ctx, &log.UserIdentifier{UserID: &custID}), errorCode, nil, stackTrace, errorData)
	assert.NilError(t, err)
	custID = "app_user_id"
	err = notifier.Notify4XX(log.GetContextWithUserIdentifier(ctx, &log.UserIdentifier{AppUserID: &custID}), errorCode, nil, stackTrace, errorData)
	assert.NilError(t, err)
	custID = "entity_id"
	err = notifier.Notify4XX(log.GetContextWithUserIdentifier(ctx, &log.UserIdentifier{EntityID: &custID}), errorCode, nil, stackTrace, errorData)
	assert.NilError(t, err)
}
