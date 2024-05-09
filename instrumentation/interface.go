// Package instrumentation define the interface for the tracing requirement of the package
package instrumentation

import (
	"github.com/sabariramc/goserverbase/v5/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v5/app/server/kafkaconsumer"
	"github.com/sabariramc/goserverbase/v5/aws"
	"github.com/sabariramc/goserverbase/v5/db/mongo"
	"github.com/sabariramc/goserverbase/v5/instrumentation/span"
	"github.com/sabariramc/goserverbase/v5/kafka"
	"github.com/sabariramc/goserverbase/v5/utils/httputil"
)

type Tracer interface {
	aws.Tracer
	mongo.Tracer
	httputil.Tracer
	kafka.ProduceTracer
	kafkaconsumer.Tracer
	httpserver.Tracer
	span.SpanOp
}
