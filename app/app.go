package baseapp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/sabariramc/goserverbase/v4/errors"
	"github.com/sabariramc/goserverbase/v4/log"
)

type ShutdownHook interface {
	Name(ctx context.Context) string
	Shutdown(ctx context.Context) error
}

type BaseApp struct {
	c             *ServerConfig
	log           *log.Logger
	errorNotifier errors.ErrorNotifier
	shutdownHooks []ShutdownHook
}

func New(appConfig ServerConfig, logger *log.Logger, errorNotifier errors.ErrorNotifier) *BaseApp {
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

func (b *BaseApp) GetLogger() *log.Logger {
	return b.log
}

func (b *BaseApp) SetLogger(l *log.Logger) {
	b.log = l
}

func (b *BaseApp) GetErrorNotifier() errors.ErrorNotifier {
	return b.errorNotifier
}

func (b *BaseApp) StartSignalMonitor(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt)
	go b.monitorSignals(ctx, c)
	return nil
}

func (b *BaseApp) PanicRecovery(ctx context.Context, rec any) (int, []byte) {
	stackTrace := string(debug.Stack())
	b.log.Error(ctx, "Recovered - Panic", rec)
	b.log.Error(ctx, "Recovered - StackTrace", stackTrace)
	err, ok := rec.(error)
	if !ok {
		blob, _ := json.Marshal(rec)
		err = fmt.Errorf("non error panic: %v", string(blob))
	}
	return b.ProcessError(ctx, stackTrace, err)
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
