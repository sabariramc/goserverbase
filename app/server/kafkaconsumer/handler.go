package kafkaconsumer

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v3/kafka"
)

func (k *KafkaConsumerServer) AddHandler(ctx context.Context, topicName string, handler KafkaEventProcessor) {
	if handler == nil {
		k.Log.Emergency(ctx, "missing handler for topic - "+topicName, nil, fmt.Errorf("handler parameter cannot be nil"))
	}
	if _, ok := k.handler[topicName]; ok {
		k.Log.Emergency(ctx, "duplicate handler for topic - "+topicName, nil, fmt.Errorf("handler for topic exist"))
	}
	k.handler[topicName] = handler
}

func (k *KafkaConsumerServer) ProcessEvent(ctx context.Context, msg *kafka.Message, handler KafkaEventProcessor) {
	defer func() {
		if rec := recover(); rec != nil {
			k.PanicRecovery(ctx, rec, msg)
		}
	}()
	err := handler(ctx, msg)
	if err != nil {
		k.ProcessError(ctx, "", err, msg)
	}
}