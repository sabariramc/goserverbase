// Package baseapp implements a lightweight abstract base framework for a microservice application.
package baseapp

import (
	"context"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/notifier"
)

// BaseApp represents a basic application structure with configuration, logging, status check, health check, and shutdown functionality.
type BaseApp struct {
	c             *Config           // Configuration for the server.
	log           log.Log           // Logger instance for the application.
	notifier      notifier.Notifier // Notifier instance for the application.
	shutdownHooks []ShutdownHook    // List of shutdown hooks to be executed during application shutdown.
	healthHooks   []HealthCheckHook // List of health check hooks.
	statusHooks   []StatusCheckHook // List of status check hooks.
	shutdownWg    sync.WaitGroup    // WaitGroup for synchronizing shutdown.
}

// New creates a new instance of BaseApp with the provided configuration.
//
// This function initializes a BaseApp instance with default or provided configurations and sets up the logging context.
// It also adds an initial wait group for the shutdown process.
func New(option ...Option) *BaseApp {
	config := GetDefaultConfig()
	for _, opt := range option {
		opt(config)
	}
	return NewWithConfig(config)
}

func NewWithConfig(config *Config) *BaseApp {
	b := &BaseApp{
		c:             config,
		notifier:      config.Notifier,
		shutdownHooks: make([]ShutdownHook, 0, 10),
		log:           config.Log,
	}
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), correlation.NewCorrelationParam(config.ServiceName))
	zone, _ := time.Now().Zone()
	b.log.Notice(ctx, "Timezone", zone)
	b.shutdownWg.Add(1)
	return b
}

// GetConfig returns the server configuration associated with the BaseApp.
//
// This function provides access to the configuration used by the BaseApp.
func (b *BaseApp) GetConfig() Config {
	return *b.c
}

// GetLogger returns the logger associated with the BaseApp.
//
// This function provides access to the logging instance used by the BaseApp.
func (b *BaseApp) GetLogger() log.Log {
	return b.log
}

// SetNotifier sets the notifier for the BaseApp.
//
// This function updates the notifier instance used by the BaseApp.
func (b *BaseApp) SetNotifier(notifier notifier.Notifier) {
	b.notifier = notifier
}

// WaitForCompleteShutDown waits until the BaseApp completes shutdown.
//
// This function blocks until the shutdown process, including all registered shutdown hooks, is complete.
func (b *BaseApp) WaitForCompleteShutDown() {
	b.shutdownWg.Wait()
}

// GetNotifier returns the notifier associated with the BaseApp.
//
// This function provides access to the notifier instance used by the BaseApp.
func (b *BaseApp) GetNotifier() notifier.Notifier {
	return b.notifier
}
