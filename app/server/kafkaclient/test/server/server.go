package server

import (
	"context"

	"github.com/sabariramc/goserverbase/v3/app/server/kafkaclient"
	"github.com/sabariramc/goserverbase/v3/kafka"
	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/sabariramc/goserverbase/v3/log/logwriter"
	"github.com/sabariramc/goserverbase/v3/utils/testutils"
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
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(ServerTestConfig.App.ServiceName))
	return ctx
}

type server struct {
	*kafkaclient.KafkaClient
	log *log.Logger
}

func (s *server) Func1(ctx context.Context, msg *kafka.Message) error {
	s.log.Info(ctx, "Kafka message", msg.GetMeta())
	s.log.Info(ctx, "Kafka body", msg.GetBody())
	return nil
}

func NewServer() *server {
	srv := &server{
		KafkaClient: kafkaclient.New(*ServerTestConfig.Kafka, *ServerTestConfig.Logger, ServerTestLMux, nil, nil),
	}
	srv.log = srv.GetLogger()
	srv.AddHandler(GetCorrelationContext(), ServerTestConfig.KafkaTestTopic, srv.Func1)
	return srv
}
