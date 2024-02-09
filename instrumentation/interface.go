package instrumentation

import (
	"github.com/sabariramc/goserverbase/v5/aws"
	"github.com/sabariramc/goserverbase/v5/db/mongo"
	"github.com/sabariramc/goserverbase/v5/utils/httputil"
)

type Tracer interface {
	aws.Tracer
	mongo.Tracer
	httputil.Tracer
}

type Span interface {
	SetTag(name string, value string)
	Finish()
}
