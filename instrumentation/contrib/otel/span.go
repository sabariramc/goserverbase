package otel

import (
	"context"
	"fmt"
	"time"

	"github.com/sabariramc/goserverbase/v5/instrumentation/span"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var spanKindMap = map[string]trace.SpanKind{
	span.SpanKindClient:   trace.SpanKindClient,
	span.SpanKindConsumer: trace.SpanKindConsumer,
	span.SpanKindInternal: trace.SpanKindInternal,
	span.SpanKindProducer: trace.SpanKindProducer,
	span.SpanKindServer:   trace.SpanKindServer,
}

func (t *tracerManager) NewSpanFromContext(ctx context.Context, operationName string, kind string, resourceName string) (context.Context, span.Span) {
	spanKind, ok := spanKindMap[kind]
	if !ok {
		spanKind = trace.SpanKindUnspecified
	}
	tr := t.Tracer("")
	opts := []trace.SpanStartOption{
		trace.WithAttributes(
			attribute.String("resource.name", resourceName),
		),
		trace.WithSpanKind(spanKind),
		trace.WithTimestamp(time.Now()),
	}
	ctx, sp := tr.Start(ctx, operationName, opts...)
	return ctx, &otelSpan{Span: sp}
}

func (t *tracerManager) GetSpanFromContext(ctx context.Context) (span.Span, bool) {
	sp := trace.SpanFromContext(ctx)
	return &otelSpan{Span: sp}, sp.IsRecording()
}

type otelSpan struct {
	trace.Span
}

func (s *otelSpan) Finish() {
	s.End()
}

func (s *otelSpan) SetAttribute(key string, value any) {
	var at attribute.KeyValue
	switch v := value.(type) {
	case string:
		at = attribute.String(key, v)
	case []string:
		at = attribute.StringSlice(key, v)
	case int:
		at = attribute.Int(key, v)
	case []int:
		at = attribute.IntSlice(key, v)
	case int64:
		at = attribute.Int64(key, v)
	case []int64:
		at = attribute.Int64Slice(key, v)
	case bool:
		at = attribute.Bool(key, v)
	case []bool:
		at = attribute.BoolSlice(key, v)
	case float64:
		at = attribute.Float64(key, v)
	case []float64:
		at = attribute.Float64Slice(key, v)
	case fmt.Stringer:
		at = attribute.Stringer(key, v)
	default:
		at = attribute.String(key, fmt.Sprintf("%v", v))
	}
	s.SetAttributes(at)
}

func (s *otelSpan) SetError(err error, stackTrace string) {
	opts := []trace.EventOption{}
	if stackTrace == "" {
		opts = append(opts, trace.WithStackTrace(true))
	}
	s.Span.RecordError(err, opts...)

}

func (s *otelSpan) SetStatus(statusCode int, description string) {
	if statusCode <= 299 {
		s.Span.SetStatus(codes.Ok, description)
	} else {
		s.Span.SetStatus(codes.Error, description)
	}
}
