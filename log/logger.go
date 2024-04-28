package log

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"
)

type LogLevel struct {
	Level        LogLevelCode
	LogLevelName string
}

type LogLevelCode uint8

const (
	TRACE     LogLevelCode = 8
	DEBUG     LogLevelCode = 7
	INFO      LogLevelCode = 6
	NOTICE    LogLevelCode = 5
	WARNING   LogLevelCode = 4
	ERROR     LogLevelCode = 3
	EMERGENCY LogLevelCode = 2
	FATAL     LogLevelCode = 0
)

var logLevelMap = map[LogLevelCode]*LogLevel{
	TRACE:     {Level: TRACE, LogLevelName: "TRACE"},
	DEBUG:     {Level: DEBUG, LogLevelName: "DEBUG"},
	INFO:      {Level: INFO, LogLevelName: "INFO"},
	NOTICE:    {Level: NOTICE, LogLevelName: "NOTICE"},
	WARNING:   {Level: WARNING, LogLevelName: "WARNING"},
	ERROR:     {Level: ERROR, LogLevelName: "ERROR"},
	EMERGENCY: {Level: EMERGENCY, LogLevelName: "CRITICAL"},
	FATAL:     {Level: FATAL, LogLevelName: "EMERGENCY"},
}

var logLevelInverseMap map[string]*LogLevel

func init() {
	logLevelInverseMap = make(map[string]*LogLevel, 8)
	for _, v := range logLevelMap {
		logLevelInverseMap[v.LogLevelName] = v
	}
}

func GetLogLevelMap(level LogLevelCode) LogLevel {
	l, ok := logLevelMap[level]
	if !ok {
		l = logLevelMap[INFO]
	}
	return LogLevel{l.Level, l.LogLevelName}
}

type LogMessage struct {
	LogLevel
	Message     string
	LogObject   []interface{}
	Timestamp   time.Time
	ModuleName  string
	ServiceName string
	File        string
}

type Logger struct {
	logLevel    *LogLevel
	lMux        LogMux
	moduleName  string
	serviceName string
	config      *Config
	audit       AuditLogWriter
}

func (l *Logger) NewResourceLogger(resourceName string) Log {
	newLog := *l
	newLog.SetModuleName(resourceName)
	return &newLog
}

func NewWithDefaultConfig(ctx context.Context, lc *Config, moduleName string) *Logger {
	return New(ctx, lc, moduleName, NewDefaultLogMux(), nil)
}

func New(ctx context.Context, lc *Config, moduleName string, lMux LogMux, audit AuditLogWriter) *Logger {
	l := &Logger{
		logLevel:    logLevelMap[INFO],
		lMux:        lMux,
		moduleName:  moduleName,
		serviceName: lc.ServiceName,
		config:      lc,
		audit:       audit,
	}
	logLevel, ok := logLevelInverseMap[lc.LogLevelName]
	if !ok {
		logLevel = logLevelInverseMap["INFO"]
		l.Warning(ctx, "Erroneous log level - log level set as INFO", nil)
	}
	if logLevel.Level == TRACE {
		l.Notice(ctx, "log level is set as TRACE", nil)
	}
	l.logLevel = logLevel
	return l
}

func (l *Logger) AddLogWriter(ctx context.Context, w LogWriter) {
	l.lMux.AddLogWriter(ctx, w)
}

func (l *Logger) GetLogLevel() LogLevel {
	return *l.logLevel
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

func (l *Logger) Trace(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, logLevelMap[TRACE], message, logObject)
}

func (l *Logger) Debug(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, logLevelMap[DEBUG], message, logObject)
}

func (l *Logger) Info(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, logLevelMap[INFO], message, logObject)
}

func (l *Logger) Notice(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, logLevelMap[NOTICE], message, logObject)
}

func (l *Logger) Warning(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, logLevelMap[WARNING], message, logObject)
}

func (l *Logger) Error(ctx context.Context, message string, logObject ...interface{}) {
	l.print(ctx, logLevelMap[ERROR], message, logObject)
}

func (l *Logger) Emergency(ctx context.Context, message string, err error, logObject ...interface{}) {
	l.print(ctx, logLevelMap[EMERGENCY], message, logObject)
	if err == nil {
		err = fmt.Errorf("%v", message)
	}
	panic(err)
}

func (l *Logger) Fatal(ctx context.Context, message string, exitCode int, logObject ...interface{}) {
	l.print(ctx, logLevelMap[FATAL], message, logObject)
	os.Exit(exitCode)
}

func (l *Logger) print(ctx context.Context, level *LogLevel, message string, logObject []interface{}) {
	if level.Level > l.logLevel.Level {
		return
	}
	msg := &LogMessage{
		LogLevel:    *level,
		Message:     message,
		LogObject:   logObject,
		Timestamp:   time.Now(),
		ModuleName:  l.moduleName,
		ServiceName: l.serviceName,
	}
	if l.config.FileTrace {
		_, fileName, lineNumber, _ := runtime.Caller(2)
		msg.File = fmt.Sprintf("%v:%v", fileName, lineNumber)
	}
	l.lMux.Print(ctx, msg)
}
