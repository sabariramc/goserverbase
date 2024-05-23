package otel

import (
	"go.mongodb.org/mongo-driver/event"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

// MongoDB returns a new MongoDB command monitor for OpenTelemetry tracing.
// This monitor can be used to trace MongoDB commands in an application.
func (t *tracerManager) MongoDB() *event.CommandMonitor {
	return otelmongo.NewMonitor()
}
