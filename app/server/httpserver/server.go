// Package httpserver contains basic HTTPServer
package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	baseapp "github.com/sabariramc/goserverbase/v6/app"
	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"github.com/sabariramc/goserverbase/v6/log"
)

// Tracer defines the interface for tracing functionality.
type Tracer interface {
	GetGinMiddleware(serviceName string) gin.HandlerFunc
	span.SpanOp
}

// HTTPServer represents an HTTP server, extends gin.Engine with additional middleware and OpenAPI documentation
// Implements ShutdownHook, HealthCheckHook and StatusCheckHook
type HTTPServer struct {
	*baseapp.BaseApp
	handler         *gin.Engine
	log             log.Log
	c               *Config
	server          *http.Server
	tracer          Tracer
	connectionCount int64
}

// New creates a new instance of HTTPServer.
func New(option ...Option) *HTTPServer {
	config := GetDefaultConfig()
	h := &HTTPServer{
		BaseApp: baseapp.NewWithConfig(config.Config),
		handler: gin.New(),
		log:     config.Log,
		c:       config,
		tracer:  config.Tracer,
	}
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), correlation.NewCorrelationParam(config.ServiceName))
	h.SetupRouter(ctx)
	h.RegisterOnShutdownHook(h)
	h.RegisterStatusCheckHook(h)
	return h
}

// ServeHTTP implements the http.Handler interface.
func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

// GetRouter returns the underlying gin.Engine instance.
func (h *HTTPServer) GetRouter() *gin.Engine {
	return h.handler
}

// Name returns the name of the HTTPServer.
// Implementation of the hook interface defined in the BaseApp
func (h *HTTPServer) Name(ctx context.Context) string {
	return "HTTPServer"
}

// Shutdown gracefully shuts down the HTTP server.
// Implementation for shutdown hook
func (h *HTTPServer) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

// GetPort returns the host and port of the HTTP server.
func (h *HTTPServer) GetPort() string {
	return fmt.Sprintf("%v:%v", h.c.Host, h.c.Port)
}

// GetSpanFromContext retrieves the telemetry span from the given context, if the server is initiated with a tracer
func (h *HTTPServer) GetSpanFromContext(ctx context.Context) (span.Span, bool) {
	if h.tracer != nil {
		return h.tracer.GetSpanFromContext(ctx)
	}
	return nil, false
}
