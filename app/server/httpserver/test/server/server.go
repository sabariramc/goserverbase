package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sabariramc/goserverbase/v4/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v4/errors"
	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/log/logwriter"
	"github.com/sabariramc/goserverbase/v4/utils/testutils"
)

var ServerTestConfig *testutils.TestConfig
var ServerTestLogger *log.Logger
var ServerTestLMux log.LogMux

func init() {
	testutils.LoadEnv("../../../.env")
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
	*httpserver.HttpServer
	log *log.Logger
}

func (s *server) Func1(w http.ResponseWriter, r *http.Request) {
	id := log.GetCustomerIdentifier(r.Context())
	corr := log.GetCorrelationParam(r.Context())
	s.log.Info(r.Context(), "identity", id)
	s.log.Info(r.Context(), "correlation", corr)
	data := s.GetBody(r)
	s.WriteJsonWithStatusCode(r.Context(), w, 200, map[string]string{"body": string(data)})
}

func (s *server) Func2(c *gin.Context) {
	w := c.Writer
	fmt.Println(c.Param("tenantId"))
	w.WriteHeader(200)
	w.Write([]byte("World"))
}

func (s *server) Func3(w http.ResponseWriter, r *http.Request) {
	s.SetHandlerErrorInContext(r.Context(), errors.NewCustomError("hello.new.custom.error", "display this", map[string]any{"one": "two"}, nil, true))
}

func (s *server) Func4(w http.ResponseWriter, r *http.Request) {
	s.log.Emergency(r.Context(), "random panic at Func4", nil, nil)
}

func (s *server) Func5(w http.ResponseWriter, r *http.Request) {
	s.SetHandlerErrorInContext(r.Context(), errors.NewHTTPClientError(403, "hello.new.custom.error", "display this", map[string]any{"one": "two"}, nil))
}

func NewServer() *server {
	srv := &server{
		HttpServer: httpserver.New(*ServerTestConfig.Http, ServerTestLogger, nil), log: ServerTestLogger,
	}
	r := srv.GetRouter().Group("/service/v1")
	tenant := r.Group("/tenant")
	tenant.GET("", gin.WrapF(srv.Func1))
	tenant.POST("", gin.WrapF(srv.Func1))
	tenant.GET("/:tenantId", srv.Func2)
	errorRoute := r.Group("/error")
	errorRoute.GET("/error1", gin.WrapF(srv.Func3))
	errorRoute.GET("/error2", gin.WrapF(srv.Func4))
	errorRoute.GET("/error3", gin.WrapF(srv.Func5))
	return srv
}
