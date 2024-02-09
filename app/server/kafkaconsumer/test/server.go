package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v5/app/server/kafkaconsumer"
	"github.com/sabariramc/goserverbase/v5/aws"
	"github.com/sabariramc/goserverbase/v5/db/mongo"
	"github.com/sabariramc/goserverbase/v5/errors"
	"github.com/sabariramc/goserverbase/v5/instrumentation"
	"github.com/sabariramc/goserverbase/v5/kafka"
	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/log/logwriter"
	"github.com/sabariramc/goserverbase/v5/testutils"
	"github.com/sabariramc/goserverbase/v5/utils"
	"github.com/sabariramc/goserverbase/v5/utils/httputil"
)

var ServerTestConfig *testutils.TestConfig
var ServerTestLogger log.Log
var ServerTestLMux log.LogMux

func init() {
	testutils.LoadEnv("../../../.env")
	testutils.Initialize()
	ServerTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter()
	ServerTestLMux = log.NewDefaultLogMux(consoleLogWriter)
	ServerTestLogger = log.NewLogger(context.TODO(), ServerTestConfig.Logger, "BaseTest", ServerTestLMux, nil)
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), log.ContextKeyCorrelation, log.GetDefaultCorrelationParam(ServerTestConfig.App.ServiceName))
	return ctx
}

type server struct {
	*kafkaconsumer.KafkaConsumerServer
	log        log.Log
	pr         *kafka.Producer
	conn       *mongo.Mongo
	coll       *mongo.Collection
	sns        *aws.SNS
	httpClient *httputil.HTTPClient
}

func (s *server) Func1(ctx context.Context, event *kafka.Message) error {
	data := make(map[string]any)
	err := event.LoadBody(&data)
	if err != nil {
		return fmt.Errorf("server.Func1: error loading body: %w", err)
	}
	s.coll.InsertOne(ctx, data)
	msg := utils.NewMessage("testFlight", "test")
	msg.AddPayload("content", data)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		s.sns.Publish(ctx, &ServerTestConfig.AWS.SNS_ARN, nil, msg, nil)
	}()
	go func() {
		defer wg.Done()
		res := make(map[string]any)
		s.httpClient.Post(ctx, ServerTestConfig.TestURL2, data, &res, nil)
		s.log.Info(ctx, "http response", res)
	}()
	s.pr.ProduceMessageWithTopic(ctx, ServerTestConfig.KafkaTestTopic2, uuid.NewString(), msg, nil)
	// s.pr.Flush(ctx)
	wg.Wait()
	return nil
}

func (s *server) Func2(ctx context.Context, msg *kafka.Message) error {
	return errors.NewCustomError("hello.new.custom.error", "display this", map[string]any{"one": "two"}, nil, true, nil)
}

func (s *server) Name(ctx context.Context) string {
	return ""
}

func NewServer(t instrumentation.Tracer) *server {
	ctx := GetCorrelationContext()
	pr, err := kafka.NewProducer(ctx, ServerTestLogger, ServerTestConfig.KafkaProducer)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating producer2", err, nil)
	}
	conn, err := mongo.New(ctx, ServerTestLogger, *ServerTestConfig.Mongo, t)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating mongo connection", err, nil)
	}
	srv := &server{
		KafkaConsumerServer: kafkaconsumer.New(*ServerTestConfig.Kafka, ServerTestLogger, nil),
		pr:                  pr,
		sns:                 aws.GetDefaultSNSClient(ServerTestLogger),
		httpClient:          httputil.NewDefaultHTTPClient(ServerTestLogger, t),
		conn:                conn,
		coll:                conn.Database("GOBaseTest").Collection("TestColl"),
	}
	srv.log = srv.GetLogger().NewResourceLogger("KafkaTestServer")
	srv.log.Trace(ctx, "config", ServerTestConfig)
	srv.AddHandler(GetCorrelationContext(), ServerTestConfig.KafkaTestTopic, srv.Func1)
	srv.AddHandler(GetCorrelationContext(), ServerTestConfig.KafkaTestTopic2, srv.Func2)
	return srv
}
