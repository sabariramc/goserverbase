// Package instrumentation defines the interface for tracing requirements used across various packages.
package instrumentation

import (
	"github.com/sabariramc/goserverbase/v6/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v6/app/server/kafkaclient"
	"github.com/sabariramc/goserverbase/v6/aws"
	"github.com/sabariramc/goserverbase/v6/db/mongo"
	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/utils/retryhttp"
)

// Tracer defines an interface that consolidates tracing requirements for various components
// including AWS services, MongoDB operations, retryHTTP, Kafka wrapper,
// HTTP server handling, and general span operations.
type Tracer interface {
	aws.Tracer
	mongo.Tracer
	retryhttp.Tracer
	kafka.ProduceTracer
	kafkaclient.Tracer
	httpserver.Tracer
	span.SpanOp
}
