package baseapp

import (
	"context"
	"net/http"
	"time"

	"sabariram.com/goserverbase/config"
	"sabariram.com/goserverbase/log"

	"github.com/gorilla/mux"
)

type ServerConfig struct {
	LoggerConfig *config.LoggerConfig
	AppConfig    *config.ServerConfig
}

type BaseApp struct {
	router         *mux.Router
	c              *ServerConfig
	logMultipluxer log.LogMultipluxer
	log            *log.Logger
	HostParams     *log.HostParams
}

func NewBaseApp(c ServerConfig, lMux log.LogMultipluxer, auditLogger log.AuditLogWriter) *BaseApp {
	b := &BaseApp{
		c:              &c,
		logMultipluxer: lMux,
		router:         mux.NewRouter().StrictSlash(true),
	}
	ctx := b.GetCorrelationContext(context.Background(), log.GetDefaultCorrelationParams(c.AppConfig.ServiceName))
	b.log = log.NewLogger(ctx, c.LoggerConfig, lMux, auditLogger)
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

func (b *BaseApp) GetRouter() *mux.Router {
	return b.router
}

func (b *BaseApp) SetRouter(router *mux.Router) {
	b.router = router
}

func (b *BaseApp) GetConfig() ServerConfig {
	return *b.c
}

func (b *BaseApp) SetConfig(c ServerConfig) {
	b.c = &c
}

func (b *BaseApp) GetLogMultipluxer() log.LogMultipluxer {
	return b.logMultipluxer
}

func (b *BaseApp) SetLogMultipluxer(logMux log.LogMultipluxer) {
	b.logMultipluxer = logMux
}

func (b *BaseApp) GetLogger() *log.Logger {
	return b.log
}

func (b *BaseApp) SetLogger(l *log.Logger) {
	b.log = l
}
