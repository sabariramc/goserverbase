package baseapp

import (
	"context"

	"github.com/sabariramc/goserverbase/v2/log"
)

func (b *BaseApp) GetContextWithCorrelation(ctx context.Context, c *log.CorrelationParam) context.Context {
	ctx = context.WithValue(ctx, log.ContextKeyCorrelation, c)
	return ctx
}

func (b *BaseApp) GetContextWithCustomerId(ctx context.Context, c *log.CustomerIdentifier) context.Context {
	ctx = context.WithValue(ctx, log.ContextKeyCustomerIdentifier, c)
	return ctx
}
