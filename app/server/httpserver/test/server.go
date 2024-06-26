package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v6/aws"
	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/db/mongo"
	"github.com/sabariramc/goserverbase/v6/errors"
	"github.com/sabariramc/goserverbase/v6/instrumentation"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/log/logwriter"
	"github.com/sabariramc/goserverbase/v6/testutils"
	"github.com/sabariramc/goserverbase/v6/utils"
	"github.com/sabariramc/goserverbase/v6/utils/retryhttp"
)

var ServerTestConfig *testutils.TestConfig
var ServerTestLogger log.Log
var ServerTestLMux log.Mux

const ServiceName = "BaseTest"

func init() {
	fmt.Println(os.Getwd())
	testutils.LoadEnv("../../../.env")
	ServerTestConfig = testutils.NewConfig()
	consoleLogWriter := logwriter.NewConsoleWriter()
	ServerTestLMux = log.NewDefaultLogMux(consoleLogWriter)
	ServerTestLogger = log.New(log.WithServiceName(ServiceName))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, correlation.NewCorrelationParam(ServiceName))
	return ctx
}

type server struct {
	*httpserver.HTTPServer
	log        log.Log
	pr         *kafka.Producer
	conn       *mongo.Mongo
	coll       *mongo.Collection
	sns        *aws.SNS
	httpClient *retryhttp.HTTPClient
	c          *testutils.TestConfig
}

func (s *server) echo(c *gin.Context) {
	w, r := c.Writer, c.Request
	id := correlation.ExtractUserIdentifier(r.Context())
	corr := correlation.ExtractCorrelationParam(r.Context())
	s.log.Info(r.Context(), "identity", id)
	s.log.Info(r.Context(), "correlation", corr)
	data, _ := s.GetRequestBody(r)
	s.WriteJSONWithStatusCode(r.Context(), w, 200, map[string]any{
		"body":        string(data),
		"headers":     r.Header,
		"queryParams": r.URL.Query(),
		"pathParams":  c.Param("any"),
	})
}

func (s *server) benc(c *gin.Context) {
	c.Status(http.StatusNoContent)
	return
}

func (s *server) testAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := make(map[string]any)
	err := s.LoadRequestJSONBody(r, &data)
	if err != nil {
		s.WriteErrorResponse(ctx, w, errors.HTTPError{StatusCode: 400, CustomError: &errors.CustomError{ErrorCode: "invalidJsonBody", ErrorMessage: "error marshalling json body"}}, "")
		return
	}
	s.coll.InsertOne(ctx, data)
	msg := utils.NewMessage("testFlight", "test")
	msg.AddPayload("content", data)
	wg := &sync.WaitGroup{}
	s.run(wg, func() { s.sns.Publish(ctx, &ServerTestConfig.AWS.SNS, nil, msg, nil) })
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
		s.WriteErrorResponse(ctx, w, &errors.HTTPError{StatusCode: 400, CustomError: &errors.CustomError{ErrorCode: "invalidJsonBody", ErrorMessage: "error marshalling json body"}}, "")
		return
	}
	msg := utils.NewMessage("testFlight", "test")
	msg.AddPayload("content", data)
	s.pr.ProduceMessageWithTopic(ctx, ServerTestConfig.KafkaTestTopic2, uuid.NewString(), msg, nil)
	w.WriteHeader(204)
}

func (s *server) error500(w http.ResponseWriter, r *http.Request) {
	s.WriteErrorResponse(r.Context(), w, &errors.CustomError{ErrorCode: "hello.new.custom.error", ErrorMessage: "display this", ErrorData: map[string]any{"one": "two"}, Notify: true}, "")

}

func (s *server) errorWithPanic(w http.ResponseWriter, r *http.Request) {
	s.log.Emergency(r.Context(), "random panic at Func4", errors.HTTPError{StatusCode: 503, CustomError: &errors.CustomError{ErrorCode: "hello.new.custom.error", ErrorMessage: "display this", ErrorData: map[string]any{"one": "two"}}})
}

func (s *server) panic(w http.ResponseWriter, r *http.Request) {
	panic("fasdfasfsadf")
}

func (s *server) errorUnauthorized(w http.ResponseWriter, r *http.Request) {
	s.WriteErrorResponse(r.Context(), w, errors.HTTPError{StatusCode: 403, CustomError: &errors.CustomError{ErrorCode: "hello.new.custom.error", ErrorMessage: "display this", ErrorData: map[string]any{"one": "two"}}}, "")
}

func (s *server) printHttpVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		s.log.Debug(r.Context(), "http proto", r.Proto)
		c.Next()
	}
}

func New(t instrumentation.Tracer) *server {
	testutils.SetAWSConfig(t)
	ctx := GetCorrelationContext()
	pr, err := kafka.NewProducer(kafka.WithProducerTracer(t))
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating producer1", err, nil)
	}
	conn, err := mongo.NewWithDefaultOptions(ServerTestLogger, t)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating mongo connection", err, nil)
	}
	srv := &server{
		HTTPServer: httpserver.New(httpserver.WithTracer(t)), log: ServerTestLogger,
		pr:         pr,
		sns:        aws.GetDefaultSNSClient(ServerTestLogger),
		httpClient: retryhttp.New(retryhttp.WithTracer(t)),
		conn:       conn,
		coll:       conn.Database("GOBaseTest").Collection("TestColl"),
		c:          ServerTestConfig,
	}
	srv.AddMiddleware(srv.printHttpVersion())
	srv.RegisterHooks(conn)
	srv.RegisterHooks(pr)
	srv.GetRouter().GET("/meta/bench", srv.benc)
	service := srv.GetRouter().Group("/service")
	echo := service.Group("/echo")
	echo.Any("", srv.echo)
	echo.GET("/*any", srv.echo)
	integrationTest := service.Group("/test")
	integrationTest.POST("/all", gin.WrapF(srv.testAll))
	integrationTest.POST("/kafka", gin.WrapF(srv.testKafka))
	errorRoute := service.Group("/error")
	errorRoute.GET("/error500", gin.WrapF(srv.error500))
	errorRoute.GET("/errorWithPanic", gin.WrapF(srv.errorWithPanic))
	errorRoute.GET("/errorUnauthorized", gin.WrapF(srv.errorUnauthorized))
	errorRoute.GET("/panic", gin.WrapF(srv.panic))
	return srv
}
