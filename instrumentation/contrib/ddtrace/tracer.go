// Package ddtrace is the implementation of instrumentation.Tracer for Datadog.
package ddtrace

import (
	"github.com/sabariramc/goserverbase/v6/instrumentation"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// tracer is an empty struct implementing the instrumentation.Tracer interface for Datadog.
type tracer struct {
}

// Init initializes the Datadog tracer and returns an instance of instrumentation.Tracer.
// This function starts the Datadog tracer.
func Init() (instrumentation.Tracer, error) {
	ddtrace.Start()
	return &tracer{}, nil
}

// ShutDown stops the Datadog tracer. This function should be called to properly
// shut down the tracer and flush any remaining traces.
func ShutDown() {
	ddtrace.Stop()
}
