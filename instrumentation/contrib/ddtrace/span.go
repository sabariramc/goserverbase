package ddtrace

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// NewSpanFromContext creates a new span from the provided context with the given operation name, kind, and resource name.
// It returns the updated context and the created span.
func (t *tracer) NewSpanFromContext(ctx context.Context, operationName string, kind string, resourceName string) (context.Context, span.Span) {
	opts := []ddtrace.StartSpanOption{
		ddtrace.Tag(ext.SpanKind, kind),
		ddtrace.Tag(ext.MessagingSystem, "kafka"),
		ddtrace.Measured(),
	}
	sp, ctx := ddtrace.StartSpanFromContext(ctx, operationName, opts...)
	return ctx, &ddtraceSpan{Span: sp}
}

// GetSpanFromContext retrieves the span from the provided context.
// It returns the span and a boolean indicating whether the span was found in the context.
func (t *tracer) GetSpanFromContext(ctx context.Context) (span.Span, bool) {
	sp, spanOk := ddtrace.SpanFromContext(ctx)
	return &ddtraceSpan{Span: sp}, spanOk
}

// ddtraceSpan is a wrapper around ddtrace.Span implementing the span.Span interface.
type ddtraceSpan struct {
	ddtrace.Span
}

// Finish finishes the span.
func (s *ddtraceSpan) Finish() {
	s.Span.Finish()
}

// spanAttributeMap maps span attribute names to corresponding Datadog tag names.
var spanAttributeMap = map[string]string{
	span.HTTPStatusCode: ext.HTTPCode,
}

// SetAttribute sets the attribute with the given name and value for the span.
func (s *ddtraceSpan) SetAttribute(name string, value any) {
	if mapName, ok := spanAttributeMap[name]; ok {
		name = mapName
	}
	s.SetTag(name, value)
}

// SetError sets the error for the span.
func (s *ddtraceSpan) SetError(err error, stackTrace string) {
	s.Span.SetTag(ext.Error, err)
}

// SetStatus sets the status code and description for the span.
func (s *ddtraceSpan) SetStatus(statusCode int, description string) {
	s.Span.SetTag(ext.HTTPCode, statusCode)
}
