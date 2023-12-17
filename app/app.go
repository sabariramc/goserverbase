package baseapp

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/sabariramc/goserverbase/v4/errors"
	"github.com/sabariramc/goserverbase/v4/log"
)

type BaseApp struct {
	c             *ServerConfig
	log           *log.Logger
	errorNotifier errors.ErrorNotifier
}

func New(appConfig ServerConfig, logger *log.Logger, errorNotifier errors.ErrorNotifier) *BaseApp {
	b := &BaseApp{
		c:             &appConfig,
		errorNotifier: errorNotifier,
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
