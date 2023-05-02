package logwriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v2/log"
	"github.com/sabariramc/goserverbase/v2/utils"
)

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
	logMessage := make(map[string]interface{})
	utils.LenientJsonTransformer(l, &logMessage)
	logMessage["appName"] = l.ServiceName
	logMessage["appNamespace"] = c.appNamespace
	logMessage["correlationId"] = cr.CorrelationId
	cid := log.GetCustomerIdentifier(ctx)
	logMessage["appUserId"] = cid.AppUserId
	logMessage["customerId"] = cid.CustomerId
	logMessage["entityId"] = cid.Id
	logMessage["level"] = l.LogLevelName
	logMessage["message"] = logMessage["short_message"].(string) + " :: " + logMessage["full_message"].(string)
	blob, _ := json.Marshal(logMessage)
	fmt.Println(string(blob))
	return nil
}
