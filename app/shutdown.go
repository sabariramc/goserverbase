package baseapp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (b *BaseApp) StartSignalMonitor(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt)
	go b.monitorSignals(ctx, c)
	return nil
}

func (b *BaseApp) AddShutdownHook(handler ShutdownHook) {
	b.shutdownHooks = append(b.shutdownHooks, handler)
}

func (b *BaseApp) Shutdown(ctx context.Context) {
	b.log.Notice(ctx, "Gracefully shutting down server", nil)
	hooksCount := len(b.shutdownHooks)
	for i, j := hooksCount-1, 1; i >= 0; i, j = i-1, j+1 {
		b.log.Notice(ctx, fmt.Sprintf("shutdown step %v of %v", j, hooksCount), nil)
		ctx, _ = context.WithTimeout(ctx, time.Second)
		b.shutdownModule(ctx, b.shutdownHooks[i])
	}
	b.log.Notice(ctx, "server shutdown completed", nil)
}

func (b *BaseApp) shutdownModule(ctx context.Context, handler ShutdownHook) {
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
	b.log.Notice(ctx, "shutdown completed for: "+handler.Name(ctx), nil)
}

func (b *BaseApp) monitorSignals(ctx context.Context, ch chan os.Signal) {
	<-ch
	b.Shutdown(ctx)
}
