package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sabariramc/goserverbase/v5/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v5/db/mongo"
	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/log/logwriter"
	"github.com/sabariramc/goserverbase/v5/testutils"
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
	log  log.Log
	conn *mongo.Mongo
	coll *mongo.Collection
	c    *testutils.TestConfig
}

func (s *server) Post(w http.ResponseWriter, r *http.Request) {
	id := log.GetCustomerIdentifier(r.Context())
	corr := log.GetCorrelationParam(r.Context())
	s.log.Info(r.Context(), "identity", id)
	s.log.Info(r.Context(), "correlation", corr)
	data, _ := s.GetRequestBody(r)
	s.WriteJSONWithStatusCode(r.Context(), w, 200, map[string]string{"body": string(data)})
}

func (s *server) Get(c *gin.Context) {
	w := c.Writer
	fmt.Println(c.Param("tenantId"))
	w.WriteHeader(200)
	w.Write([]byte("World"))
}

func (s *server) Name(ctx context.Context) string {
	return s.c.HTTP.ServiceName
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.conn.Disconnect(ctx)
}

func NewServer() *server {
	ctx := GetCorrelationContext()
	conn, err := mongo.New(ctx, ServerTestLogger, *ServerTestConfig.Mongo)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating mongo connection", err, nil)
	}
	srv := &server{
		HTTPServer: httpserver.New(*ServerTestConfig.HTTP, ServerTestLogger, nil), log: ServerTestLogger,
		conn: conn,
		coll: conn.Database("GOBaseTest").Collection("TestColl"),
		c:    ServerTestConfig,
	}
	srv.RegisterOnShutdown(srv)
	r := srv.GetRouter().Group("/vault/v1")
	r.GET("", srv.Get)
	r.POST("", gin.WrapF(srv.Post))
	return srv
}
