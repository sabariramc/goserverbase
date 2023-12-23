package kafkaconsumer

import (
	"context"
	"fmt"

	e "errors"

	"github.com/sabariramc/goserverbase/v4/kafka"
	"github.com/sabariramc/goserverbase/v4/log"
)

func (k *KafkaConsumerServer) StartConsumer(ctx context.Context) {
	corr := &log.CorrelationParam{CorrelationId: fmt.Sprintf("%v:KafkaConsumerServer", k.c.ServiceName)}
	ctx, k.shutdown = context.WithCancel(log.GetContextWithCorrelation(ctx, corr))
	k.StartSignalMonitor(ctx)
	pollCtx, cancelPoll := context.WithCancel(log.GetContextWithCorrelation(context.Background(), corr))
	k.shutdownPoll = cancelPoll
	k.log.Notice(pollCtx, "Starting kafka consumer", nil)
	defer func() {
		if rec := recover(); rec != nil {
			k.PanicRecovery(pollCtx, rec)
		}
	}()
	k.Subscribe(pollCtx)
	defer k.wg.Wait()
	defer cancelPoll()
	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		err := k.client.Poll(pollCtx, k.ch)
		if !e.Is(err, context.Canceled) {
			k.log.Emergency(pollCtx, "Kafka consumer exited", nil, fmt.Errorf("KafkaConsumerServer.StartConsumer: process exit: %w", err))
		}
	}()
	k.log.Notice(pollCtx, "Kafka consumer started", nil)
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-k.ch:
			if !ok {
				return
			}
			topicName := (*msg).Topic
			handler := k.handler[topicName]
			if handler == nil {
				k.log.Emergency(pollCtx, "missing handler for topic - "+topicName, nil, fmt.Errorf("KafkaConsumerServer.StartConsumer: missing handler for topic: %v", topicName))
			}
			emMsg := &kafka.Message{Message: msg}
			msgCtx := k.GetMessageContext(emMsg)
			k.ProcessEvent(msgCtx, emMsg, handler)
		}
	}
}
