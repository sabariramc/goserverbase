package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	baseapp "github.com/sabariramc/goserverbase/v2/app"
	"github.com/sabariramc/goserverbase/v2/errors"
	"github.com/sabariramc/goserverbase/v2/log"
)

type HttpServer struct {
	*baseapp.BaseApp
	handler *chi.Mux
	docMeta APIDocumentation
	Log     *log.Logger
	c       *HttpServerConfig
}

func New(appConfig HttpServerConfig, loggerConfig log.Config, lMux log.LogMux, errorNotifier errors.ErrorNotifier, auditLogger log.AuditLogWriter) *HttpServer {
	b := baseapp.New(*appConfig.ServerConfig, loggerConfig, lMux, errorNotifier, auditLogger)
	if appConfig.Log.ContentLength <= 0 {
		appConfig.Log.ContentLength = 1024
	}
	h := &HttpServer{
		BaseApp: b,
		handler: chi.NewRouter(),
		docMeta: APIDocumentation{
			Server: make([]DocumentServer, 0),
			Routes: make(APIRoute, 0),
		},
		Log: b.GetLogger(),
		c:   &appConfig,
	}
	ctx := b.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParams(appConfig.ServiceName))
	h.SetupRouter(ctx)
	return h
}

func (b *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.handler.ServeHTTP(w, r)
}

func (b *HttpServer) GetAPIDocument() APIDocumentation {
	return b.docMeta
}

func (b *HttpServer) GetRouter() *chi.Mux {
	return b.handler
}

func (b *HttpServer) StartServer() {
	b.Log.Notice(context.TODO(), fmt.Sprintf("Server starting at %v", b.GetPort()), nil)
	err := http.ListenAndServe(b.GetPort(), b)
	b.Log.Emergency(context.Background(), "Server crashed", nil, err)
}

func (b *HttpServer) GetPort() string {
	return fmt.Sprintf("%v:%v", b.c.Host, b.c.Port)
}

func (b *HttpServer) AddServerHost(server DocumentServer) {
	b.docMeta.Server = append(b.docMeta.Server, server)
}
