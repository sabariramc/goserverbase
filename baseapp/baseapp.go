package baseapp

import (
	"context"
	"net/http"
	"time"

	"github.com/sabariramc/goserverbase/config"
	"github.com/sabariramc/goserverbase/errors"
	"github.com/sabariramc/goserverbase/log"

	"github.com/julienschmidt/httprouter"
)

type ServerConfig struct {
	*config.ServerConfig
	LoggerConfig *log.Config
}

type BaseApp struct {
	router        *httprouter.Router
	c             *ServerConfig
	log           *log.Logger
	errorNotifier errors.ErrorNotifier
	docMeta       APIDocumentation
}

func NewBaseApp(c ServerConfig, lMux log.LogMux, errorNotifier errors.ErrorNotifier, auditLogger log.AuditLogWriter) *BaseApp {
	b := &BaseApp{
		c:             &c,
		router:        httprouter.New(),
		errorNotifier: errorNotifier,
		docMeta: APIDocumentation{
			Server: make([]DocumentServer, 0),
			Routes: make(APIRoute, 0),
		},
	}
	ctx := b.GetCorrelationContext(context.Background(), log.GetDefaultCorrelationParams(c.ServiceName))
	b.log = log.NewLogger(ctx, c.LoggerConfig, c.LoggerConfig.ServiceName, lMux, auditLogger)
	zone, _ := time.Now().Zone()
	b.log.Notice(ctx, "Server Timezone", zone)
	b.RegisterDefaultRoutes(ctx)
	return b
}

func (b *BaseApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(b.GetCorrelationContext(r.Context(), b.GetHttpCorrelationParams(r)))
	st := time.Now()
	b.router.ServeHTTP(w, r)
	b.log.Info(r.Context(), "Request processing time in ms", time.Since(st).Milliseconds())
}

func (b *BaseApp) GetRouter() *httprouter.Router {
	return b.router
}

func (b *BaseApp) GetAPIDocument() APIDocumentation {
	return b.docMeta
}

func (b *BaseApp) SetRouter(router *httprouter.Router) {
	b.router = router
	b.docMeta = APIDocumentation{
		Server: make([]DocumentServer, 0),
		Routes: make(APIRoute, 0),
	}
}

func (b *BaseApp) GetConfig() ServerConfig {
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
