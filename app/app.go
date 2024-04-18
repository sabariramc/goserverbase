package baseapp

import (
	"context"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v5/errors"
	"github.com/sabariramc/goserverbase/v5/log"
)

type BaseApp struct {
	c             *ServerConfig
	log           log.Log
	notifier      errors.ErrorNotifier
	shutdownHooks []ShutdownHook
	healthHooks   []HealthCheckHook
	statusHooks   []StatusCheckHook
	shutdownWg    sync.WaitGroup
}

func New(appConfig ServerConfig, logger log.Log, notifier errors.ErrorNotifier) *BaseApp {
	b := &BaseApp{
		c:             &appConfig,
		notifier:      notifier,
		shutdownHooks: make([]ShutdownHook, 0, 10),
	}
	ctx := b.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam(appConfig.ServiceName))
	b.log = logger.NewResourceLogger("BaseApp")
	zone, _ := time.Now().Zone()
	b.log.Notice(ctx, "Timezone", zone)
	b.shutdownWg.Add(1)
	return b
}

func (b *BaseApp) GetConfig() ServerConfig {
	return *b.c
}

func (b *BaseApp) GetLogger() log.Log {
	return b.log
}

func (b *BaseApp) SetErrorNotifier(errorNotifier errors.ErrorNotifier) {
	b.notifier = errorNotifier
}

func (b *BaseApp) WaitForCompleteShutDown() {
	b.shutdownWg.Wait()
}

func (b *BaseApp) GetNotifier() errors.ErrorNotifier {
	return b.notifier
}
