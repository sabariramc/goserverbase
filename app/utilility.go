package baseapp

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/sabariramc/goserverbase/v2/log"
)

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, fmt.Errorf("baseapp.GetBytes: %w", err)
	}
	return buf.Bytes(), nil
}

func (b *BaseApp) GetContextWithCorrelation(ctx context.Context, c *log.CorrelationParam) context.Context {
	ctx = context.WithValue(ctx, log.ContextKeyCorrelation, c)
	return ctx
}

func (b *BaseApp) GetContextWithCustomerId(ctx context.Context, c *log.CustomerIdentifier) context.Context {
	ctx = context.WithValue(ctx, log.ContextKeyCustomerIdentifier, c)
	return ctx
}
