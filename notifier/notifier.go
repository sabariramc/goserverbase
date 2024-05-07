package notifier

import (
	"context"
)

const (
	NotificationCode4XX = "4XX"
	NotificationCode5XX = "5XX"
)

type Notifier interface {
	Notify5XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error
	Notify4XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error
}
