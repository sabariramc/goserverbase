package baseapp

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/log"
)

func (b *BaseApp) GetContextWithCorrelation(ctx context.Context, c *log.CorrelationParam) context.Context {
	return log.GetContextWithCorrelation(ctx, c)
}

func (b *BaseApp) GetContextWithCustomerId(ctx context.Context, c *log.CustomerIdentifier) context.Context {
	return log.GetContextWithCustomerId(ctx, c)
}
