package errors

import (
	"context"
)

type ErrorNotifier interface {
	Send(ctx context.Context, err error)
}
