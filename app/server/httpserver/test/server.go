package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v5/app/server/httpserver"
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
	fmt.Println(os.Getwd())
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
	*httpserver.HTTPServer
	log        log.Log
	pr         *kafka.Producer
	conn       *mongo.Mongo
	coll       *mongo.Collection
	sns        *aws.SNS
	httpClient *httputil.HTTPClient
	c          *testutils.TestConfig
}

func (s *server) Func1(w http.ResponseWriter, r *http.Request) {
	id := log.GetCustomerIdentifier(r.Context())
	corr := log.GetCorrelationParam(r.Context())
	s.log.Info(r.Context(), "identity", id)
	s.log.Info(r.Context(), "correlation", corr)
	data, _ := s.GetRequestBody(r)
	s.WriteJSONWithStatusCode(r.Context(), w, 200, map[string]string{"body": string(data)})
}

func (s *server) Func2(c *gin.Context) {
	w := c.Writer
	fmt.Println(c.Param("tenantId"))
	w.WriteHeader(200)
	w.Write([]byte("World"))
}

func (s *server) benc(c *gin.Context) {
	return
}

func (s *server) testRequest(c *gin.Context) {
	body, _ := s.GetCacheRequestBody(c.Request)
	s.log.Notice(c.Request.Context(), "request body", string(body))
	w := c.Writer
	w.WriteHeader(200)
	w.Write([]byte(uuid.New().String()))
}

func (s *server) testAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := make(map[string]any)
	err := s.LoadRequestJSONBody(r, &data)
	if err != nil {
		s.SetErrorInContext(ctx, errors.NewHTTPClientError(400, "invalidJsonBody", "error marshalling json body", nil, nil, err))
	}
	s.coll.InsertOne(ctx, data)
	msg := utils.NewMessage("testFlight", "test")
	msg.AddPayload("content", data)
	wg := &sync.WaitGroup{}
	s.run(wg, func() { s.sns.Publish(ctx, &ServerTestConfig.AWS.SNS_ARN, nil, msg, nil) })
	s.run(wg, func() {
		res := make(map[string]any)
		s.httpClient.Post(ctx, ServerTestConfig.TestURL1+"/service/v1/echo/12/2", data, &res, nil)
		s.log.Info(ctx, "http response", res)
	})
	s.run(wg, func() {
		res := make(map[string]any)
		s.httpClient.Post(ctx, ServerTestConfig.TestURL1+"/service/v1/write", data, &res, nil)
		s.log.Info(ctx, "http response", res)
	})
	s.run(wg, func() {
		res := make(map[string]any)
		s.httpClient.Post(ctx, ServerTestConfig.TestURL2, data, &res, nil)
		s.log.Info(ctx, "http response", res)
	})
	s.pr.ProduceMessageWithTopic(ctx, ServerTestConfig.KafkaTestTopic, uuid.NewString(), msg, nil)
	wg.Wait()
	w.WriteHeader(204)
}

func (s *server) run(wg *sync.WaitGroup, fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}

func (s *server) testKafka(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := make(map[string]any)
	err := s.LoadRequestJSONBody(r, &data)
	if err != nil {
		s.SetErrorInContext(ctx, errors.NewHTTPClientError(400, "invalidJsonBody", "error marshalling json body", nil, nil, err))
	}
	msg := utils.NewMessage("testFlight", "test")
	msg.AddPayload("content", data)
	s.pr.ProduceMessageWithTopic(ctx, ServerTestConfig.KafkaTestTopic2, uuid.NewString(), msg, nil)
	w.WriteHeader(204)
}

func (s *server) Func3(w http.ResponseWriter, r *http.Request) {
	s.SetErrorInContext(r.Context(), errors.NewCustomError("hello.new.custom.error", "display this", map[string]any{"one": "two"}, nil, true, nil))
}

func (s *server) Func4(w http.ResponseWriter, r *http.Request) {
	s.log.Emergency(r.Context(), "random panic at Func4", nil, nil)
}

func (s *server) Func5(w http.ResponseWriter, r *http.Request) {
	s.SetErrorInContext(r.Context(), errors.NewHTTPClientError(403, "hello.new.custom.error", "display this", map[string]any{"one": "two"}, nil, nil))
}

func (s *server) Name(ctx context.Context) string {
	return s.c.HTTP.ServiceName
}

func (s *server) Shutdown(ctx context.Context) error {
	s.pr.Close(ctx)
	return s.conn.Disconnect(ctx)
}

func (s *server) printHttpVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		s.log.Debug(r.Context(), "http proto", r.Proto)
		c.Next()
	}
}

func NewServer(t instrumentation.Tracer) *server {
	ctx := GetCorrelationContext()
	pr, err := kafka.NewProducer(ctx, ServerTestLogger, ServerTestConfig.KafkaProducer)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating producer1", err, nil)
	}
	conn, err := mongo.New(ctx, ServerTestLogger, *ServerTestConfig.Mongo, t)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating mongo connection", err, nil)
	}
	srv := &server{
		HTTPServer: httpserver.New(*ServerTestConfig.HTTP, ServerTestLogger, nil), log: ServerTestLogger,
		pr:         pr,
		sns:        aws.GetDefaultSNSClient(ServerTestLogger),
		httpClient: httputil.NewDefaultHTTPClient(ServerTestLogger),
		conn:       conn,
		coll:       conn.Database("GOBaseTest").Collection("TestColl"),
		c:          ServerTestConfig,
	}
	srv.AddMiddleware(srv.printHttpVersion())
	srv.RegisterOnShutdown(srv)
	r := srv.GetRouter().Group("/service/v1")
	r.POST("/benc", srv.benc)
	tenant := r.Group("/tenant")
	tenant.GET("", gin.WrapF(srv.Func1))
	tenant.POST("", gin.WrapF(srv.Func1))
	tenant.GET("/:tenantId", srv.Func2)
	resource := r.Group("/test")
	resource.POST("/all", gin.WrapF(srv.testAll))
	resource.POST("/kafka", gin.WrapF(srv.testKafka))
	resource.POST("/req", srv.testRequest)
	resource.GET("/req", srv.testRequest)
	errorRoute := r.Group("/error")
	errorRoute.GET("/error1", gin.WrapF(srv.Func3))
	errorRoute.GET("/error2", gin.WrapF(srv.Func4))
	errorRoute.GET("/error3", gin.WrapF(srv.Func5))
	return srv
}
