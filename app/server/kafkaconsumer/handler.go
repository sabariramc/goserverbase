package kafkaconsumer

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v3/kafka"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func (k *KafkaConsumerServer) AddHandler(ctx context.Context, topicName string, handler KafkaEventProcessor) {
	if handler == nil {
		k.Log.Emergency(ctx, "missing handler for topic - "+topicName, nil, fmt.Errorf("handler parameter cannot be nil"))
	}
	k.handler[topicName] = handler
}

func (k *KafkaConsumerServer) ProcessEvent(ctx context.Context, msg *kafka.Message, handler KafkaEventProcessor) {
	span, spanOk := tracer.SpanFromContext(ctx)
	defer func() {
		if spanOk {
			span.Finish()
		}
	}()
	defer func() {
		if rec := recover(); rec != nil {
			k.PanicRecovery(ctx, rec, msg)
			if spanOk {
				err, errOk := rec.(error)
				if !errOk {
					err = fmt.Errorf("panic during execution")
				}
				span.SetTag(ext.Error, err)
			}
		}
	}()
	if spanOk {
		span.SetTag("topic.key", msg.GetKey())
	}
	err := handler(ctx, msg)
	if err != nil {
		statusCode, _ := k.ProcessError(ctx, "", err, msg)
		if spanOk && statusCode >= 500 {
			span.SetTag(ext.Error, err)
		}
	}

}
