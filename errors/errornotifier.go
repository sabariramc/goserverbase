package errors

import (
	"context"
)

const (
	ErrorCode4XX = "4XX"
	ErrorCode5XX = "5XX"
)

type ErrorNotifier interface {
	Send5XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error
	Send4XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error
}
