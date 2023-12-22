package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	server  *http.Server
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
	h.BaseApp.AddShutdownHook(h)
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

func (h *HTTPServer) Name(ctx context.Context) string {
	return "HTTPServer"
}

func (h *HTTPServer) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *HTTPServer) BootstrapServer(ctx context.Context) error {
	h.server = &http.Server{Addr: h.GetPort(), Handler: h}
	return h.StartSignalMonitor(ctx)
}

func (h *HTTPServer) StartServer() {
	corr := &log.CorrelationParam{CorrelationId: fmt.Sprintf("%v-HTTP-SERVER", h.c.ServiceName)}
	ctx := log.GetContextWithCorrelation(context.TODO(), corr)
	h.log.Notice(ctx, fmt.Sprintf("Server starting at %v", h.GetPort()), nil)
	err := h.BootstrapServer(ctx)
	if err != nil {
		h.log.Emergency(context.Background(), "Server bootstrap failed", err, nil)
	}
	err = h.server.ListenAndServe()
	time.Sleep(time.Second)
	if err != nil && err != http.ErrServerClosed {
		h.log.Emergency(context.Background(), "Server crashed", err, nil)
	}
}

func (h *HTTPServer) StartTLSServer() {
	corr := &log.CorrelationParam{CorrelationId: fmt.Sprintf("%v-HTTP2-SERVER", h.c.ServiceName)}
	ctx := log.GetContextWithCorrelation(context.TODO(), corr)
	h.log.Notice(ctx, fmt.Sprintf("Server starting at %v", h.GetPort()), nil)
	err := h.BootstrapServer(ctx)
	if err != nil {
		h.log.Emergency(context.Background(), "Server bootstrap failed", err, nil)
	}
	err = h.server.ListenAndServeTLS(h.c.HTTP2Config.PublicKeyPath, h.c.HTTP2Config.PrivateKeyPath)
	time.Sleep(time.Second)
	if err != nil && err != http.ErrServerClosed {
		h.log.Emergency(context.Background(), "Server crashed", err, nil)
	}
}

func (h *HTTPServer) GetPort() string {
	return fmt.Sprintf("%v:%v", h.c.Host, h.c.Port)
}

func (h *HTTPServer) AddServerHost(server DocumentServer) {
	h.docMeta.Server = append(h.docMeta.Server, server)
}
