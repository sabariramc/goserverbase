package errors

import (
	"context"
)

const (
	ERROR_4xx = "4XX"
	ERROR_5xx = "5XX"
)

type ErrorNotifier interface {
	Send5XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}, customerIdentifier interface{}) error
	Send4XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}, customerIdentifier interface{}) error
}
