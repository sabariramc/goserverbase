package opentelemetry

import (
	"context"
	"net/http"
	"net/http/httptrace"

	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
)

func (t *tracer) HTTPWrapTransport(rt http.RoundTripper) http.RoundTripper {
	return ddtrace.WrapRoundTripper(rt)
}

func (t *tracer) HTTPRequestTrace(ctx context.Context) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{}
}
