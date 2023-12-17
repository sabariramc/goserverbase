package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	baseapp "github.com/sabariramc/goserverbase/v4/app"
	"github.com/sabariramc/goserverbase/v4/errors"
	"github.com/sabariramc/goserverbase/v4/log"
)

type HTTPServer struct {
	*baseapp.BaseApp
	handler *gin.Engine
	docMeta APIDocumentation
	log     *log.Logger
	c       *HTTPServerConfig
}

func New(appConfig HTTPServerConfig, logger *log.Logger, errorNotifier errors.ErrorNotifier) *HTTPServer {
	b := baseapp.New(appConfig.ServerConfig, logger, errorNotifier)
	if appConfig.Log.ContentLength <= 0 {
		appConfig.Log.ContentLength = 1024
	}
	h := &HTTPServer{
		BaseApp: b,
		handler: gin.New(),
		docMeta: APIDocumentation{
			Server: make([]DocumentServer, 0),
			Routes: make(APIRoute, 0),
		},
		log: b.GetLogger().NewResourceLogger("HTTPServer"),
		c:   &appConfig,
	}
	ctx := b.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam(appConfig.ServiceName))
	h.SetupRouter(ctx)
	return h
}

func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func (h *HTTPServer) GetAPIDocument() APIDocumentation {
	return h.docMeta
}

func (h *HTTPServer) GetRouter() *gin.Engine {
	return h.handler
}

func (h *HTTPServer) StartServer() {
	h.log.Notice(context.TODO(), fmt.Sprintf("Server starting at %v", h.GetPort()), nil)
	err := http.ListenAndServe(h.GetPort(), h)
	h.log.Emergency(context.Background(), "Server crashed", nil, err)
}

func (h *HTTPServer) StartTLSServer(certFile, keyFile string) {
	h.log.Notice(context.TODO(), fmt.Sprintf("Server starting at %v", h.GetPort()), nil)
	err := http.ListenAndServeTLS(h.GetPort(), certFile, keyFile, h)
	h.log.Emergency(context.Background(), "Server crashed", nil, err)
}

func (h *HTTPServer) GetPort() string {
	return fmt.Sprintf("%v:%v", h.c.Host, h.c.Port)
}

func (h *HTTPServer) AddServerHost(server DocumentServer) {
	h.docMeta.Server = append(h.docMeta.Server, server)
}
