package trace

import (
	"context"

	"github.com/sabariramc/goserverbase/v3/kafka"
	ktrace "github.com/sabariramc/goserverbase/v3/kafka/api/trace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func StartSpan(ctx context.Context, serviceName string, msg *kafka.Message) ddtrace.Span {
	opts := []tracer.StartSpanOption{
		tracer.ServiceName(serviceName),
		tracer.ResourceName(msg.Topic),
		tracer.SpanType(ext.SpanTypeMessageConsumer),
		tracer.Tag("messaging.kafka.topic", msg.Message.Topic),
		tracer.Tag(ext.MessagingKafkaPartition, msg.Message.Partition),
		tracer.Tag("messaging.kafka.offset", msg.Message.Offset),
		tracer.Tag("messaging.kafka.key", msg.GetKey()),
		tracer.Tag("messaging.kafka.timestamp", msg.Message.Time.UnixMilli()),
		tracer.Tag(ext.Component, "kafka"),
		tracer.Tag(ext.SpanKind, ext.SpanKindConsumer),
		tracer.Tag(ext.MessagingSystem, "kafka"),
		tracer.Measured(),
	}
	// kafka supports headers, so try to extract a span context
	carrier := ktrace.NewMessageCarrier(msg.Message)
	if spanctx, err := tracer.Extract(carrier); err == nil {
		opts = append(opts, tracer.ChildOf(spanctx))
	}
	span, _ := tracer.StartSpanFromContext(ctx, "kafka.consume", opts...)
	// reinject the span context so consumers can pick it up
	tracer.Inject(span.Context(), carrier)
	return span
}
