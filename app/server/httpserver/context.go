package httpserver

import (
	"context"
	"fmt"
)

func (b *HttpServer) SetContextError(ctx context.Context, err error) {
	iSetter := ctx.Value(ContextKeyError)
	if iSetter == nil {
		return
	}
	setter, ok := iSetter.(func(error))
	if !ok {
		panic(fmt.Errorf("context error handler corrupted, error to handle: %w", err))
	}
	setter(err)
}
