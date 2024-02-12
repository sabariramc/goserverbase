package opentelemetry

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/instrumentation/span"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// A MessageCarrier injects and extracts traces from a sarama.ProducerMessage.
type MessageCarrier struct {
	msg *kafka.Message
}

var _ interface {
	ddtrace.TextMapReader
	ddtrace.TextMapWriter
} = (*MessageCarrier)(nil)

// ForeachKey iterates over every header.
func (c MessageCarrier) ForeachKey(handler func(key, val string) error) error {
	for _, h := range c.msg.Headers {
		err := handler(string(h.Key), string(h.Value))
		if err != nil {
			return err
		}
	}
	return nil
}

// Set sets a header.
func (c MessageCarrier) Set(key, val string) {
	// ensure uniqueness of keys
	for i := 0; i < len(c.msg.Headers); i++ {
		if string(c.msg.Headers[i].Key) == key {
			c.msg.Headers = append(c.msg.Headers[:i], c.msg.Headers[i+1:]...)
			i--
		}
	}
	c.msg.Headers = append(c.msg.Headers, kafka.Header{
		Key:   key,
		Value: []byte(val),
	})
}

// NewMessageCarrier creates a new MessageCarrier.
func NewKafkaCarrier(msg *kafka.Message) MessageCarrier {
	return MessageCarrier{msg}
}

func (t *tracer) KafkaInject(ctx context.Context, msg *kafka.Message) {
	traceMsg := trace.NewMessageCarrier(msg)
	ddtrace.Inject(ctx, traceMsg)
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
