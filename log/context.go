package log

import "context"

type ContextVariable string

const (
	ContextKeyCorrelation        ContextVariable = "correlationParam"
	ContextKeyCustomerIdentifier ContextVariable = "customerIdentity"
)

func GetContextWithCorrelation(ctx context.Context, c *CorrelationParam) context.Context {
	ctx = context.WithValue(ctx, ContextKeyCorrelation, c)
	return ctx
}

func GetContextWithCustomerID(ctx context.Context, c *CustomerIdentifier) context.Context {
	ctx = context.WithValue(ctx, ContextKeyCustomerIdentifier, c)
	return ctx
}
