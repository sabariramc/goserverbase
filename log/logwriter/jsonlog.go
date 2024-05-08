package logwriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v5/log"
)

func DefaultLogMapper(ctx context.Context, msg *log.LogMessage) map[string]any {
	cr := log.ExtractCorrelationParam(ctx)

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

type LogMapper func(context.Context, *log.LogMessage) map[string]any

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

func (c *JSONLLogWriter) WriteMessage(ctx context.Context, l *log.LogMessage) error {
	msg := c.logMapper(ctx, l)
	blob, _ := json.Marshal(msg)
	fmt.Println(string(blob))
	return nil
}
