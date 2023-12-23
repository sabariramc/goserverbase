package kafkaconsumer

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v4/kafka"
	ckafka "github.com/segmentio/kafka-go"
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

func (k *KafkaConsumerServer) Commit(ctx context.Context) error {
	return k.client.Commit(ctx)
}

func (k *KafkaConsumerServer) StoreMessage(ctx context.Context, msg *kafka.Message) {
	k.client.StoreMessage(ctx, msg.Message)
}

func (k *KafkaConsumerServer) Subscribe(ctx context.Context) {
	topicList := make([]string, 0, len(k.handler))
	for h := range k.handler {
		topicList = append(topicList, h)
	}
	ch := make(chan *ckafka.Message)
	k.ch = ch
	client, err := kafka.NewConsumer(ctx, k.log, k.c.KafkaConsumerConfig, topicList...)
	if err != nil {
		k.log.Emergency(ctx, "Error occurred during client creation", fmt.Errorf("KafkaConsumerServer.Subscribe: error creating kafka consumer: %w", err), map[string]any{
			"topicList": topicList,
			"config":    k.c.KafkaConsumerConfig,
		})
	}
	k.client = client
}
