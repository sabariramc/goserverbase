package opentelemetry

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/instrumentation/span"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (t *tracer) NewSpanFromContext(ctx context.Context, operationName string) (context.Context, span.Span) {
	tr := otel.Tracer("")
	ctx, sp := tr.Start(ctx, operationName)
	return ctx, &otelSpan{Span: sp}
}

func (t *tracer) GetSpanFromContext(ctx context.Context) (span.Span, bool) {
	sp := trace.SpanFromContext(ctx)
	return &otelSpan{Span: sp}, sp.IsRecording()
}

type otelSpan struct {
	trace.Span
}

func (s *otelSpan) Finish() {
	s.End()
}

func (s *otelSpan) SetAttribute(name string, value string) {
	s.SetAttributes(attribute.String(name, value))
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
