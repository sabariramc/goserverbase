package logwriter

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v4/log"
)

type ConsoleWriter struct {
	BaseLogWriter
}

func NewConsoleWriter(hostParam log.HostParams) *ConsoleWriter {
	return &ConsoleWriter{
		BaseLogWriter: BaseLogWriter{hostParam: &hostParam},
	}
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
	fmt.Printf("[%v] [%v] [%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp, l.LogLevelName, cr.CorrelationId, l.ServiceName, l.ModuleName, l.ShortMessage, GetLogObjectType(l.FullMessage), ParseLogObject(l.FullMessage, true))
	return nil
}
