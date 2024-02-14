package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	baseapp "github.com/sabariramc/goserverbase/v5/app"
	"github.com/sabariramc/goserverbase/v5/errors"
	"github.com/sabariramc/goserverbase/v5/instrumentation/span"
	"github.com/sabariramc/goserverbase/v5/log"
)

type Tracer interface {
	GetGinMiddleware(serviceName string) gin.HandlerFunc
	span.SpanOp
}

type HTTPServer struct {
	*baseapp.BaseApp
	handler *gin.Engine
	log     log.Log
	c       *HTTPServerConfig
	server  *http.Server
	tracer  Tracer
}

func New(appConfig HTTPServerConfig, logger log.Log, tr Tracer, errorNotifier errors.ErrorNotifier) *HTTPServer {
	b := baseapp.New(appConfig.ServerConfig, logger, errorNotifier)
	h := &HTTPServer{
		BaseApp: b,
		handler: gin.New(),
		log:     b.GetLogger().NewResourceLogger("HTTPServer"),
		c:       &appConfig,
		tracer:  tr,
	}
	ctx := b.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam(appConfig.ServiceName))
	h.SetupRouter(ctx)
	h.RegisterOnShutdownHook(h)
	return h
}

func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
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

func (h *HTTPServer) GetPort() string {
	return fmt.Sprintf("%v:%v", h.c.Host, h.c.Port)
}
