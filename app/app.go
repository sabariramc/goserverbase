package baseapp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/sabariramc/goserverbase/v4/errors"
	"github.com/sabariramc/goserverbase/v4/log"
)

type Shutdown interface {
	Name(ctx context.Context) string
	Close(ctx context.Context) error
}

type BaseApp struct {
	c             *ServerConfig
	log           *log.Logger
	errorNotifier errors.ErrorNotifier
	shutdownHooks []Shutdown
}

func New(appConfig ServerConfig, logger *log.Logger, errorNotifier errors.ErrorNotifier) *BaseApp {
	b := &BaseApp{
		c:             &appConfig,
		errorNotifier: errorNotifier,
		shutdownHooks: make([]Shutdown, 0, 10),
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
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, syscall.SIGTERM, os.Interrupt)
	// go b.monitorSignals(ctx, c)
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

func (b *BaseApp) AddShutdownHook(handler Shutdown) {
	b.shutdownHooks = append(b.shutdownHooks, handler)
}

func (b *BaseApp) Shutdown(ctx context.Context) {
	b.log.Notice(ctx, "waiting for gracefully shutting down of server", nil)
	for _, handler := range b.shutdownHooks {
		ctx, _ = context.WithTimeout(ctx, time.Second)
		b.shutdownModule(ctx, handler)
	}
	b.log.Notice(ctx, "server shutdown", nil)
}

func (b *BaseApp) shutdownModule(ctx context.Context, handler Shutdown) {
	defer func() {
		if rec := recover(); rec != nil {
			b.log.Error(ctx, "panic shutting down service: "+handler.Name(ctx), rec)
		}
	}()
	err := handler.Close(ctx)
	if err != nil {
		b.log.Error(ctx, "error shutting down service: "+handler.Name(ctx), err)
		return
	}
	b.log.Notice(ctx, "shutdown completed for service: "+handler.Name(ctx), nil)
}

func (b *BaseApp) monitorSignals(ctx context.Context, ch chan os.Signal) {
	<-ch
	b.Shutdown(ctx)
}
