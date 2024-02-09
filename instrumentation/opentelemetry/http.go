package opentelemetry

import (
	"context"
	"net/http"
	"net/http/httptrace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func (t *tracer) HTTPWrapTransport(rt http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(
		rt,
		otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
			return otelhttptrace.NewClientTrace(ctx)
		}),
	)
}

func (t *tracer) HTTPRequestTrace(ctx context.Context) *httptrace.ClientTrace {
	return otelhttptrace.NewClientTrace(ctx)
}
