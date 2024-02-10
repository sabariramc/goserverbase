package opentelemetry

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

func (t *tracer) InitiateKafkaMessageSpanFromContext(ctx context.Context, msg *kafka.Message) (context.Context, span.Span) {
	msgCtx := t.KafkaExtract(ctx, msg)
	tr := otel.Tracer("")
	opts := []trace.SpanStartOption{
		trace.WithAttributes(
			attribute.String("messaging.kafka.topic", msg.Topic),
			attribute.Int("messaging.kafka.partition", msg.Partition),
			attribute.Int64("messaging.kafka.offset", msg.Offset),
			attribute.String("messaging.kafka.key", string(msg.Key)),
			attribute.Int64("messaging.kafka.key", msg.Time.UnixMilli()),
		),
		trace.WithNewRoot(),
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithTimestamp(msg.Time),
	}
	spanCtx, span := tr.Start(msgCtx, "kafka.consume", opts...)
	return spanCtx, &otelSpan{Span: span}
}
