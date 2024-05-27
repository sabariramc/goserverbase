package httpserver

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sabariramc/goserverbase/v6/errors"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NotFound returns a handler function for responding with a 404 Not Found status.
func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(HttpHeaderContentType, HttpContentTypeJSON)
		w.WriteHeader(http.StatusNotFound)
		body, _ := errors.HTTPError{StatusCode: http.StatusNotFound, CustomError: &errors.CustomError{ErrorCode: "URL_NOT_FOUND", ErrorMessage: "Invalid path", ErrorDescription: map[string]string{
			"path": r.URL.Path,
		}}}.GetErrorResponse()
		w.Write(body)
	}
}

// MethodNotAllowed returns a handler function for responding with a 405 Method Not Allowed status.
func MethodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(HttpHeaderContentType, HttpContentTypeJSON)
		w.WriteHeader(http.StatusMethodNotAllowed)
		body, _ := errors.HTTPError{StatusCode: http.StatusMethodNotAllowed, CustomError: &errors.CustomError{ErrorCode: "METHOD_NOT_ALLOWED", ErrorMessage: "Invalid method", ErrorDescription: map[string]string{
			"path":   r.URL.Path,
			"method": r.Method,
		}}}.GetErrorResponse()
		w.Write(body)
	}
}

// SetupRouter configures routes and middleware for the HTTPServer.
func (h *HTTPServer) SetupRouter(ctx context.Context) {
	h.handler.NoRoute(gin.WrapF(NotFound()))
	h.handler.NoMethod(gin.WrapF(MethodNotAllowed()))
	h.handler.GET("/meta/health", gin.WrapF(h.HealthCheck))
	h.handler.GET("/meta/status", gin.WrapF(h.Status))
	h.handler.HandleMethodNotAllowed = true
	h.SetupDocumentation(ctx)
	if h.tracer != nil {
		h.handler.Use(h.tracer.GetGinMiddleware(h.c.ServiceName))
	}
	h.handler.Use(h.SetCorrelationMiddleware(), h.RequestTimerMiddleware(), h.LogRequestResponseMiddleware(), h.PanicHandleMiddleware())
}

// SetupDocumentation configures routes for serving OpenAPI documentation.
// By default documentation is served in <<host>>/meta/docs/index.html
// The local root folder for the documentation can be configured with config.SwaggerRootFolder, the root folder should contain swagger.yaml
func (h *HTTPServer) SetupDocumentation(ctx context.Context) {
	h.handler.GET("/meta/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler,
		ginSwagger.URL(h.c.DocHost+"/meta/static/swagger.yaml"),
		ginSwagger.DefaultModelsExpandDepth(-1), func(c *ginSwagger.Config) {
			c.Title = h.c.ServiceName
		}))
	h.handler.StaticFS("/meta/static", http.Dir(h.c.RootFolder))
}

// AddMiddleware adds custom middleware to the HTTPServer.
func (h *HTTPServer) AddMiddleware(middleware ...gin.HandlerFunc) {
	h.handler.Use(middleware...)
}
