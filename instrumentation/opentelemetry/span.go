package opentelemetry

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/instrumentation/span"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type otelSpan struct {
	trace.Span
}

func (t *tracer) NewSpanFromContext(ctx context.Context, operationName string) (context.Context, span.Span) {
	tr := otel.Tracer("")

	ctx, sp := tr.Start(ctx, operationName)
	return ctx, &otelSpan{Span: sp}
}

func (s *otelSpan) Finish() {
	s.End()
}

func (s *otelSpan) SetTag(name string, value string) {
	s.SetAttributes(attribute.String(name, value))
}

func (s *otelSpan) SetError(err error) {}
