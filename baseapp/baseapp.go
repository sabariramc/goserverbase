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
	LoggerConfig *config.LoggerConfig
	AppConfig    *config.ServerConfig
}

type BaseApp struct {
	router        *httprouter.Router
	c             *ServerConfig
	log           *log.Logger
	errorNotifier errors.ErrorNotifier
}

func NewBaseApp(c ServerConfig, lMux log.LogMultipluxer, auditLogger log.AuditLogWriter, errorNotifier errors.ErrorNotifier) *BaseApp {
	b := &BaseApp{
		c:      &c,
		router: httprouter.New(),
	}
	ctx := b.GetCorrelationContext(context.Background(), log.GetDefaultCorrelationParams(c.AppConfig.ServiceName))
	b.log = log.NewLogger(ctx, c.LoggerConfig, lMux, auditLogger, c.LoggerConfig.ServiceName)
	zone, _ := time.Now().Zone()
	b.log.Notice(ctx, "Server Timezone", zone)
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

func (b *BaseApp) SetRouter(router *httprouter.Router) {
	b.router = router
}

func (b *BaseApp) GetConfig() ServerConfig {
	return *b.c
}

func (b *BaseApp) SetConfig(c ServerConfig) {
	b.c = &c
}

func (b *BaseApp) GetLogger() *log.Logger {
	return b.log
}

func (b *BaseApp) SetLogger(l *log.Logger) {
	b.log = l
}
