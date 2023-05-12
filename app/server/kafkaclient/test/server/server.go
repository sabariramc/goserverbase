package server

import (
	"context"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sabariramc/goserverbase/v2/app/server/kafkaclient"
	"github.com/sabariramc/goserverbase/v2/log"
	"github.com/sabariramc/goserverbase/v2/log/logwriter"
	"github.com/sabariramc/goserverbase/v2/utils/testutils"
)

var ServerTestConfig *testutils.TestConfig
var ServerTestLogger *log.Logger
var ServerTestLMux log.LogMux

func init() {
	testutils.Initialize()
	ServerTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter(log.HostParams{
		Version:     ServerTestConfig.Logger.Version,
		Host:        ServerTestConfig.Http.Host,
		ServiceName: ServerTestConfig.App.ServiceName,
	})
	ServerTestLMux = log.NewDefaultLogMux(consoleLogWriter)
	ServerTestLogger = log.NewLogger(context.TODO(), ServerTestConfig.Logger, "BaseTest", ServerTestLMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParams(ServerTestConfig.App.ServiceName))
	return ctx
}

type server struct {
	*kafkaclient.KafkaClient
	log *log.Logger
}

func (s *server) Func1(context.Context, *ckafka.Message) error {
	return nil
}

func NewServer() *server {
	srv := &server{
		KafkaClient: kafkaclient.New(*ServerTestConfig.Kafka, *ServerTestConfig.Logger, ServerTestLMux, nil, nil),
	}
	srv.log = srv.GetLogger()
	srv.AddHandler(ServerTestConfig.KafkaTestTopic, srv.Func1)
	return srv
}
