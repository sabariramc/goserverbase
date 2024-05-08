package logwriter

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v5/log"
)

type ConsoleWriter struct {
}

func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}

func (c *ConsoleWriter) Start(logChannel chan log.MuxLogMessage) {
	for log := range logChannel {
		_ = c.WriteMessage(log.Ctx, &log.LogMessage)
	}
}

func (c *ConsoleWriter) GetBufferSize() int {
	return 1
}

const TimeFormat = "2006-01-02T15:04:05.000Z07:00"

func (c *ConsoleWriter) WriteMessage(ctx context.Context, l *log.LogMessage) error {
	cr := log.ExtractCorrelationParam(ctx)
	if l.File != "" {
		fmt.Println(l.File)
	}
	if l.LogObject != nil {
		fmt.Printf("[%v] [%v] [%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp.Format(TimeFormat), l.LogLevelName, cr.CorrelationID, l.ServiceName, l.ModuleName, l.Message, GetLogObjectType(l.LogObject), ParseLogObject(l.LogObject, true))
	} else {
		fmt.Printf("[%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp.Format(TimeFormat), l.LogLevelName, cr.CorrelationID, l.ServiceName, l.ModuleName, l.Message)
	}
	return nil
}
