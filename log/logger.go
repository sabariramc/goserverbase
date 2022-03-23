package log

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"sabariram.com/goserverbase/config"
)

const ParseErrorMsg = "******************ERROR DURING MARSHAL OF FULLMESSAGE*******************"

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
	ShortMessage    string `json:"short_message"`
	FullMessage     string `json:"full_message"`
	FullMessageType string
	Timestamp       time.Time `json:"timestamp"`
}

type AuditLogMessage struct {
	Application    string                 `json:"application"`
	Actor          string                 `json:"actor"`
	Action         string                 `json:"action"`
	Target         string                 `json:"target"`
	Description    string                 `json:"description"`
	Timestamp      time.Time              `json:"timestamp"`
	Correlation    CorrelationParmas      `json:"correlation"`
	AdditionalData map[string]interface{} `json:"additonalData"`
}

type Logger struct {
	logLevel    LogLevel
	lMux        LogMultipluxer
	hostParams  *HostParams
	auditLogger AuditLogWriter
}

func NewLogger(ctx context.Context, lc *config.LoggerConfig, lMux LogMultipluxer, auditLogger AuditLogWriter) *Logger {
	l := &Logger{
		logLevel:    INFO,
		lMux:        lMux,
		auditLogger: auditLogger,
		hostParams: &HostParams{
			Version:     lc.Version,
			Host:        lc.Host,
			ServiceName: lc.ServiceName,
		},
	}
	logLevel := lc.LogLevel
	if logLevel > int(DEBUG) || logLevel < int(EMERGENCY) {
		l.Warning(ctx, "Erronous log level - log set to INFO", nil)
		logLevel = int(INFO)
	}
	l.logLevel = LogLevel(logLevel)
	return l
}

func (l *Logger) Audit(ctx context.Context, actor, action, target, desciption string, additionalData map[string]interface{}) {
	_ = l.auditLogger.WriteAuditMessage(ctx, &AuditLogMessage{
		Application:    l.hostParams.ServiceName,
		Actor:          actor,
		Action:         action,
		Description:    desciption,
		Target:         target,
		Timestamp:      time.Now(),
		AdditionalData: additionalData,
	})
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
			blob, err := json.Marshal(v)
			if err != nil {
				msg = ParseErrorMsg
			} else {
				msg = string(blob)
			}
		}
	}
	message := &LogMessage{
		LogLevelMap:     *level,
		ShortMessage:    shortMessage,
		FullMessage:     msg,
		FullMessageType: ,
		Timestamp:       time.Now()}
	l.lMux.Print(ctx, message)
}
