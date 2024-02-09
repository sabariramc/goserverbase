package instrumentation

import (
	"context"

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
	kafkaconsumer.ConsumerTracer
	NewSpanFromContext(ctx context.Context, operationName string) (context.Context, span.Span)
}
