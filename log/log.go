package log

import "context"

// Log defines interface for logging used in the rest of the package
type Log interface {
	NewResourceLogger(resourceName string) Log
	Audit(ctx context.Context, msg interface{}) error
	Trace(ctx context.Context, message string, logObject ...interface{})
	Debug(ctx context.Context, message string, logObject ...interface{})
	Info(ctx context.Context, message string, logObject ...interface{})
	Notice(ctx context.Context, message string, logObject ...interface{})
	Warning(ctx context.Context, message string, logObject ...interface{})
	Error(ctx context.Context, message string, logObject ...interface{})
	Emergency(ctx context.Context, message string, err error, logObject ...interface{})
	Fatal(ctx context.Context, message string, exitCode int, logObject ...interface{})
	GetLogLevel() LogLevel
	AddLogWriter(context.Context, LogWriter)
}
