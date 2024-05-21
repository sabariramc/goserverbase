// Package notifier defines an interface for the system to send messages regarding various system events
// that the engineering team should be immediately aware of. It can be used in incident reporting.
package notifier

import (
	"context"
)

// Constants representing notification codes for different types of errors.
const (
	NotificationCode4XX = "4XX" // Notification code for 4XX errors (client errors).
	NotificationCode5XX = "5XX" // Notification code for 5XX errors (server errors).
)

// Notifier defines an interface for sending notifications about system events.
type Notifier interface {
	// Notify5XX sends a notification for 5XX errors (server errors).
	Notify5XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error

	// Notify4XX sends a notification for 4XX errors (client errors).
	Notify4XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error
}
