package kafkaconsumer

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/log"
)

func (k *KafkaConsumerServer) GetCorrelationParams(headers map[string]string) *log.CorrelationParam {
	cr := &log.CorrelationParam{}
	cr.ExtractFromHeader(headers)
	if cr.CorrelationID == "" {
		return log.GetDefaultCorrelationParam(k.c.ServiceName)
	}
	return cr
}

func (k *KafkaConsumerServer) GetCustomerID(headers map[string]string) *log.UserIdentifier {
	id := &log.UserIdentifier{}
	id.ExtractFromHeader(headers)
	return id
}

func (k *KafkaConsumerServer) GetMessageContext(msg *kafka.Message) context.Context {
	msgCtx := context.Background()
	corr := k.GetCorrelationParams(msg.GetHeaders())
	identity := k.GetCustomerID(msg.GetHeaders())
	msgCtx = log.GetContextWithCorrelationParam(msgCtx, k.GetCorrelationParams(msg.GetHeaders()))
	msgCtx = log.GetContextWithUserIdentifier(msgCtx, identity)
	if k.tracer != nil {
		var span span.Span
		msgCtx, span = k.tracer.StartKafkaSpanFromMessage(msgCtx, msg.Message)
		span.SetAttribute("correlationId", corr.CorrelationID)
		span.SetAttribute("messaging.kafka.topic", msg.Topic)
		span.SetAttribute("messaging.kafka.partition", msg.Partition)
		span.SetAttribute("messaging.kafka.offset", msg.Offset)
		span.SetAttribute("messaging.kafka.key", string(msg.Key))
		span.SetAttribute("messaging.kafka.timestamp", msg.Time.UnixMilli())
		data := identity.GetPayload()
		for key, value := range data {
			if value != "" {
				span.SetAttribute("user."+key, value)
			}
		}
	}
	return msgCtx
}
