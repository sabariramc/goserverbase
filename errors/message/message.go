// Package message abstracts the creation of error message in the format of utils.Message
package message

import (
	"context"
	"time"

	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/utils"
)

/*CreateErrorMessage creates utils.Message with the provided attributes*/
func CreateErrorMessage(ctx context.Context, serviceName, errorCode string, err error, stackTrace string, errorData interface{}, alertType string) *utils.Message {
	msg := utils.NewMessage("error", errorCode)
	msg.AddPayload("category", utils.Payload{"entity": map[string]string{"category": alertType}})
	msg.AddPayload("source", utils.Payload{"entity": map[string]string{"source": serviceName}})
	stackTracePayload := map[string]string{
		"stackTrace": stackTrace,
	}
	if err != nil {
		stackTracePayload["error"] = err.Error()
	}
	msg.AddPayload("stackTrace", utils.Payload{"entity": stackTracePayload})
	msg.AddPayload("version", utils.Payload{"entity": map[string]string{"version": "v1"}})
	msg.AddPayload("errorData", utils.Payload{"entity": map[string]interface{}{"errorData": errorData}})
	msg.AddPayload("identity", utils.Payload{"entity": correlation.ExtractUserIdentifier(ctx).GetPayload()})
	msg.AddPayload("correlation", utils.Payload{"entity": correlation.ExtractCorrelationParam(ctx).GetPayload()})
	msg.AddPayload("timestamp", utils.Payload{"entity": map[string]int64{"timestamp": time.Now().Unix()}})
	return msg
}
