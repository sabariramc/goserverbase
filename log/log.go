package log

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/log/logwriter"
	"github.com/sabariramc/goserverbase/v6/log/message"
)

// Log defines the interface for logging used throughout the package.
type Log interface {
	// NewResourceLogger creates a new logger instance with the specified resource name.
	NewResourceLogger(resourceName string) Log

	// Audit logs an audit message.
	Audit(ctx context.Context, msg interface{}) error

	// Trace logs a trace-level message.
	Trace(ctx context.Context, message string, logObject ...interface{})

	// Debug logs a debug-level message.
	Debug(ctx context.Context, message string, logObject ...interface{})

	// Info logs an info-level message.
	Info(ctx context.Context, message string, logObject ...interface{})

	// Notice logs a notice-level message.
	Notice(ctx context.Context, message string, logObject ...interface{})

	// Warning logs a warning-level message.
	Warning(ctx context.Context, message string, logObject ...interface{})

	// Error logs an error-level message.
	Error(ctx context.Context, message string, logObject ...interface{})

	// Emergency logs an emergency-level message and includes an error.
	Emergency(ctx context.Context, message string, err error, logObject ...interface{})

	// Fatal logs a fatal-level message and exits the program with the specified exit code.
	Fatal(ctx context.Context, message string, exitCode int, logObject ...interface{})

	// GetLogLevel returns the current log level of the logger.
	GetLogLevel() message.LogLevel

	// AddLogWriter adds a new log writer to the logger.
	AddLogWriter(context.Context, logwriter.LogWriter)
}
