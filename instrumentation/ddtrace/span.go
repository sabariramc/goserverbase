package opentelemetry

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/instrumentation/span"
	"go.opentelemetry.io/otel/trace"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func (t *tracer) NewSpanFromContext(ctx context.Context, operationName string) (context.Context, span.Span) {
	 ddtrace.StartSpanFromContext(ctx, operationName)
}

func (t *tracer) GetSpanFromContext(ctx context.Context) (span.Span, bool) {
	sp := trace.SpanFromContext(ctx)
	return &otelSpan{Span: sp}, sp.IsRecording()
}

type otelSpan struct {
	ddtrace.Span
}

func (s *otelSpan) Finish() {
	s.Finish()
}

func (s *otelSpan) SetAttribute(name string, value string) {
	s.SetTag(name, value)
}

func (s *otelSpan) SetError(err error, stackTrace string) {
	s.SetError(err, stackTrace)
}

func (s *otelSpan) SetStatus(statusCode int, description string) {
	s.SetStatus(statusCode, description)
}
