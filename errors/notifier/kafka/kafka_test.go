package kafka_test

import (
	"context"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/v3/errors/notifier/kafka"
	pKafka "github.com/sabariramc/goserverbase/v3/kafka"
	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/sabariramc/goserverbase/v3/log/logwriter"
	"github.com/sabariramc/goserverbase/v3/utils/testutils"
	"gotest.tools/assert"
)

var KafkaTestConfig *testutils.TestConfig
var KafkaTestLogger *log.Logger

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../../../.env")
	KafkaTestConfig = testutils.NewConfig()
	hostParams := log.HostParams{
		Version:     KafkaTestConfig.Logger.Version,
		Host:        KafkaTestConfig.Http.Host,
		ServiceName: KafkaTestConfig.App.ServiceName,
	}
	consoleLogWriter := logwriter.NewConsoleWriter(hostParams)

	lMux := log.NewDefaultLogMux(consoleLogWriter)
	KafkaTestLogger = log.NewLogger(context.TODO(), KafkaTestConfig.Logger, "KafkaTest", lMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(KafkaTestConfig.App.ServiceName))
	return ctx
}

func TestErrorNotification(t *testing.T) {
	p, _ := pKafka.NewProducer(context.TODO(), KafkaTestLogger, KafkaTestConfig.KafkaProducerConfig, KafkaTestConfig.App.ServiceName, KafkaTestConfig.KafkaTestTopic, nil)
	notifier := kafka.New(context.TODO(), KafkaTestLogger, "Test", p)
	ctx := GetCorrelationContext()
	err := notifier.Send4XX(log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{CustomerId: "customer_id_test"}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
	err = notifier.Send4XX(log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{AppUserId: "app_user_id"}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
	err = notifier.Send4XX(log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{Id: "entity_id"}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
}

func TestErrorNotification2(t *testing.T) {
	p := pKafka.NewHTTPProducer(context.TODO(), KafkaTestLogger, KafkaTestConfig.KafkaHTTPProxyURL, KafkaTestConfig.KafkaTestTopic, time.Second)
	notifier := kafka.New(context.TODO(), KafkaTestLogger, "Test", p)
	ctx := GetCorrelationContext()
	err := notifier.Send4XX(log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{CustomerId: "customer_id_test"}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
	err = notifier.Send4XX(log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{AppUserId: "app_user_id"}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
	err = notifier.Send4XX(log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{Id: "entity_id"}), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
	assert.NilError(t, err)
}
