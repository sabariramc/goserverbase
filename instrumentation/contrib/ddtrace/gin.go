package ddtrace

import (
	"github.com/gin-gonic/gin"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
)

// GetGinMiddleware returns a middleware handler for Gin framework that instruments requests with Datadog tracing.
// It takes the serviceName parameter specifying the service name to be used for tracing.
func (t *tracer) GetGinMiddleware(serviceName string) gin.HandlerFunc {
	return ddtrace.Middleware(serviceName)
}
