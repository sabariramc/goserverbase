package ddtrace

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"github.com/segmentio/kafka-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// MessageCarrier implements the Datadog carrier interface for kafka.Message (github.com/segmentio/kafka-go).
// It allows injecting and extracting trace context into/from Kafka messages.
type MessageCarrier struct {
	msg *kafka.Message
}

// ForeachKey iterates over every header in the carrier and invokes the given handler function for each key-value pair.
func (c MessageCarrier) ForeachKey(handler func(key, val string) error) error {
	for _, h := range c.msg.Headers {
		err := handler(string(h.Key), string(h.Value))
		if err != nil {
			return err
		}
	}
	return nil
}

// Set sets a header in the carrier.
func (c MessageCarrier) Set(key, val string) {
	// Ensure uniqueness of keys
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

// NewKafkaCarrier creates a new MessageCarrier for the given kafka.Message.
func NewKafkaCarrier(msg *kafka.Message) MessageCarrier {
	return MessageCarrier{msg}
}

// KafkaInject injects the trace context from the given context into the kafka.Message using Datadog's propagation mechanism.
func (t *tracer) KafkaInject(ctx context.Context, msg *kafka.Message) {
	traceMsg := NewKafkaCarrier(msg)
	sp, _ := ddtrace.SpanFromContext(ctx)
	ddtrace.Inject(sp.Context(), traceMsg)
}

// KafkaExtract extracts the trace context from the given kafka.Message using Datadog's propagation mechanism and returns a new context with the extracted trace context.
// If extraction fails or no trace context is found, it returns the original context.
func (t *tracer) KafkaExtract(ctx context.Context, msg *kafka.Message) context.Context {
	spanCtx, err := ddtrace.Extract(NewKafkaCarrier(msg))
	if err != nil {
		return ctx
	}
	span := ddtrace.StartSpan("", ddtrace.ChildOf(spanCtx))
	return ddtrace.ContextWithSpan(ctx, span)
}

// StartKafkaSpanFromMessage starts a new Kafka consumer span from the given kafka.Message.
// It creates a new Datadog span with the appropriate options and injects any available parent span context from the message headers.
func (t *tracer) StartKafkaSpanFromMessage(ctx context.Context, msg *kafka.Message) (context.Context, span.Span) {
	opts := []ddtrace.StartSpanOption{
		ddtrace.ResourceName(msg.Topic),
		ddtrace.SpanType(ext.SpanTypeMessageConsumer),
		ddtrace.Tag(ext.SpanKind, ext.SpanKindConsumer),
		ddtrace.Tag(ext.MessagingSystem, "kafka"),
		ddtrace.Measured(),
	}
	carrier := NewKafkaCarrier(msg)
	if spanCtx, err := ddtrace.Extract(carrier); err == nil {
		opts = append(opts, ddtrace.ChildOf(spanCtx))
	}
	sp, ctx := ddtrace.StartSpanFromContext(ctx, "kafka.consume", opts...)
	return ctx, &ddtraceSpan{Span: sp}
}
