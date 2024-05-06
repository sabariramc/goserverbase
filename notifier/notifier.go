package notifier

import (
	"context"
)

const (
	NotificationCode4XX = "4XX"
	NotificationCode5XX = "5XX"
)

type Notifier interface {
	Send5XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error
	Send4XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error
}
