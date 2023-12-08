package kafkaconsumer

import (
	"context"

	"github.com/sabariramc/goserverbase/v4/app/server/kafkaconsumer/trace"
	"github.com/sabariramc/goserverbase/v4/kafka"
	"github.com/sabariramc/goserverbase/v4/log"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
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

func (k *KafkaConsumerServer) GetMessageContext(msg *kafka.Message) context.Context {
	msgCtx := context.Background()
	msgCtx = k.GetContextWithCorrelation(msgCtx, k.GetCorrelationParams(msg.GetHeaders()))
	msgCtx = k.GetContextWithCustomerId(msgCtx, k.GetCustomerId(msg.GetHeaders()))
	span := trace.StartSpan(msgCtx, k.c.ServiceName, msg)
	return tracer.ContextWithSpan(msgCtx, span)
}
