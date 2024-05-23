// Package span defines the interface for tracing spans used in various packages.
package span

import "context"

// Span represents a tracing span with methods to set attributes, status, errors, and to finish the span.
type Span interface {
	SetAttribute(key string, value any)
	SetStatus(statusCode int, description string)
	SetError(err error, stackTrace string)
	Finish()
}

// SpanOp represents operations that can be performed with spans, including creating new spans and retrieving existing spans from context.
type SpanOp interface {
	NewSpanFromContext(ctx context.Context, operationName string, kind string, resourceName string) (context.Context, Span)
	GetSpanFromContext(ctx context.Context) (Span, bool)
}

// Constants representing different kinds of spans.
const (
	SpanKindServer   = "server"
	SpanKindClient   = "client"
	SpanKindConsumer = "consumer"
	SpanKindProducer = "producer"
	SpanKindInternal = "internal"
)

// Constants representing attribute keys for spans.
const (
	HTTPStatusCode = "http.response.status_code"
)
