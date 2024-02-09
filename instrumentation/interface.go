package instrumentation

import (
	"github.com/sabariramc/goserverbase/v5/aws"
	"github.com/sabariramc/goserverbase/v5/db/mongo"
)

type Tracer interface {
	aws.Tracer
	mongo.Tracer
}

type Span interface {
	SetTag(name string, value string)
	Finish()
}
