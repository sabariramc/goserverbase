package opentelemetry

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/instrumentation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type span struct {
	trace.Span
}

func (t *tracer) NewSpanFromContext(ctx context.Context, operationName string) (context.Context, instrumentation.Span) {
	tr := otel.Tracer("")
	
	ctx, sp := tr.Start(ctx, operationName)
	return ctx, &span{Span: sp}
}

func (s *span) Finish() {
	s.End()
}

func (s *span) SetTag(name string, value string) {
	s.SetAttributes(attribute.String(name, value))
}

func (s *span) SetError(err error) {}
