package httpserver

import (
	"context"
	"fmt"
)

func (h *HttpServer) SetHandlerErrorInContext(ctx context.Context, err error) {
	iSetter := ctx.Value(ContextKeyHandlerError)
	if iSetter == nil {
		return
	}
	setter, ok := iSetter.(func(error))
	if !ok {
		h.log.Emergency(ctx, "Unable to set context error", nil, fmt.Errorf("context error handler corrupted, error to handle: %w", err))
	}
	setter(err)
}
