package trace

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/kafka"
	ktrace "github.com/sabariramc/goserverbase/v5/kafka/api/trace"
	"github.com/sabariramc/goserverbase/v5/log"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func StartSpan(ctx context.Context, serviceName string, msg *kafka.Message) context.Context {
	corr := log.GetCorrelationParam(ctx)
	opts := []tracer.StartSpanOption{
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
		tracer.Tag("correlationId", corr.CorrelationId),
		tracer.Measured(),
	}
	// kafka supports headers, so try to extract a span context
	carrier := ktrace.NewMessageCarrier(msg.Message)
	if spanctx, err := tracer.Extract(carrier); err == nil {
		opts = append(opts, tracer.ChildOf(spanctx))
	}
	_, ctx = tracer.StartSpanFromContext(ctx, "kafka.consume", opts...)
	return ctx
}
