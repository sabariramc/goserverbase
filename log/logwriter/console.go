package logwriter

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"sabariram.com/goserverbase/log"
)

type ConsoleWriter struct {
	BaseLogWriter
}

func NewConsoleWriter(hostParam log.HostParams) *ConsoleWriter {
	return &ConsoleWriter{
		BaseLogWriter: BaseLogWriter{hostParam: &hostParam},
	}
}

func (c *ConsoleWriter) Start(logChannel chan log.MultipluxerLogMessage) {
	for log := range logChannel {
		_ = c.WriteMessage(log.Ctx, &log.LogMessage)
	}
}

func (c *ConsoleWriter) WriteMessage(ctx context.Context, l *log.LogMessage) error {
	cr := GetCorrelationParam(ctx)
	b, err := json.Marshal(l.FullMessage)
	var fullMessage string
	if err != nil {
		fullMessage = parseErrorMsg
	} else {
		fullMessage = string(b)
	}
	fmt.Printf("[%v] [%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp, l.LogLevelName, cr.CorrelationId, cr.ServiceName, l.ShortMessage, reflect.TypeOf(l.FullMessage), fullMessage)
	return nil
}
