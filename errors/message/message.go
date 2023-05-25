package message

import (
	"context"
	"time"

	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/sabariramc/goserverbase/v3/utils"
)

func CreateErrorMessage(ctx context.Context, serviceName, errorCode string, err error, stackTrace string, errorData interface{}, alertType string) *utils.Message {
	msg := utils.NewMessage("error", errorCode)
	msg.AddPayload("category", utils.Payload{"entity": map[string]string{"category": alertType}})
	msg.AddPayload("source", utils.Payload{"entity": map[string]string{"source": serviceName}})
	msg.AddPayload("stackTrace", utils.Payload{"entity": map[string]interface{}{"stackTrace": stackTrace, "error": err}})
	msg.AddPayload("version", utils.Payload{"entity": map[string]string{"version": "v1"}})
	msg.AddPayload("errorData", utils.Payload{"entity": map[string]interface{}{"errorData": errorData}})
	msg.AddPayload("identity", utils.Payload{"entity": log.GetCustomerIdentifier(ctx).GetPayload()})
	msg.AddPayload("correlation", utils.Payload{"entity": log.GetCorrelationParam(ctx).GetPayload()})
	msg.AddPayload("timestamp", utils.Payload{"entity": map[string]int64{"timestamp": time.Now().UnixMilli()}})
	return msg
}
