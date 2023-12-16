package httpserver

import (
	"context"
	"fmt"
)

func (h *HttpServer) SetErrorInContext(ctx context.Context, err error) {
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

func (h *HttpServer) SetStackTrackInContext(ctx context.Context, stackTrace string) {
	iSetter := ctx.Value(ContextKeyHandlerErrorStackTrace)
	if iSetter == nil {
		return
	}
	setter, ok := iSetter.(func(string))
	if !ok {
		h.log.Emergency(ctx, "Unable to set context stack trace", nil, fmt.Errorf("context stack trace handler corrupted, stack trace to handle: %v", stackTrace))
	}
	setter(stackTrace)
}
