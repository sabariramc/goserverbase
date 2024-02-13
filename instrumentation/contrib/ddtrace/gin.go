package ddtrace

import (
	"github.com/gin-gonic/gin"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
)

func (t *tracer) GetGinMiddleware(serviceName string) gin.HandlerFunc {
	return ddtrace.Middleware(serviceName)
}
