package kafkaconsumer

import (
	"context"
	"fmt"
	"sync"

	e "errors"

	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/kafka"
)

// StartConsumer starts the Kafka consumer server. And starts background process for message poll, signal monitoring and set up cleanup steps when the server shutdowns
func (k *KafkaConsumerServer) StartConsumer(ctx context.Context) {
	corr := &correlation.CorrelationParam{CorrelationID: fmt.Sprintf("%v:KafkaConsumerServer", k.c.ServerConfig.ServiceName)}
	ctx, k.shutdown = context.WithCancel(correlation.GetContextWithCorrelationParam(ctx, corr))
	defer k.WaitForCompleteShutDown()
	k.shutdownWG.Add(1)
	defer k.shutdownWG.Wait()
	k.StartSignalMonitor(ctx)
	pollCtx, cancelPoll := context.WithCancel(correlation.GetContextWithCorrelationParam(context.Background(), corr))
	k.shutdownPoll = cancelPoll
	k.log.Notice(ctx, "Starting kafka consumer", nil)
	defer func() {
		if rec := recover(); rec != nil {
			defer k.shutdown()
			stackTrace, err := k.PanicRecovery(ctx, rec)
			k.log.Error(ctx, "Panic error", err)
			k.log.Error(ctx, "Panic stack tace", stackTrace)
		}
	}()
	go k.HealthCheckMonitor(pollCtx)
	k.Subscribe(ctx)
	var pollWg sync.WaitGroup
	defer pollWg.Wait()
	pollWg.Add(1)
	go func() {
		defer pollWg.Done()
		err := k.client.Poll(pollCtx, k.ch)
		if err != nil && !e.Is(err, context.Canceled) {
			k.log.Emergency(ctx, "Kafka consumer exited", err, fmt.Errorf("KafkaConsumerServer.StartConsumer: process exit: %w", err))
		}
	}()
	k.log.Notice(ctx, "Kafka consumer started", nil)
	k.requestWG.Add(1)
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-k.ch:
			if !ok {
				k.requestWG.Done()
				return
			}
			topicName := (*msg).Topic
			handler := k.handler[topicName]
			if handler == nil {
				k.log.Emergency(ctx, "missing handler for topic - "+topicName, nil, fmt.Errorf("KafkaConsumerServer.StartConsumer: missing handler for topic: %v", topicName))
			}
			emMsg := &kafka.Message{Message: msg}
			msgCtx := k.GetMessageContext(emMsg)
			k.requestWG.Add(1)
			k.ProcessEvent(msgCtx, emMsg, handler)
			k.requestWG.Done()
		}
	}
}
