package opentelemetry

import (
	"github.com/gin-gonic/gin"
	gintrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
)

func (t *tracer) GetGinMiddleware(serviceName string) gin.HandlerFunc {
	return gintrace.Middleware(serviceName)
}
