package span

import "context"

type Span interface {
	SetTag(name string, value string)
	SetStatus(statusCode int, description string)
	SetError(stackTrace string, err error)
	Finish()
}

type SpanOp interface {
	NewSpanFromContext(ctx context.Context, operationName string) (context.Context, Span)
	GetSpanFromContext(ctx context.Context) (Span, bool)
}
