// Package ddtrace is the implementation of instrumentation.Tracer for datadog
package ddtrace

import (
	"github.com/sabariramc/goserverbase/v5/instrumentation"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type tracer struct {
}

func Init() (instrumentation.Tracer, error) {
	ddtrace.Start()
	return &tracer{}, nil
}

func ShutDown() {
	ddtrace.Stop()
}
