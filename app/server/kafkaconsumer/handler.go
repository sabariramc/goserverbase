package kafkaconsumer

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v4/kafka"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func (k *KafkaConsumerServer) AddHandler(ctx context.Context, topicName string, handler KafkaEventProcessor) {
	if handler == nil {
		k.log.Emergency(ctx, "missing handler for topic - "+topicName, nil, fmt.Errorf("handler parameter cannot be nil"))
	}
	if _, ok := k.handler[topicName]; ok {
		k.log.Emergency(ctx, "duplicate handler for topic - "+topicName, nil, fmt.Errorf("handler for topic exist"))
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
	err := handler(ctx, msg)
	if err != nil {
		statusCode, _ := k.ProcessError(ctx, "", err, msg)
		if spanOk && statusCode >= 500 {
			span.SetTag(ext.Error, err)
		}
	}

}
