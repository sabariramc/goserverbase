package span

import "context"

type Span interface {
	SetAttribute(key string, value any)
	SetStatus(statusCode int, description string)
	SetError(err error, stackTrace string)
	Finish()
}

type SpanOp interface {
	NewSpanFromContext(ctx context.Context, operationName string, kind string, resourceName string) (context.Context, Span)
	GetSpanFromContext(ctx context.Context) (Span, bool)
}

const (
	SpanKindServer = "server"

	SpanKindClient = "client"

	SpanKindConsumer = "consumer"

	SpanKindProducer = "producer"

	SpanKindInternal = "internal"
)
