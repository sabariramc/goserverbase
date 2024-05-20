//
package message

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

var logLevelInverseMap map[string]*LogLevel = func() map[string]*LogLevel {
	a := make(map[string]*LogLevel, 8)
	for _, v := range logLevelMap {
		a[v.LogLevelName] = v
	}
	return a
}()

func GetLogLevelWithName(level string) LogLevel {
	logLevel, ok := logLevelInverseMap[level]
	if !ok {
		logLevel = logLevelInverseMap["ERROR"]
	}
	return *logLevel
}

func GetLogLevel(level LogLevelCode) LogLevel {
	logLevel, ok := logLevelMap[LogLevelCode(level)]
	if !ok {
		logLevel = logLevelMap[ERROR]
	}
	return *logLevel
}
