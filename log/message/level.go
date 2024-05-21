package message

// LogLevel represents the log level structure with its code and name.
type LogLevel struct {
	Level        LogLevelCode // Level is the numeric code of the log level.
	LogLevelName string       // LogLevelName is the name of the log level.
}

// LogLevelCode is a type representing the numeric code for log levels.
type LogLevelCode uint8

// Log level constants.
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

// logLevelMap maps LogLevelCode to LogLevel.
var logLevelMap = map[LogLevelCode]*LogLevel{
	TRACE:     {Level: TRACE, LogLevelName: "TRACE"},
	DEBUG:     {Level: DEBUG, LogLevelName: "DEBUG"},
	INFO:      {Level: INFO, LogLevelName: "INFO"},
	NOTICE:    {Level: NOTICE, LogLevelName: "NOTICE"},
	WARNING:   {Level: WARNING, LogLevelName: "WARNING"},
	ERROR:     {Level: ERROR, LogLevelName: "ERROR"},
	EMERGENCY: {Level: EMERGENCY, LogLevelName: "EMERGENCY"},
	FATAL:     {Level: FATAL, LogLevelName: "FATAL"},
}

// logLevelInverseMap maps log level names to LogLevel.
var logLevelInverseMap = func() map[string]*LogLevel {
	m := make(map[string]*LogLevel, len(logLevelMap))
	for _, v := range logLevelMap {
		m[v.LogLevelName] = v
	}
	return m
}()

// GetLogLevelWithName returns the LogLevel for the given log level name.
// If the log level name is not found, it returns the LogLevel for "ERROR".
func GetLogLevelWithName(level string) LogLevel {
	logLevel, ok := logLevelInverseMap[level]
	if !ok {
		logLevel = logLevelInverseMap["ERROR"]
	}
	return *logLevel
}

// GetLogLevel returns the LogLevel for the given LogLevelCode.
// If the LogLevelCode is not found, it returns the LogLevel for ERROR.
func GetLogLevel(level LogLevelCode) LogLevel {
	logLevel, ok := logLevelMap[level]
	if !ok {
		logLevel = logLevelMap[ERROR]
	}
	return *logLevel
}
