package logwriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v4/log"
)

func DefaultLogMapper(ctx context.Context, msg *log.LogMessage) map[string]any {
	cr := log.GetCorrelationParam(ctx)

	return map[string]any{
		"LogMessage":       msg,
		"CorrelationParam": cr,
		"CorrelationID":    cr.CorrelationId,
		"Level":            msg.LogLevelName,
		"Message":          msg.Message,
		"LogObject":        ParseLogObject(msg.LogObject, false),
	}
}

type LogMapper func(context.Context, *log.LogMessage) map[string]any

type JSONLConsoleWriter struct {
	hostParam    *log.HostParams
	appNamespace string
	logMapper    LogMapper
}

func NewJSONLConsoleWriter(hostParam log.HostParams, appNamespace string, mapper LogMapper) *JSONLConsoleWriter {
	if mapper == nil {
		mapper = DefaultLogMapper
	}
	return &JSONLConsoleWriter{
		hostParam:    &hostParam,
		appNamespace: appNamespace,
		logMapper:    mapper,
	}
}

func (c *JSONLConsoleWriter) WriteMessage(ctx context.Context, l *log.LogMessage) error {
	msg := c.logMapper(ctx, l)
	blob, _ := json.Marshal(msg)
	fmt.Println(string(blob))
	return nil
}
