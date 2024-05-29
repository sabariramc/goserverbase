package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/app/server/kafkaclient"
	"github.com/sabariramc/goserverbase/v6/aws"
	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/db/mongo"
	"github.com/sabariramc/goserverbase/v6/errors"
	"github.com/sabariramc/goserverbase/v6/instrumentation"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/testutils"
	"github.com/sabariramc/goserverbase/v6/utils"
	"github.com/sabariramc/goserverbase/v6/utils/retryhttp"
)

var ServerTestConfig *testutils.TestConfig
var ServerTestLogger log.Log
var ServerTestLMux log.Mux

const ServiceName = "KafkaClientTest"

func init() {
	testutils.LoadEnv("../../../.env")

	ServerTestConfig = testutils.NewConfig()
	ServerTestLogger = log.New(log.WithServiceName(ServiceName))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, correlation.NewCorrelationParam(ServiceName))
	return ctx
}

type server struct {
	*kafkaclient.KafkaConsumerServer
	log        log.Log
	pr         *kafka.Producer
	conn       *mongo.Mongo
	coll       *mongo.Collection
	sns        *aws.SNS
	httpClient *retryhttp.HTTPClient
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
		s.sns.Publish(ctx, &ServerTestConfig.AWS.SNS, nil, msg, nil)
	}()
	go func() {
		defer wg.Done()
		res := make(map[string]any)
		s.httpClient.Post(ctx, ServerTestConfig.TestURL2, data, &res, nil)
		s.log.Info(ctx, "http response", res)
	}()
	s.pr.ProduceMessageWithTopic(ctx, ServerTestConfig.KafkaTestTopic2, uuid.NewString(), msg, nil)
	wg.Wait()
	return nil
}

func (s *server) Func2(ctx context.Context, msg *kafka.Message) error {
	return &errors.CustomError{ErrorCode: "hello.new.custom.error", ErrorMessage: "display this", ErrorData: map[string]any{"one": "two"}, Notify: true}
}

func (s *server) Name(ctx context.Context) string {
	return ""
}

func NewServer(t instrumentation.Tracer) *server {
	testutils.SetAWSConfig(t)
	ctx := GetCorrelationContext()
	pr, err := kafka.NewProducer(kafka.WithProducerTracer(t))
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating producer2", err, nil)
	}
	conn, err := mongo.NewWithDefaultOptions(ServerTestLogger, t)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating mongo connection", err, nil)
	}
	srv := &server{
		KafkaConsumerServer: kafkaclient.New(kafkaclient.WithLog(ServerTestLogger), kafkaclient.WithTracer(t)),
		pr:                  pr,
		sns:                 aws.GetDefaultSNSClient(ServerTestLogger),
		httpClient:          retryhttp.New(retryhttp.WithTracer(t)),
		conn:                conn,
		coll:                conn.Database("GOBaseTest").Collection("TestColl"),
	}
	srv.RegisterHooks(conn)
	srv.RegisterHooks(pr)
	srv.log = srv.GetLogger().NewResourceLogger("KafkaTestServer")
	srv.log.Trace(ctx, "config", ServerTestConfig)
	srv.AddHandler(GetCorrelationContext(), ServerTestConfig.KafkaTestTopic, srv.Func1)
	srv.AddHandler(GetCorrelationContext(), ServerTestConfig.KafkaTestTopic2, srv.Func2)
	return srv
}
