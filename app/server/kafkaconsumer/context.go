package kafkaconsumer

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/instrumentation/span"
	"github.com/sabariramc/goserverbase/v5/kafka"
	"github.com/sabariramc/goserverbase/v5/log"
)

func (k *KafkaConsumerServer) GetCorrelationParams(headers map[string]string) *log.CorrelationParam {
	cr := &log.CorrelationParam{}
	cr.LoadFromHeader(headers)
	if cr.CorrelationId == "" {
		return log.GetDefaultCorrelationParam(k.c.ServiceName)
	}
	return cr
}

func (k *KafkaConsumerServer) GetCustomerID(headers map[string]string) *log.CustomerIdentifier {
	id := &log.CustomerIdentifier{}
	id.LoadFromHeader(headers)
	return id
}

func (k *KafkaConsumerServer) GetMessageContext(msg *kafka.Message) context.Context {
	msgCtx := context.Background()
	corr := k.GetCorrelationParams(msg.GetHeaders())
	identity := k.GetCustomerID(msg.GetHeaders())
	msgCtx = k.GetContextWithCorrelation(msgCtx, k.GetCorrelationParams(msg.GetHeaders()))
	msgCtx = k.GetContextWithCustomerId(msgCtx, identity)
	if k.tracer != nil {
		var span span.Span
		msgCtx, span = k.tracer.InitiateKafkaMessageSpanFromContext(msgCtx, msg.Message)
		span.SetAttribute("correlationId", corr.CorrelationId)
		data := identity.GetPayload()
		for key, value := range data {
			if value != "" {
				span.SetAttribute("customer."+key, value)
			}
		}
	}
	return msgCtx
}
