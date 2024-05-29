package baseapp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// StartSignalMonitor starts monitoring for OS signals such as SIGTERM and SIGINT and initiates shutdown on receiving them.
//
// This function sets up a channel to receive OS signals and starts a goroutine to monitor those signals.
// When a signal is received, it triggers the server shutdown process.
func (b *BaseApp) StartSignalMonitor(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt)
	go b.monitorSignals(ctx, c)
	return nil
}

// RegisterOnShutdownHook registers a shutdown hook to be executed during server shutdown.
//
// This function appends the provided shutdown handler to the list of shutdown hooks in the BaseApp.
func (b *BaseApp) RegisterOnShutdownHook(handler ShutdownHook) {
	b.shutdownHooks = append(b.shutdownHooks, handler)
}

// Shutdown gracefully shuts down the server by executing registered shutdown hooks.
//
// This function iterates through all registered shutdown hooks, executing each one within a specified timeout context.
// It logs the progress of the shutdown process and ensures all hooks are processed before completing the shutdown.
func (b *BaseApp) Shutdown(ctx context.Context) {
	b.log.Notice(ctx, "Gracefully shutting down server", nil)
	hooksCount := len(b.shutdownHooks)
	for i, hook := range b.shutdownHooks {
		shutdownCtx, _ := context.WithTimeout(ctx, time.Second*2)
		b.log.Notice(ctx, fmt.Sprintf("starting step %v of %v - %v", i+1, hooksCount, hook.Name(ctx)), nil)
		b.processShutdownHook(shutdownCtx, hook)
		b.log.Notice(ctx, fmt.Sprintf("completed step %v of %v", i+1, hooksCount), nil)
	}
	b.log.Notice(ctx, "server shutdown completed", nil)
	b.shutdownWg.Done()
}

// processShutdownHook executes the shutdown logic for a single shutdown hook.
//
// This function runs the shutdown logic for the provided handler within a deferred recovery block to handle any panics.
// It logs any errors that occur during the shutdown process of the handler.
func (b *BaseApp) processShutdownHook(ctx context.Context, handler ShutdownHook) {
	defer func() {
		if rec := recover(); rec != nil {
			b.log.Error(ctx, "panic shutting down: "+handler.Name(ctx), rec)
		}
	}()
	err := handler.Shutdown(ctx)
	if err != nil {
		b.log.Error(ctx, "error shutting down: "+handler.Name(ctx), err)
		return
	}
}

// monitorSignals monitors OS signals and initiates server shutdown upon receiving them.
//
// This function blocks until an OS signal is received on the provided channel, then calls the Shutdown method.
func (b *BaseApp) monitorSignals(ctx context.Context, ch chan os.Signal) {
	<-ch
	b.Shutdown(ctx)
}
