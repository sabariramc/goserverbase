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
	ctx := b.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam(appConfig.ServiceName))
	h.SetupRouter(ctx)
	return h
}

func (h *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func (h *HttpServer) GetAPIDocument() APIDocumentation {
	return h.docMeta
}

func (h *HttpServer) GetRouter() *chi.Mux {
	return h.handler
}

func (h *HttpServer) StartServer() {
	h.Log.Notice(context.TODO(), fmt.Sprintf("Server starting at %v", h.GetPort()), nil)
	err := http.ListenAndServe(h.GetPort(), h)
	h.Log.Emergency(context.Background(), "Server crashed", nil, err)
}

func (h *HttpServer) GetPort() string {
	return fmt.Sprintf("%v:%v", h.c.Host, h.c.Port)
}

func (h *HttpServer) AddServerHost(server DocumentServer) {
	h.docMeta.Server = append(h.docMeta.Server, server)
}
