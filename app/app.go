// Package baseapp implements lightweight abstract base framework for a microservice application
package baseapp

import (
	"context"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/notifier"
)

type BaseApp struct {
	c             *ServerConfig
	log           log.Log
	notifier      notifier.Notifier
	shutdownHooks []ShutdownHook
	healthHooks   []HealthCheckHook
	statusHooks   []StatusCheckHook
	shutdownWg    sync.WaitGroup
}

func New(appConfig ServerConfig, logger log.Log, notifier notifier.Notifier) *BaseApp {
	b := &BaseApp{
		c:             &appConfig,
		notifier:      notifier,
		shutdownHooks: make([]ShutdownHook, 0, 10),
	}
	ctx := log.GetContextWithCorrelationParam(context.Background(), log.GetDefaultCorrelationParam(appConfig.ServiceName))
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

func (b *BaseApp) SetNotifier(notifier notifier.Notifier) {
	b.notifier = notifier
}

func (b *BaseApp) WaitForCompleteShutDown() {
	b.shutdownWg.Wait()
}

func (b *BaseApp) GetNotifier() notifier.Notifier {
	return b.notifier
}
