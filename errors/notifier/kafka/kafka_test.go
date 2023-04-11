package kafka_test

import (
	"context"
	"testing"

	"github.com/sabariramc/goserverbase/errors/notifier/kafka"
	pKafka "github.com/sabariramc/goserverbase/kafka"
	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/log/logwriter"
	"github.com/sabariramc/goserverbase/utils/testutils"
)

var KafkaTestConfig *testutils.TestConfig
var KafkaTestLogger *log.Logger

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../../../.env")
	KafkaTestConfig = testutils.NewConfig()
	hostParams := log.HostParams{
		Version:     KafkaTestConfig.Logger.Version,
		Host:        KafkaTestConfig.App.Host,
		ServiceName: KafkaTestConfig.App.ServiceName,
	}
	consoleLogWriter := logwriter.NewConsoleWriter(hostParams)

	lMux := log.NewDefaultLogMux(consoleLogWriter)
	KafkaTestLogger = log.NewLogger(context.TODO(), KafkaTestConfig.Logger, "KafkaTest", lMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParams(KafkaTestConfig.App.ServiceName))
	return ctx
}

func TestErrorNotification(t *testing.T) {
	p, _ := pKafka.NewProducer(context.TODO(), KafkaTestLogger, KafkaTestConfig.KafkaProducerConfig, KafkaTestConfig.KafkaTestTopic)
	notifier := kafka.New(context.TODO(), KafkaTestLogger, KafkaTestConfig.KafkaHTTPProxyURL, KafkaTestConfig.KafkaTestTopic, "Test", p)
	notifier.Send4XX(GetCorrelationContext(), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
}
