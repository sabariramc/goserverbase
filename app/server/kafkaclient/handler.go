package kafkaclient

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v2/kafka"
)

func (k *KafkaClient) AddHandler(ctx context.Context, topicName string, handler KafkaEventProcessor) {
	if handler == nil {
		k.Log.Emergency(ctx, "missing handler for topic - "+topicName, nil, fmt.Errorf("handler parameter cannot be nil"))
	}
	k.handler[topicName] = handler
}

func (k *KafkaClient) ProcessEvent(ctx context.Context, msg *kafka.Message, handler KafkaEventProcessor) {
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
