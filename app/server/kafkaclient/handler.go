package kafkaclient

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v2/kafka"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func (k *KafkaClient) AddHandler(ctx context.Context, topicName string, handler KafkaEventProcessor) {
	if handler == nil {
		k.Log.Emergency(ctx, "missing handler for topic - "+topicName, nil, fmt.Errorf("handler parameter cannot be nil"))
	}
	k.handler[topicName] = handler
}

func (k *KafkaClient) ProcessEvent(ctx context.Context, msg *kafka.Message, handler KafkaEventProcessor) {
	span, spanOk := tracer.SpanFromContext(ctx)
	defer func() {
		if rec := recover(); rec != nil {
			k.PanicRecovery(ctx, rec, msg)
			if spanOk {
				err, errOk := rec.(error)
				span.Finish(func(cfg *ddtrace.FinishConfig) {
					if errOk {
						cfg.Error = err
					} else {
						cfg.Error = fmt.Errorf("panic during execution")
					}
					cfg.StackFrames = 15
				})
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
			span.Finish(func(cfg *ddtrace.FinishConfig) {
				cfg.Error = err
			})
		}
	}

}
