package message

import (
	"context"
	"time"
)

type Log struct {
	LogLevel
	Message     string
	LogObject   []interface{}
	Timestamp   time.Time
	ModuleName  string
	ServiceName string
	File        string
}

type MuxLog struct {
	Ctx context.Context
	Log
}
