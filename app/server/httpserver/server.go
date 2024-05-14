package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	baseapp "github.com/sabariramc/goserverbase/v5/app"
	"github.com/sabariramc/goserverbase/v5/instrumentation/span"
	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/notifier"
)

type Tracer interface {
	GetGinMiddleware(serviceName string) gin.HandlerFunc
	span.SpanOp
}

type HTTPServer struct {
	*baseapp.BaseApp
	handler         *gin.Engine
	log             log.Log
	c               *HTTPServerConfig
	server          *http.Server
	tracer          Tracer
	connectionCount int64
}

func New(appConfig HTTPServerConfig, logger log.Log, tr Tracer, errorNotifier notifier.Notifier) *HTTPServer {
	b := baseapp.New(appConfig.ServerConfig, logger, errorNotifier)
	h := &HTTPServer{
		BaseApp: b,
		handler: gin.New(),
		log:     b.GetLogger().NewResourceLogger("HTTPServer"),
		c:       &appConfig,
		tracer:  tr,
	}
	ctx := log.GetContextWithCorrelationParam(context.Background(), log.GetDefaultCorrelationParam(appConfig.ServiceName))
	h.SetupRouter(ctx)
	h.RegisterOnShutdownHook(h)
	h.RegisterStatusCheckHook(h)
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

func (h *HTTPServer) GetSpanFromContext(ctx context.Context) (span.Span, bool) {
	if h.tracer != nil {
		return h.tracer.GetSpanFromContext(ctx)
	}
	return nil, false
}
