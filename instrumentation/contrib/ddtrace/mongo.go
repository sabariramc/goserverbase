package ddtrace

import (
	"go.mongodb.org/mongo-driver/event"
	mongotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go.mongodb.org/mongo-driver/mongo"
)

// MongoDB returns a command monitor for MongoDB tracing using Datadog.
// This monitor can be used to trace MongoDB commands and operations.
func (t *tracer) MongoDB() *event.CommandMonitor {
	return mongotrace.NewMonitor()
}
