package log

import (
	"context"
	"fmt"
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
	CRITICAL  LogLevelCode = 2
	ALERT     LogLevelCode = 1
	EMERGENCY LogLevelCode = 0
)

var logLevelMap = map[LogLevelCode]*LogLevel{
	DEBUG:     {Level: DEBUG, LogLevelName: "DEBUG"},
	INFO:      {Level: INFO, LogLevelName: "INFO"},
	NOTICE:    {Level: NOTICE, LogLevelName: "NOTICE"},
	WARNING:   {Level: WARNING, LogLevelName: "WARNING"},
	ERROR:     {Level: ERROR, LogLevelName: "ERROR"},
	CRITICAL:  {Level: CRITICAL, LogLevelName: "CRITICAL"},
	ALERT:     {Level: ALERT, LogLevelName: "ALERT"},
	EMERGENCY: {Level: EMERGENCY, LogLevelName: "EMERGENCY"},
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
	LogObject   interface{}
	Timestamp   time.Time
	ModuleName  string
	ServiceName string
}

type Logger struct {
	logLevel    LogLevelCode
	lMux        LogMux
	hostParams  *HostParams
	moduleName  string
	serviceName string
	config      *Config
	audit       AuditLogWriter
}

func (l *Logger) NewResourceLogger(resourceName string) *Logger {
	newLog := *l
	newLog.SetModuleName(resourceName)
	return &newLog
}

func NewLogger(ctx context.Context, lc *Config, moduleName string, lMux LogMux, audit AuditLogWriter) *Logger {
	l := &Logger{
		logLevel:    INFO,
		lMux:        lMux,
		moduleName:  moduleName,
		serviceName: lc.ServiceName,
		config:      lc,
		hostParams: &HostParams{
			Version:     lc.Version,
			Host:        lc.Host,
			ServiceName: lc.ServiceName,
		},
		audit: audit,
	}
	logLevel := lc.LogLevel
	if logLevel > int(DEBUG) || logLevel < int(EMERGENCY) {
		l.Warning(ctx, "Erroneous log level - log set to INFO", nil)
		logLevel = int(INFO)
	}
	l.logLevel = LogLevelCode(logLevel)
	return l
}

func (l *Logger) GetLogLevel() LogLevel {
	return GetLogLevelMap(l.logLevel)
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

func (l *Logger) Debug(ctx context.Context, message string, logObject interface{}) {
	l.print(ctx, logLevelMap[DEBUG], message, logObject)
}

func (l *Logger) Info(ctx context.Context, message string, logObject interface{}) {
	l.print(ctx, logLevelMap[INFO], message, logObject)
}

func (l *Logger) Notice(ctx context.Context, message string, logObject interface{}) {
	l.print(ctx, logLevelMap[NOTICE], message, logObject)
}

func (l *Logger) Warning(ctx context.Context, message string, logObject interface{}) {
	l.print(ctx, logLevelMap[WARNING], message, logObject)
}

func (l *Logger) Error(ctx context.Context, message string, logObject interface{}) {
	l.print(ctx, logLevelMap[ERROR], message, logObject)
}

func (l *Logger) Critical(ctx context.Context, message string, logObject interface{}) {
	l.print(ctx, logLevelMap[CRITICAL], message, logObject)
}

func (l *Logger) Alert(ctx context.Context, message string, logObject interface{}) {
	l.print(ctx, logLevelMap[ALERT], message, logObject)
}

func (l *Logger) Emergency(ctx context.Context, message string, err error, logObject interface{}) {
	l.print(ctx, logLevelMap[EMERGENCY], message, logObject)
	panic(fmt.Errorf("%v : %w", message, err))
}

func (l *Logger) Log(ctx context.Context, logLevel int, message string, err error, logObject interface{}) {
	if logLevel == int(EMERGENCY) {
		l.Emergency(ctx, message, err, logObject)
		return
	}
	level := GetLogLevelMap(LogLevelCode(logLevel))
	l.print(ctx, &level, message, logObject)
}

func (l *Logger) print(ctx context.Context, level *LogLevel, message string, logObject interface{}) {
	if level.Level > l.logLevel {
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
	l.lMux.Print(ctx, msg)
}
