package baseapp

import (
	"context"
	"time"

	"github.com/sabariramc/goserverbase/v5/errors"
	"github.com/sabariramc/goserverbase/v5/log"
)

type ShutdownHook interface {
	Name(ctx context.Context) string
	Shutdown(ctx context.Context) error
}

type BaseApp struct {
	c             *ServerConfig
	log           log.Log
	errorNotifier errors.ErrorNotifier
	shutdownHooks []ShutdownHook
}

func New(appConfig ServerConfig, logger log.Log, errorNotifier errors.ErrorNotifier) *BaseApp {
	b := &BaseApp{
		c:             &appConfig,
		errorNotifier: errorNotifier,
		shutdownHooks: make([]ShutdownHook, 0, 10),
	}
	ctx := b.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam(appConfig.ServiceName))
	b.log = logger.NewResourceLogger("BaseApp")
	zone, _ := time.Now().Zone()
	b.log.Notice(ctx, "Timezone", zone)
	return b
}

func (b *BaseApp) GetConfig() ServerConfig {
	return *b.c
}

func (b *BaseApp) GetLogger() log.Log {
	return b.log
}

func (b *BaseApp) SetErrorNotifier(errorNotifier errors.ErrorNotifier) {
	b.errorNotifier = errorNotifier
}
