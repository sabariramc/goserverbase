package logwriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v2/log"
)

type logIgMessage struct {
	*log.LogMessage
	*log.CorrelationParam
	*log.CustomerIdentifier
	AppName       string `json:"appName"`
	AppNamespace  string `json:"appNamespace"`
	CorrelationId string `json:"correlationId"`
	Level         string `json:"level"`
	Message       string `json:"message"`
}

type LogIqConsoleWriter struct {
	hostParam    *log.HostParams
	appNamespace string
}

func NewLogIqConsoleWriter(hostParam log.HostParams, appNamespace string) *LogIqConsoleWriter {
	return &LogIqConsoleWriter{
		hostParam:    &hostParam,
		appNamespace: appNamespace,
	}
}

func (c *LogIqConsoleWriter) WriteMessage(ctx context.Context, l *log.LogMessage) error {
	cr := log.GetCorrelationParam(ctx)
	msg := logIgMessage{
		LogMessage:         l,
		CorrelationParam:   cr,
		CustomerIdentifier: log.GetCustomerIdentifier(ctx),
		AppName:            l.ServiceName,
		AppNamespace:       c.appNamespace,
		CorrelationId:      cr.CorrelationId,
		Level:              l.LogLevelName,
		Message:            l.ShortMessage + " :: " + l.FullMessage,
	}
	blob, _ := json.Marshal(msg)
	fmt.Println(string(blob))
	return nil
}