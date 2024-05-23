package otel

import (
	"context"
	"net/http"
	"net/http/httptrace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// HTTPWrapTransport wraps the given [http.RoundTripper] with OpenTelemetry instrumentation.
// This allows tracing HTTP requests and responses.
func (t *tracerManager) HTTPWrapTransport(rt http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(
		rt,
		otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
			return otelhttptrace.NewClientTrace(ctx)
		}),
	)
}

// HTTPRequestTrace creates a new [httptrace.ClientTrace] for tracing HTTP requests using OpenTelemetry.
// This can be used to trace the lifecycle of an HTTP request.
func (t *tracerManager) HTTPRequestTrace(ctx context.Context) *httptrace.ClientTrace {
	return otelhttptrace.NewClientTrace(ctx)
}
