package otel

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// GetGinMiddleware returns a Gin middleware handler for OpenTelemetry instrumentation.
// This middleware will trace incoming HTTP requests to the specified service name.
func (t *tracerManager) GetGinMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}
