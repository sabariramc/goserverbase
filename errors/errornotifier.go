package errors

import (
	"context"
)

type ErrorNotifier interface {
	Send(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}, customerIdentifier interface{})
}
