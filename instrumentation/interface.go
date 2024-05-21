// Package instrumentation define the interface for the tracing requirement for the rest of the package
package instrumentation

import (
	"github.com/sabariramc/goserverbase/v6/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v6/app/server/kafkaconsumer"
	"github.com/sabariramc/goserverbase/v6/aws"
	"github.com/sabariramc/goserverbase/v6/db/mongo"
	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/utils/retryhttp"
)

type Tracer interface {
	aws.Tracer
	mongo.Tracer
	retryhttp.Tracer
	kafka.ProduceTracer
	kafkaconsumer.Tracer
	httpserver.Tracer
	span.SpanOp
}
