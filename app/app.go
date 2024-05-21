// Package baseapp implements lightweight abstract base framework for a microservice application
package baseapp

import (
	"context"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/notifier"
	"github.com/sabariramc/goserverbase/v6/correlation"
)

// BaseApp represents a basic application structure with configuration, logging, status check, health check and  shutdown functionality.
type BaseApp struct {
	c             *ServerConfig     // Configuration for the server.
	log           log.Log           // Logger instance for the application.
	notifier      notifier.Notifier // Notifier instance for the application.
	shutdownHooks []ShutdownHook    // List of shutdown hooks to be executed during application shutdown.
	healthHooks   []HealthCheckHook // List of health check hooks.
	statusHooks   []StatusCheckHook // List of status check hooks.
	shutdownWg    sync.WaitGroup    // WaitGroup for synchronizing shutdown.
}

// New creates a new instance of BaseApp with the provided configuration, logger, and notifier.
func New(appConfig ServerConfig, logger log.Log, notifier notifier.Notifier) *BaseApp {
	b := &BaseApp{
		c:             &appConfig,
		notifier:      notifier,
		shutdownHooks: make([]ShutdownHook, 0, 10),
	}
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), correlation.GetDefaultCorrelationParam(appConfig.ServiceName))
	b.log = logger.NewResourceLogger("BaseApp")
	zone, _ := time.Now().Zone()
	b.log.Notice(ctx, "Timezone", zone)
	b.shutdownWg.Add(1)
	return b
}

// GetConfig returns the server configuration associated with the BaseApp.
func (b *BaseApp) GetConfig() ServerConfig {
	return *b.c
}

// GetLogger returns the logger associated with the BaseApp.
func (b *BaseApp) GetLogger() log.Log {
	return b.log
}

// SetNotifier sets the notifier for the BaseApp.
func (b *BaseApp) SetNotifier(notifier notifier.Notifier) {
	b.notifier = notifier
}

// WaitForCompleteShutDown waits until the BaseApp completes shutdown.
func (b *BaseApp) WaitForCompleteShutDown() {
	b.shutdownWg.Wait()
}

// GetNotifier returns the notifier associated with the BaseApp.
func (b *BaseApp) GetNotifier() notifier.Notifier {
	return b.notifier
}
