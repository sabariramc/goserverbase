package logwriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v6/trace"
	"github.com/sabariramc/goserverbase/v6/log/message"
)

func DefaultLogMapper(ctx context.Context, msg *message.LogMessage) map[string]any {
	cr := trace.ExtractCorrelationParam(ctx)

	return map[string]any{
		"LogMessage":       msg,
		"CorrelationParam": cr,
		"CorrelationID":    cr.CorrelationID,
		"Level":            msg.LogLevelName,
		"Message":          msg.Message,
		"LogObject":        ParseLogObject(msg.LogObject, false),
		"FilePtr":          msg.File,
		"Timestamp":        msg.Timestamp,
	}
}

type LogMapper func(context.Context, *message.LogMessage) map[string]any

// JSONLLogWriter writes log to console in JSONL format
type JSONLLogWriter struct {
	logMapper LogMapper
}

func NewJSONLConsoleWriter(mapper LogMapper) *JSONLLogWriter {
	if mapper == nil {
		mapper = DefaultLogMapper
	}
	return &JSONLLogWriter{
		logMapper: mapper,
	}
}

func (c *JSONLLogWriter) WriteMessage(ctx context.Context, l *message.LogMessage) error {
	msg := c.logMapper(ctx, l)
	blob, _ := json.Marshal(msg)
	fmt.Println(string(blob))
	return nil
}
