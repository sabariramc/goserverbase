package baseapp

import (
	"context"
	"net/http"
	"time"

	"sabariram.com/goserverbase/config"
	"sabariram.com/goserverbase/log"

	"github.com/gorilla/mux"
)

type BaseApp struct {
	router         *mux.Router
	config         *config.Config
	logMultipluxer *log.LogMultipluxer
	log            *log.Logger
	HostParams     *log.HostParams
}

func NewBaseApp(c *config.Config, logWriter ...log.LogWriter) *BaseApp {
	b := &BaseApp{
		config: c,
		logMultipluxer: log.NewLogMultipluxer(
			uint8(c.Logger.BufferSize),
			logWriter...,
		),
		router: mux.NewRouter().StrictSlash(true),
	}
	ctx := b.GetCorrelationContext(context.Background(), b.GetDefaultCorrelationParams())
	b.log = log.NewLogger(ctx, c, b.GetLogMultipluxer().GetChannel())
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

func (b *BaseApp) GetConfig() *config.Config {
	return b.config
}

func (b *BaseApp) SetConfig(c *config.Config) {
	b.config = c
}

func (b *BaseApp) GetLogMultipluxer() *log.LogMultipluxer {
	return b.logMultipluxer
}

func (b *BaseApp) SetLogMultipluxer(logMux *log.LogMultipluxer) {
	b.logMultipluxer = logMux
}

func (b *BaseApp) GetLogger() *log.Logger {
	return b.log
}

func (b *BaseApp) SetLogger(l *log.Logger) {
	b.log = l
}
