package errors

import (
	"context"
)

type ErrorNotifier interface {
	Send5XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}, customerIdentifier interface{})
	Send4XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}, customerIdentifier interface{})
}
