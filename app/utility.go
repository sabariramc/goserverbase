package baseapp

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/log"
)

func (b *BaseApp) GetContextWithCorrelation(ctx context.Context, c *log.CorrelationParam) context.Context {
	return log.GetContextWithCorrelationParam(ctx, c)
}

func (b *BaseApp) GetContextWithCustomerId(ctx context.Context, c *log.UserIdentifier) context.Context {
	return log.GetContextWithUserIdentifier(ctx, c)
}
