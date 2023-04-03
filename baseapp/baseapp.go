package baseapp

import (
	"context"
	"net/http"
	"time"

	"github.com/sabariramc/goserverbase/config"
	"github.com/sabariramc/goserverbase/errors"
	"github.com/sabariramc/goserverbase/log"
)

type CustomHandler interface {
	http.Handler
	SetNotFound(http.HandlerFunc)
	SetMethodNotAllowed(http.HandlerFunc)
	RegisterRoute(method, path string, handler http.HandlerFunc)
}

type BaseApp struct {
	handler       CustomHandler
	c             *config.ServerConfig
	lConfig       *log.Config
	log           *log.Logger
	errorNotifier errors.ErrorNotifier
	docMeta       APIDocumentation
}

func New(appConfig config.ServerConfig, handler CustomHandler, loggerConfig log.Config, lMux log.LogMux, errorNotifier errors.ErrorNotifier, auditLogger log.AuditLogWriter) *BaseApp {
	b := &BaseApp{
		c:             &appConfig,
		lConfig:       &loggerConfig,
		handler:       handler,
		errorNotifier: errorNotifier,
		docMeta: APIDocumentation{
			Server: make([]DocumentServer, 0),
			Routes: make(APIRoute, 0),
		},
	}
	ctx := b.GetCorrelationContext(context.Background(), log.GetDefaultCorrelationParams(appConfig.ServiceName))
	b.log = log.NewLogger(ctx, &loggerConfig, loggerConfig.ServiceName, lMux, auditLogger)
	zone, _ := time.Now().Zone()
	b.log.Notice(ctx, "Timezone", zone)
	b.RegisterDefaultRoutes(ctx)
	return b
}

func (b *BaseApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.SetContextMiddleware(b.RequestTimerMiddleware(b.LogRequestResponseMiddleware(b.HandleExceptionMiddleware(b.handler)))).ServeHTTP(w, r)
}

func (b *BaseApp) GetAPIDocument() APIDocumentation {
	return b.docMeta
}

func (b *BaseApp) GetConfig() config.ServerConfig {
	return *b.c
}

func (b *BaseApp) GetLogger() *log.Logger {
	return b.log
}

func (b *BaseApp) SetLogger(l *log.Logger) {
	b.log = l
}

func (b *BaseApp) AddServerHost(server DocumentServer) {
	b.docMeta.Server = append(b.docMeta.Server, server)
}
