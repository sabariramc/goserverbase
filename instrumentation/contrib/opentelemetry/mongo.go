package opentelemetry

import (
	"go.mongodb.org/mongo-driver/event"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

func (t *tracer) MongoDB() *event.CommandMonitor {
	return otelmongo.NewMonitor()
}
