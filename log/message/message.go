package message

import (
	"context"
	"time"
)

type LogMessage struct {
	LogLevel
	Message     string
	LogObject   []interface{}
	Timestamp   time.Time
	ModuleName  string
	ServiceName string
	File        string
}

type MuxLogMessage struct {
	Ctx context.Context
	LogMessage
}
