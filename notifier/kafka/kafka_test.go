package kafka_test

import (
	"context"
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

func TestErrorNotification(t *testing.T) {
	p, _ := pKafka.NewProducer(context.TODO(), TestLogger, TestConfig.KafkaProducer, nil)
	notifier := kafka.New(context.TODO(), TestLogger, "Test", TestConfig.KafkaTestTopic, p)
	ctx := GetCorrelationContext()
	custId := "cust_test_id"
	err := notifier.Notify4XX(log.GetContextWithCustomerID(ctx, &log.CustomerIdentifier{UserID: &custId}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
	custId = "app_user_id"
	err = notifier.Notify4XX(log.GetContextWithCustomerID(ctx, &log.CustomerIdentifier{AppUserID: &custId}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
	custId = "entity_id"
	err = notifier.Notify4XX(log.GetContextWithCustomerID(ctx, &log.CustomerIdentifier{EntityID: &custId}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
}
