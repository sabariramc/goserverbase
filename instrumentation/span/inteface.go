package span

import "context"

type Span interface {
	SetAttribute(name string, value string)
	SetStatus(statusCode int, description string)
	SetError(err error, stackTrace string)
	Finish()
}

type SpanOp interface {
	NewSpanFromContext(ctx context.Context, operationName string) (context.Context, Span)
	GetSpanFromContext(ctx context.Context) (Span, bool)
}
