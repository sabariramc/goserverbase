package kafkaclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"github.com/sabariramc/goserverbase/v6/kafka"
	ckafka "github.com/segmentio/kafka-go"
)

// AddHandler adds a handler for processing Kafka events for the specified topic.
func (k *KafkaClient) AddHandler(ctx context.Context, topicName string, handler KafkaEventProcessor) {
	if handler == nil {
		k.log.Emergency(ctx, "missing handler for topic - "+topicName, nil, fmt.Errorf("KafkaClient.AddHandler: handler parameter cannot be nil"))
	}
	if _, ok := k.handler[topicName]; ok {
		k.log.Emergency(ctx, "duplicate handler for topic - "+topicName, nil, fmt.Errorf("KafkaClient.AddHandler: handler for topic exist"))
	}
	k.handler[topicName] = handler
}

// ProcessEvent processes a Kafka message using the specified handler.
func (k *KafkaClient) ProcessEvent(ctx context.Context, msg *kafka.Message, handler KafkaEventProcessor) {
	span, spanOk := k.GetSpanFromContext(ctx)
	defer func() {
		if spanOk {
			span.Finish()
		}
	}()
	defer func() {
		if rec := recover(); rec != nil {
			stackTrace, err := k.PanicRecovery(ctx, rec)
			statusCode, _ := k.ProcessError(ctx, stackTrace, err)
			if spanOk {
				span.SetError(err, stackTrace)
				span.SetStatus(statusCode, http.StatusText(statusCode))
			}
		}
	}()
	err := handler(ctx, msg)
	if err != nil {
		statusCode, _ := k.ProcessError(ctx, "", err)
		if spanOk {
			span.SetError(err, "")
			span.SetStatus(statusCode, http.StatusText(statusCode))
		}
		return
	}
	if spanOk {
		span.SetStatus(http.StatusOK, http.StatusText(http.StatusOK))
	}
}

// Commit commits the current offset of the Kafka consumer.
func (k *KafkaClient) Commit(ctx context.Context) error {
	return k.client.Commit(ctx)
}

// StoreMessage stores the given Kafka message.
func (k *KafkaClient) StoreMessage(ctx context.Context, msg *kafka.Message) error {
	return k.client.StoreMessage(ctx, msg.Message)
}

// Subscribe subscribes to Kafka topics and starts consuming messages.
func (k *KafkaClient) Subscribe(ctx context.Context) {
	topicList := make([]string, 0, len(k.handler))
	for h := range k.handler {
		topicList = append(topicList, h)
	}
	ch := make(chan *ckafka.Message)
	k.ch = ch
	client, err := kafka.NewPoller(kafka.WithConsumerTracer(k.tracer), kafka.WithConsumerTopic(topicList))
	if err != nil {
		k.log.Emergency(ctx, "Error occurred during client creation", fmt.Errorf("KafkaClient.Subscribe: error creating kafka consumer: %w", err), map[string]any{
			"topicList": topicList,
			"config":    k.c.ConsumerConfig,
		})
	}
	k.client = client
}

// GetSpanFromContext retrieves the OpenTelemetry span from the given context.
func (k *KafkaClient) GetSpanFromContext(ctx context.Context) (span.Span, bool) {
	if k.tracer != nil {
		return k.tracer.GetSpanFromContext(ctx)
	}
	return nil, false
}
