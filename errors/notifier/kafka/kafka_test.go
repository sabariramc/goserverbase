package kafka_test

import (
	"context"
	"testing"

	"github.com/sabariramc/goserverbase/v2/errors/notifier/kafka"
	pKafka "github.com/sabariramc/goserverbase/v2/kafka"
	"github.com/sabariramc/goserverbase/v2/log"
	"github.com/sabariramc/goserverbase/v2/log/logwriter"
	"github.com/sabariramc/goserverbase/v2/utils/testutils"
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
	p, _ := pKafka.NewProducer(context.TODO(), KafkaTestLogger, KafkaTestConfig.KafkaProducerConfig, KafkaTestConfig.KafkaTestTopic)
	notifier := kafka.New(context.TODO(), KafkaTestLogger, KafkaTestConfig.KafkaHTTPProxyURL, KafkaTestConfig.KafkaTestTopic, "Test", p)
	notifier.Send4XX(GetCorrelationContext(), "com.testing.error", nil, "testing", map[string]any{"check": "Testing error"})
}
