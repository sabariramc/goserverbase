package ddtrace

import (
	"go.mongodb.org/mongo-driver/event"
	mongotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go.mongodb.org/mongo-driver/mongo"
)

func (t *tracer) MongoDB() *event.CommandMonitor {
	return mongotrace.NewMonitor()
}
