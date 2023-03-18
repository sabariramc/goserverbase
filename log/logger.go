package log

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

const ParseErrorMsg = "******************ERROR DURING MARSHAL OF FULL MESSAGE*******************"

type LogLevel struct {
	Level        LogLevelCode
	LogLevelName string
}

type LogLevelCode uint8

const (
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
	ShortMessage    string
	FullMessage     string
	FullMessageType string
	Timestamp       time.Time
	ModuleName      string
	ServiceName     string
}

type Logger struct {
	logLevel    LogLevelCode
	lMux        LogMux
	hostParams  *HostParams
	moduleName  string
	serviceName string
	config      *Config
}

func NewLogger(ctx context.Context, lc *Config, lMux LogMux, moduleName string) *Logger {
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
	}
	logLevel := lc.LogLevel
	if logLevel > int(DEBUG) || logLevel < int(EMERGENCY) {
		l.Warning(ctx, "Erroneous log level - log set to INFO", nil)
		logLevel = int(INFO)
	}
	l.logLevel = LogLevelCode(logLevel)
	return l
}

func (l *Logger) SetModuleName(moduleName string) {
	l.moduleName = moduleName
}

func (l *Logger) Debug(ctx context.Context, shortMessage string, fullMessage interface{}) {
	l.print(ctx, logLevelMap[DEBUG], shortMessage, fullMessage)
}

func (l *Logger) Info(ctx context.Context, shortMessage string, fullMessage interface{}) {
	l.print(ctx, logLevelMap[INFO], shortMessage, fullMessage)
}

func (l *Logger) Notice(ctx context.Context, shortMessage string, fullMessage interface{}) {
	l.print(ctx, logLevelMap[NOTICE], shortMessage, fullMessage)
}

func (l *Logger) Warning(ctx context.Context, shortMessage string, fullMessage interface{}) {
	l.print(ctx, logLevelMap[WARNING], shortMessage, fullMessage)
}

func (l *Logger) Error(ctx context.Context, shortMessage string, fullMessage interface{}) {
	l.print(ctx, logLevelMap[ERROR], shortMessage, fullMessage)
}

func (l *Logger) Critical(ctx context.Context, shortMessage string, fullMessage interface{}) {
	l.print(ctx, logLevelMap[CRITICAL], shortMessage, fullMessage)
}

func (l *Logger) Alert(ctx context.Context, shortMessage string, fullMessage interface{}) {
	l.print(ctx, logLevelMap[ALERT], shortMessage, fullMessage)
}

func (l *Logger) Emergency(ctx context.Context, shortMessage string, fullMessage interface{}, err error) {
	l.print(ctx, logLevelMap[EMERGENCY], shortMessage, fullMessage)
	panic(fmt.Errorf("%v : %w", shortMessage, err))
}

func (l *Logger) print(ctx context.Context, level *LogLevel, shortMessage string, fullMessage interface{}) {
	if level.Level > l.logLevel {
		return
	}
	var msg string
	var msgType string
	if fullMessage == nil {
		msg = shortMessage
		msgType = "nil"
	} else {
		msgType = reflect.TypeOf(fullMessage).Name()
		switch v := fullMessage.(type) {
		case string:
			msg = v
		case error:
			msg = v.Error()
		default:
			blob, err := json.MarshalIndent(v, "", "    ")
			if err != nil {
				msg = ParseErrorMsg
			} else {
				msg = string(blob)
			}
		}
	}
	message := &LogMessage{
		LogLevel:        *level,
		ShortMessage:    shortMessage,
		FullMessage:     msg,
		FullMessageType: msgType,
		Timestamp:       time.Now(),
		ModuleName:      l.moduleName,
		ServiceName:     l.serviceName,
	}
	l.lMux.Print(ctx, message)
}
