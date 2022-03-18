package log

import (
	"context"
	"time"

	"sabariram.com/goserverbase/config"
)

type LogLevelMap struct {
	Level        LogLevel `json:"level"`
	LogLevelName string   `json:"_x_level_name"`
}

type LogLevel uint8

const (
	DEBUG     LogLevel = 7
	INFO      LogLevel = 6
	NOTICE    LogLevel = 5
	WARNING   LogLevel = 4
	ERROR     LogLevel = 3
	CRITICAL  LogLevel = 2
	ALERT     LogLevel = 1
	EMERGENCY LogLevel = 0
)

var logLevelMap = map[LogLevel]*LogLevelMap{
	DEBUG:     {Level: DEBUG, LogLevelName: "DEBUG"},
	INFO:      {Level: INFO, LogLevelName: "INFO"},
	NOTICE:    {Level: NOTICE, LogLevelName: "NOTICE"},
	WARNING:   {Level: WARNING, LogLevelName: "WARNING"},
	ERROR:     {Level: ERROR, LogLevelName: "ERROR"},
	CRITICAL:  {Level: CRITICAL, LogLevelName: "CRITICAL"},
	ALERT:     {Level: ALERT, LogLevelName: "ALERT"},
	EMERGENCY: {Level: EMERGENCY, LogLevelName: "EMERGENCY"},
}

func GetLogLevelMap(level LogLevel) LogLevelMap {
	l, ok := logLevelMap[level]
	if !ok {
		l = logLevelMap[INFO]
	}
	return LogLevelMap{l.Level, l.LogLevelName}
}

type LogMessage struct {
	LogLevelMap
	ShortMessage string      `json:"short_message"`
	FullMessage  interface{} `json:"full_message"`
	Timestamp    time.Time   `json:"timestamp"`
}

type Logger struct {
	logLevel         LogLevel
	multipluxChannel chan MultipluxerLogMessage
	hostParams       *HostParams
}

func NewLogger(ctx context.Context, config *config.Config, multipluxChannel chan MultipluxerLogMessage) *Logger {
	l := &Logger{
		logLevel:         INFO,
		multipluxChannel: multipluxChannel,
		hostParams: &HostParams{
			Version: config.Logger.Version,
			Host:    config.Logger.Host,
		},
	}
	logLevel := config.Logger.LogLevel
	if logLevel > int(DEBUG) || logLevel < int(EMERGENCY) {
		l.Warning(ctx, "Erronous log level - log set to INFO", nil)
		logLevel = int(INFO)
	}
	l.logLevel = LogLevel(logLevel)
	return l
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
	panic(err)
}

func (l *Logger) print(ctx context.Context, level *LogLevelMap, shortMessage string, fullMessage interface{}) {
	if level.Level > l.logLevel {
		return
	}
	if fullMessage == nil {
		fullMessage = shortMessage
	}
	message := &LogMessage{
		LogLevelMap:  *level,
		ShortMessage: shortMessage,
		FullMessage:  fullMessage,
		Timestamp:    time.Now()}
	l.multipluxChannel <- MultipluxerLogMessage{
		LogMessage: *message,
		Ctx:        ctx,
	}
}
