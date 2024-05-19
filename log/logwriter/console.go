package logwriter

import (
	"context"
	"fmt"

	cuCtx "github.com/sabariramc/goserverbase/v6/context"
	"github.com/sabariramc/goserverbase/v6/log/message"
)

// ConsoleLogWriter write logs to console
type ConsoleLogWriter struct {
}

func NewConsoleWriter() *ConsoleLogWriter {
	return &ConsoleLogWriter{}
}

func (c *ConsoleLogWriter) Start(logChannel chan message.MuxLog) {
	for log := range logChannel {
		_ = c.WriteMessage(log.Ctx, &log.Log)
	}
}

func (c *ConsoleLogWriter) GetBufferSize() int {
	return 1
}

func (c *ConsoleLogWriter) WriteMessage(ctx context.Context, l *message.Log) error {
	cr := cuCtx.ExtractCorrelationParam(ctx)
	if l.File != "" {
		fmt.Println(l.File)
	}
	if l.LogObject != nil {
		fmt.Printf("[%v] [%v] [%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp.Format(timeFormat), l.LogLevelName, cr.CorrelationID, l.ServiceName, l.ModuleName, l.Message, GetLogObjectType(l.LogObject), ParseLogObject(l.LogObject, true))
	} else {
		fmt.Printf("[%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp.Format(timeFormat), l.LogLevelName, cr.CorrelationID, l.ServiceName, l.ModuleName, l.Message)
	}
	return nil
}
