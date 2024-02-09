package opentelemetry

import (
	"context"

	cKafka "github.com/sabariramc/goserverbase/v5/kafka"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

// HeaderCarrier adapts kafka.Message to satisfy the TextMapCarrier interface.

type KafkaCarrier struct {
	*cKafka.Message
	header map[string]string
}

func NewKafkaCarrier(msg *kafka.Message) *KafkaCarrier {
	cr := &KafkaCarrier{
		Message: &cKafka.Message{
			Message: msg,
		},
	}
	cr.header = cr.GetHeaders()
	return cr
}

// Get returns the value associated with the passed key.
func (kc *KafkaCarrier) Get(key string) string {
	val, ok := kc.header[key]
	if !ok {
		return ""
	}
	return string(val)
}

// Set stores the key-value pair.
func (kc *KafkaCarrier) Set(key string, value string) {
	kc.Message.Headers = append(kc.Message.Headers, kafka.Header{
		Key:   key,
		Value: []byte(value),
	})
}

// Keys lists the keys stored in this carrier.
func (kc *KafkaCarrier) Keys() []string {
	keys := make([]string, 0, len(kc.header))
	for k := range kc.header {
		keys = append(keys, k)
	}
	return keys
}

func (t *tracer) KafkaInject(ctx context.Context, msg *kafka.Message) {
	otel.GetTextMapPropagator().Inject(ctx, NewKafkaCarrier(msg))
}

func (t *tracer) KafkaExtract(ctx context.Context, msg *kafka.Message) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, NewKafkaCarrier(msg))
}
