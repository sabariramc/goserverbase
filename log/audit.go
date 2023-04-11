package log

import "context"

type AuditLogWriter interface {
	WriteMessage(context.Context, interface{}) error
}
