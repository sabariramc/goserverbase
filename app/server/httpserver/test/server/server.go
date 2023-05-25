package server

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sabariramc/goserverbase/v3/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v3/errors"
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
	*httpserver.HttpServer
}

func (s *server) Func1(w http.ResponseWriter, r *http.Request) {
	id := log.GetCustomerIdentifier(r.Context())
	corr := log.GetCorrelationParam(r.Context())
	s.Log.Info(r.Context(), "identity", id)
	s.Log.Info(r.Context(), "correlation", corr)
	data, _ := io.ReadAll(r.Body)
	s.WriteJsonWithStatusCode(r.Context(), w, 200, map[string]string{"body": string(data)})
}

func (s *server) Func2(w http.ResponseWriter, r *http.Request) {
	fmt.Println(chi.URLParam(r, "tenantId"))
	w.WriteHeader(200)
	w.Write([]byte("World"))
}

func (s *server) Func3(w http.ResponseWriter, r *http.Request) {
	s.SetHandlerErrorInContext(r.Context(), errors.NewCustomError("hello.new.custom.error", "display this", map[string]any{"one": "two"}, nil, true))
}

func (s *server) Func4(w http.ResponseWriter, r *http.Request) {
	s.Log.Emergency(r.Context(), "random panic at Func4", nil, nil)
}

func (s *server) Func5(w http.ResponseWriter, r *http.Request) {
	s.SetHandlerErrorInContext(r.Context(), errors.NewHTTPClientError(403, "hello.new.custom.error", "display this", map[string]any{"one": "two"}, nil))
}

func NewServer() *server {
	srv := &server{
		HttpServer: httpserver.New(*ServerTestConfig.Http, *ServerTestConfig.Logger, ServerTestLMux, nil, nil),
	}
	r := chi.NewRouter()
	r.Route(
		"/tenant", func(r chi.Router) {
			r.Get("/", srv.Func1)
			r.Post("/", srv.Func1)

			r.Route("/{tenantId}", func(r chi.Router) {
				r.Get("/", srv.Func2)
			})
		},
	)
	r.Route(
		"/error", func(r chi.Router) {
			r.Get("/error1", srv.Func3)
			r.Get("/error2", srv.Func4)
			r.Get("/error3", srv.Func5)
		},
	)
	srv.GetRouter().Mount("/service/v1", r)
	return srv
}
