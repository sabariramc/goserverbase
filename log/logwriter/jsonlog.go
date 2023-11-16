package logwriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v4/log"
)

type logIgMessage struct {
	*log.LogMessage
	*log.CorrelationParam
	*log.CustomerIdentifier
	AppName       string `json:"appName"`
	AppNamespace  string `json:"appNamespace"`
	CorrelationID string `json:"correlationId"`
	Level         string `json:"level"`
	Message       string `json:"message"`
}

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

type JsonLogConsoleWriter struct {
	hostParam    *log.HostParams
	appNamespace string
	logMapper    LogMapper
}

func NewLogIqConsoleWriter(hostParam log.HostParams, appNamespace string, mapper LogMapper) *JsonLogConsoleWriter {
	if mapper == nil {
		mapper = DefaultLogMapper
	}
	return &JsonLogConsoleWriter{
		hostParam:    &hostParam,
		appNamespace: appNamespace,
		logMapper:    mapper,
	}
}

func (c *JsonLogConsoleWriter) WriteMessage(ctx context.Context, l *log.LogMessage) error {
	msg := c.logMapper(ctx, l)
	blob, _ := json.Marshal(msg)
	fmt.Println(string(blob))
	return nil
}
