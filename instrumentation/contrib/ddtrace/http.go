package ddtrace

import (
	"context"
	"net/http"
	"net/http/httptrace"

	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
)

// HTTPWrapTransport wraps the provided http.RoundTripper with Datadog's tracing instrumentation.
// It returns a new http.RoundTripper that traces HTTP requests and responses.
func (t *tracer) HTTPWrapTransport(rt http.RoundTripper) http.RoundTripper {
	return ddtrace.WrapRoundTripper(rt)
}

// HTTPRequestTrace returns a new httptrace.ClientTrace for tracing HTTP requests.
// This function is provided for compatibility with the instrumentation.Tracer interface and currently returns an empty httptrace.ClientTrace.
func (t *tracer) HTTPRequestTrace(ctx context.Context) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{}
}
