package otel

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/instrumentation/span"
	cKafka "github.com/sabariramc/goserverbase/v5/kafka"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	return val
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

func (t *tracerManager) KafkaInject(ctx context.Context, msg *kafka.Message) {
	otel.GetTextMapPropagator().Inject(ctx, NewKafkaCarrier(msg))
}

func (t *tracerManager) StartKafkaSpanFromMessage(ctx context.Context, msg *kafka.Message) (context.Context, span.Span) {
	msgCtx := otel.GetTextMapPropagator().Extract(ctx, NewKafkaCarrier(msg))
	tr := otel.Tracer("")
	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithTimestamp(msg.Time),
		trace.WithAttributes(attribute.String("messaging.system", "kafka"), attribute.String("resource.name", msg.Topic)),
	}
	spanCtx, span := tr.Start(msgCtx, "kafka.consume", opts...)
	return spanCtx, &otelSpan{Span: span}
}
