package kafkaconsumer

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v4/kafka"
)

func (k *KafkaConsumerServer) AddHandler(ctx context.Context, topicName string, handler KafkaEventProcessor) {
	if handler == nil {
		k.log.Emergency(ctx, "missing handler for topic - "+topicName, nil, fmt.Errorf("KafkaConsumerServer.AddHandler: handler parameter cannot be nil"))
	}
	if _, ok := k.handler[topicName]; ok {
		k.log.Emergency(ctx, "duplicate handler for topic - "+topicName, nil, fmt.Errorf("KafkaConsumerServer.AddHandler: handler for topic exist"))
	}
	k.handler[topicName] = handler
}

func (k *KafkaConsumerServer) ProcessEvent(ctx context.Context, msg *kafka.Message, handler KafkaEventProcessor) {
	defer func() {
		if rec := recover(); rec != nil {
			k.PanicRecovery(ctx, rec)
		}
	}()
	err := handler(ctx, msg)
	if err != nil {
		k.ProcessError(ctx, "", err)
	}
}
