package logwriter

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v4/log"
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

func (c *ConsoleWriter) WriteMessage(ctx context.Context, l *log.LogMessage) error {
	cr := log.GetCorrelationParam(ctx)
	if l != nil {
		fmt.Printf("[%v] [%v] [%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp, l.LogLevelName, cr.CorrelationId, l.ServiceName, l.ModuleName, l.Message, GetLogObjectType(l.LogObject), ParseLogObject(l.LogObject, true))
	} else {
		fmt.Printf("[%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp, l.LogLevelName, cr.CorrelationId, l.ServiceName, l.ModuleName, l.Message)
	}
	return nil
}
