package kafka_test

import (
	"context"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/v4/errors/notifier/kafka"
	pKafka "github.com/sabariramc/goserverbase/v4/kafka"
	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/log/logwriter"
	"github.com/sabariramc/goserverbase/v4/utils/testutils"
	"gotest.tools/assert"
)

var TestConfig *testutils.TestConfig
var TestLogger *log.Logger

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../../../.env")
	TestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter(TestConfig.Logger.HostParams)
	lMux := log.NewDefaultLogMux(consoleLogWriter)
	TestLogger = log.NewLogger(context.TODO(), TestConfig.Logger, "KafkaTest", lMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(TestConfig.App.ServiceName))
	return ctx
}

func TestErrorNotification(t *testing.T) {
	p, _ := pKafka.NewProducer(context.TODO(), TestLogger, TestConfig.KafkaProducer, TestConfig.App.ServiceName, TestConfig.KafkaTestTopic)
	notifier := kafka.New(context.TODO(), TestLogger, "Test", p)
	ctx := GetCorrelationContext()
	err := notifier.Send4XX(log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{CustomerId: "customer_id_test"}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
	err = notifier.Send4XX(log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{AppUserId: "app_user_id"}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
	err = notifier.Send4XX(log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{Id: "entity_id"}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
}

func TestErrorNotification2(t *testing.T) {
	p := pKafka.NewHTTPProducer(context.TODO(), TestLogger, TestConfig.KafkaHTTPProxyURL, TestConfig.KafkaTestTopic, time.Second)
	notifier := kafka.New(context.TODO(), TestLogger, "Test", p)
	ctx := GetCorrelationContext()
	err := notifier.Send4XX(log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{CustomerId: "customer_id_test"}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
	err = notifier.Send4XX(log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{AppUserId: "app_user_id"}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
	err = notifier.Send4XX(log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{Id: "entity_id"}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
}
