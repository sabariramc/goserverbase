package logwriter

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/log/message"
)

// ConsoleLogWriter writes logs to the console.
type ConsoleLogWriter struct {
}

// NewConsoleWriter creates a new instance of ConsoleLogWriter.
func NewConsoleWriter() *ConsoleLogWriter {
	return &ConsoleLogWriter{}
}

// Start starts the ConsoleLogWriter to listen on the provided log channel and write logs to the console.
func (c *ConsoleLogWriter) Start(logChannel chan message.MuxLogMessage) {
	for log := range logChannel {
		_ = c.WriteMessage(log.Ctx, &log.LogMessage)
	}
}

// GetBufferSize returns the buffer size for the ConsoleLogWriter.
func (c *ConsoleLogWriter) GetBufferSize() int {
	return 1
}

// WriteMessage writes a log message to the console.
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


