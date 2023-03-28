package errors

import (
	"context"
)

type ErrorNotifier interface {
	Send(ctx context.Context, errorCode string, err error, errorData interface{}, customerIdentifier interface{})
}
