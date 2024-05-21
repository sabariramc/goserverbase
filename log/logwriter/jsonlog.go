package logwriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/log/message"
)

// DefaultLogMapper is the default log mapper function.
func DefaultLogMapper(ctx context.Context, msg *message.LogMessage) map[string]interface{} {
	cr := correlation.ExtractCorrelationParam(ctx)

	return map[string]interface{}{
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

// LogMapper defines the function signature for log mapping.
type LogMapper func(context.Context, *message.LogMessage) map[string]interface{}

// JSONLLogWriter writes logs to console in JSONL format.
type JSONLLogWriter struct {
	logMapper LogMapper
}

// NewJSONLConsoleWriter creates a new JSONLLogWriter instance.
func NewJSONLConsoleWriter(mapper LogMapper) *JSONLLogWriter {
	if mapper == nil {
		mapper = DefaultLogMapper
	}
	return &JSONLLogWriter{
		logMapper: mapper,
	}
}

// WriteMessage writes a log message in JSONL format.
func (c *JSONLLogWriter) WriteMessage(ctx context.Context, l *message.LogMessage) error {
	msg := c.logMapper(ctx, l)
	blob, _ := json.Marshal(msg)
	fmt.Println(string(blob))
	return nil
}
