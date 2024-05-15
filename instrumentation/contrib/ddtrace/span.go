package ddtrace

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func (t *tracer) NewSpanFromContext(ctx context.Context, operationName string, kind string, resourceName string) (context.Context, span.Span) {
	opts := []ddtrace.StartSpanOption{
		ddtrace.Tag(ext.SpanKind, kind),
		ddtrace.Tag(ext.MessagingSystem, "kafka"),
		ddtrace.Measured(),
	}
	sp, ctx := ddtrace.StartSpanFromContext(ctx, operationName, opts...)
	return ctx, &ddtraceSpan{Span: sp}
}

func (t *tracer) GetSpanFromContext(ctx context.Context) (span.Span, bool) {
	sp, spanOk := ddtrace.SpanFromContext(ctx)
	return &ddtraceSpan{Span: sp}, spanOk
}

type ddtraceSpan struct {
	ddtrace.Span
}

func (s *ddtraceSpan) Finish() {
	s.Span.Finish()
}

var spanAttributeMap = map[string]string{
	span.HTTPStatusCode: ext.HTTPCode,
}

func (s *ddtraceSpan) SetAttribute(name string, value any) {
	if mapName, ok := spanAttributeMap[name]; ok {
		name = mapName
	}
	s.SetTag(name, value)
}

func (s *ddtraceSpan) SetError(err error, stackTrace string) {
	s.Span.SetTag(ext.Error, err)
}

func (s *ddtraceSpan) SetStatus(statusCode int, description string) {
	s.Span.SetTag(ext.HTTPCode, statusCode)
}
