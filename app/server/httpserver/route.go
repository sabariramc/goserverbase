package httpserver

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sabariramc/goserverbase/v5/errors"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(HttpHeaderContentType, HttpContentTypeJSON)
		w.WriteHeader(http.StatusNotFound)
		body, _ := errors.NewHTTPClientError(http.StatusNotFound, "URL_NOT_FOUND", "Invalid path", nil, map[string]string{
			"path": r.URL.Path,
		}, nil).GetErrorResponse()
		w.Write(body)
	}
}

func MethodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(HttpHeaderContentType, HttpContentTypeJSON)
		w.WriteHeader(http.StatusMethodNotAllowed)
		body, _ := errors.NewHTTPClientError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Invalid method", nil, map[string]string{
			"path":   r.URL.Path,
			"method": r.Method,
		}, nil).GetErrorResponse()
		w.Write(body)
	}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(204)
}

func (h *HTTPServer) SetupRouter(ctx context.Context) {
	h.handler.NoRoute(gin.WrapF(NotFound()))
	h.handler.NoMethod(gin.WrapF(MethodNotAllowed()))
	h.handler.GET("/meta/health", gin.WrapF(HealthCheck))
	h.handler.HandleMethodNotAllowed = true
	h.SetupDocumentation(ctx)
	h.handler.Use(h.SetContextMiddleware(), h.RequestTimerMiddleware(), h.LogRequestResponseMiddleware(), h.HandleExceptionMiddleware())
}

func (h *HTTPServer) SetupDocumentation(ctx context.Context) {
	h.handler.GET("/meta/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler,
		ginSwagger.URL(h.c.DocHost+"/meta/static/swagger.yaml"),
		ginSwagger.DefaultModelsExpandDepth(-1), func(c *ginSwagger.Config) {
			c.Title = h.c.ServiceName
		}))
	h.handler.StaticFS("/meta/static", http.Dir(h.c.SwaggerRootFolder))
}

func (h *HTTPServer) AddMiddleware(middleware ...gin.HandlerFunc) {
	h.handler.Use(middleware...)
}
