package kafkaconsumer

import (
	"github.com/sabariramc/goserverbase/v3/log"
)

func (k *KafkaConsumerServer) GetCorrelationParams(headers map[string]string) *log.CorrelationParam {
	correlationId, ok := headers["x-correlation-id"]
	if !ok {
		return log.GetDefaultCorrelationParam(k.c.ServiceName)
	}
	return &log.CorrelationParam{
		CorrelationId: correlationId,
		ScenarioId:    headers["x-scenario-id"],
		ScenarioName:  headers["x-scenario-name"],
		SessionId:     headers["x-session-id"],
	}
}

func (k *KafkaConsumerServer) GetCustomerId(headers map[string]string) *log.CustomerIdentifier {
	return &log.CustomerIdentifier{
		AppUserId:  headers["x-app-user-id"],
		CustomerId: headers["x-customer-id"],
		Id:         headers["x-entity-id"],
	}
}
