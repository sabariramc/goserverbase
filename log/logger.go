// Package log implements modules for logging.
package log

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	m "github.com/sabariramc/goserverbase/v6/log/message"
)

// Logger represents the implementation of the log interface.

type Logger struct {
	logLevel    m.LogLevel     // logLevel represents the log level.
	mux         Mux            // mux represents the multiplexer for handling log messages.
	moduleName  string         // moduleName represents the name of the module.
	serviceName string         // serviceName represents the name of the service.
	audit       AuditLogWriter // audit represents the audit log writer.
	fileTrace   bool           // fileTrace indicates whether file tracing is enabled.
}

// New creates a new Logger instance with the specified options.
func New(options ...Option) *Logger {
	config := GetDefaultConfig()
	for _, opt := range options {
		opt(&config)
	}
	l := &Logger{
		logLevel:    config.LogLevel,
		mux:         config.Mux,
		serviceName: config.ServiceName,
		moduleName:  config.ModuleName,
		fileTrace:   config.FileTrace,
		audit:       config.Audit,
	}
	if config.LogLevel.Level == m.TRACE {
		l.Notice(context.Background(), "log level is set as TRACE", nil)
	}
	return l
}

// NewResourceLogger creates a new Logger instance with the specified resource name.
func (l *Logger) NewResourceLogger(resourceName string) Log {
	newLog := *l
	newLog.SetModuleName(resourceName)
	return &newLog
}

// AddLogWriter adds a log writer to the logger.
func (l *Logger) AddLogWriter(ctx context.Context, w LogWriter) {
	l.mux.AddLogWriter(ctx, w)
}

// GetLogLevel returns the current log level.
func (l *Logger) GetLogLevel() m.LogLevel {
	return l.logLevel
}

// SetModuleName sets the module name for the logger.
func (l *Logger) SetModuleName(moduleName string) {
	l.moduleName = moduleName
}

// Audit writes an audit log message.
func (l *Logger) Audit(ctx context.Context, msg interface{}) error {
	if l.audit == nil {
		return nil
	}
	return l.audit.WriteMessage(ctx, msg)
}

// Trace logs a message with TRACE level.
func (l *Logger) Trace(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, m.TRACE, message, logObject)
}

// Debug logs a message with DEBUG level.
func (l *Logger) Debug(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, m.DEBUG, message, logObject)
}

// Info logs a message with INFO level.
func (l *Logger) Info(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, m.INFO, message, logObject)
}

// Notice logs a message with NOTICE level.
func (l *Logger) Notice(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, m.NOTICE, message, logObject)
}

// Warning logs a message with WARNING level.
func (l *Logger) Warning(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, m.WARNING, message, logObject)
}

// Error logs a message with ERROR level.
func (l *Logger) Error(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, m.ERROR, message, logObject)
}

// Emergency logs a message with EMERGENCY level.
func (l *Logger) Emergency(ctx context.Context, message string, err error, logObject ...interface{}) {
	l.print(ctx, m.EMERGENCY, message, logObject)
	if err == nil {
		err = fmt.Errorf("%v", message)
	}
	panic(err)
}

// Fatal logs a message with FATAL level and exits the program with the specified exit code.
func (l *Logger) Fatal(ctx context.Context, message string, exitCode int, logObject ...interface{}) {
	l.print(ctx, m.FATAL, message, logObject)
	os.Exit(exitCode)
}

// print prints the log message with the specified level.
func (l *Logger) print(ctx context.Context, level m.LogLevelCode, message string, logObject []interface{}) {
	if level > l.logLevel.Level {
		return
	}
	msg := &m.LogMessage{
		LogLevel:    m.GetLogLevel(level),
		Message:     message,
		LogObject:   logObject,
		Timestamp:   time.Now(),
		ModuleName:  l.moduleName,
		ServiceName: l.serviceName,
	}
	if l.fileTrace {
		_, fileName, lineNumber, _ := runtime.Caller(2)
		msg.File = fmt.Sprintf("%v:%v", fileName, lineNumber)
	}
	l.mux.Print(ctx, msg)
}
