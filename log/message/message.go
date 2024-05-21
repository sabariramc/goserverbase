package message

import (
	"context"
	"time"
)

// LogMessage represents a log entry with various details.
type LogMessage struct {
	LogLevel                  // Embedding LogLevel struct for log level details.
	Message     string        // Message is the log message content.
	LogObject   []interface{} // LogObject contains additional objects related to the log message.
	Timestamp   time.Time     // Timestamp is the time when the log message was created.
	ModuleName  string        // ModuleName is the name of the module generating the log.
	ServiceName string        // ServiceName is the name of the service generating the log.
	File        string        // File is the file name and line number where the log was generated.
}

// MuxLogMessage represents a log entry along with its context.
type MuxLogMessage struct {
	Ctx        context.Context // Ctx is the context in which the log message was created.
	LogMessage                 // Embedding LogMessage struct for log entry details.
}
