package logwriter

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v6/log/message"
	"github.com/sabariramc/goserverbase/v6/correlation"
)

// ConsoleLogWriter write logs to console
type ConsoleLogWriter struct {
}

func NewConsoleWriter() *ConsoleLogWriter {
	return &ConsoleLogWriter{}
}

func (c *ConsoleLogWriter) Start(logChannel chan message.MuxLogMessage) {
	for log := range logChannel {
		_ = c.WriteMessage(log.Ctx, &log.LogMessage)
	}
}

func (c *ConsoleLogWriter) GetBufferSize() int {
	return 1
}

func (c *ConsoleLogWriter) WriteMessage(ctx context.Context, l *message.LogMessage) error {
	cr := correlation.ExtractCorrelationParam(ctx)
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
