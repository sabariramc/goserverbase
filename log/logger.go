// Package log implements modules for logging
package log

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	m "github.com/sabariramc/goserverbase/v6/log/message"
)

// Logger implementation of Log interface
/*
	Environment Variables
	- SERVICE_NAME: Service name for log messages printed, any string is accepted
	- LOG__LEVEL: Log level for the messages, only following values are accepted
			- TRACE
			- DEBUG
			- INFO
			- NOTICE
			- WARNING
			- ERROR
			- CRITICAL
			- EMERGENCY
Can be overridden by [Config] / [Option]
*/
type Logger struct {
	logLevel    m.LogLevel
	mux         Mux
	moduleName  string
	serviceName string
	audit       AuditLogWriter
	fileTrace   bool
}

func New(options ...Option) *Logger {
	config := defaultConfig
	for _, fn := range options {
		fn(&config)
	}
	return NewWithConfig(context.Background(), config)
}

func NewWithConfig(ctx context.Context, config Config) *Logger {
	l := &Logger{
		logLevel:    config.LogLevel,
		mux:         config.Mux,
		serviceName: config.ServiceName,
		moduleName:  config.ModuleName,
		fileTrace:   config.FileTrace,
		audit:       config.Audit,
	}
	if config.LogLevel.Level == m.TRACE {
		l.Notice(ctx, "log level is set as TRACE", nil)
	}
	return l
}

func (l *Logger) NewResourceLogger(resourceName string) Log {
	newLog := *l
	newLog.SetModuleName(resourceName)
	return &newLog
}

func (l *Logger) AddLogWriter(ctx context.Context, w LogWriter) {
	l.mux.AddLogWriter(ctx, w)
}

func (l *Logger) GetLogLevel() m.LogLevel {
	return l.logLevel
}

func (l *Logger) SetModuleName(moduleName string) {
	l.moduleName = moduleName
}

func (l *Logger) Audit(ctx context.Context, msg interface{}) error {
	if l.audit == nil {
		return nil
	}
	return l.audit.WriteMessage(ctx, msg)
}

// Trace use for package logs
func (l *Logger) Trace(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, m.TRACE, message, logObject)
}

func (l *Logger) Debug(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, m.DEBUG, message, logObject)
}

func (l *Logger) Info(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, m.INFO, message, logObject)
}

func (l *Logger) Notice(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, m.NOTICE, message, logObject)
}

func (l *Logger) Warning(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, m.WARNING, message, logObject)
}

func (l *Logger) Error(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, m.ERROR, message, logObject)
}

func (l *Logger) Emergency(ctx context.Context, message string, err error, logObject ...interface{}) {
	l.print(ctx, m.EMERGENCY, message, logObject)
	if err == nil {
		err = fmt.Errorf("%v", message)
	}
	panic(err)
}

func (l *Logger) Fatal(ctx context.Context, message string, exitCode int, logObject ...interface{}) {
	l.print(ctx, m.FATAL, message, logObject)
	os.Exit(exitCode)
}

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
