package kafkaconsumer

import (
	"context"

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
	msgCtx = k.GetContextWithCorrelation(msgCtx, k.GetCorrelationParams(msg.GetHeaders()))
	msgCtx = k.GetContextWithCustomerId(msgCtx, k.GetCustomerID(msg.GetHeaders()))
	if k.t != nil {
		msgCtx, _ = k.t.InitiateKafkaMessageSpanFromContext(msgCtx, msg.Message)
	}
	return msgCtx
}
