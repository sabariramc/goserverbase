package ddtrace

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"github.com/segmentio/kafka-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// MessageCarrier implements datadog carrier interface for kafka.Message(github.com/segmentio/kafka-go)
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

// NewKafkaCarrier creates a new MessageCarrier.
func NewKafkaCarrier(msg *kafka.Message) MessageCarrier {
	return MessageCarrier{msg}
}

func (t *tracer) KafkaInject(ctx context.Context, msg *kafka.Message) {
	traceMsg := NewKafkaCarrier(msg)
	sp, _ := ddtrace.SpanFromContext(ctx)
	ddtrace.Inject(sp.Context(), traceMsg)
}

func (t *tracer) KafkaExtract(ctx context.Context, msg *kafka.Message) context.Context {
	spanCtx, err := ddtrace.Extract(NewKafkaCarrier(msg))
	if err != nil {
		return ctx
	}
	span := ddtrace.StartSpan("", ddtrace.ChildOf(spanCtx))
	return ddtrace.ContextWithSpan(ctx, span)
}

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
