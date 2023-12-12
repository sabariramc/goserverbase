package kafkaconsumer

import (
	"context"

	"github.com/sabariramc/goserverbase/v4/app/server/kafkaconsumer/trace"
	"github.com/sabariramc/goserverbase/v4/kafka"
	"github.com/sabariramc/goserverbase/v4/log"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func (k *KafkaConsumerServer) GetCorrelationParams(headers map[string]string) *log.CorrelationParam {
	cr := &log.CorrelationParam{}
	cr.LoadFromHeader(headers)
	if cr.CorrelationId == "" {
		return log.GetDefaultCorrelationParam(k.c.ServiceName)
	}
	return cr
}

func (k *KafkaConsumerServer) GetCustomerId(headers map[string]string) *log.CustomerIdentifier {
	id := &log.CustomerIdentifier{}
	id.LoadFromHeader(headers)
	return id
}

func (k *KafkaConsumerServer) GetMessageContext(msg *kafka.Message) context.Context {
	msgCtx := context.Background()
	msgCtx = k.GetContextWithCorrelation(msgCtx, k.GetCorrelationParams(msg.GetHeaders()))
	msgCtx = k.GetContextWithCustomerId(msgCtx, k.GetCustomerId(msg.GetHeaders()))
	span := trace.StartSpan(msgCtx, k.c.ServiceName, msg)
	return tracer.ContextWithSpan(msgCtx, span)
}
