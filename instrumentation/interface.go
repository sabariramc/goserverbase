package instrumentation

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/aws"
	"github.com/sabariramc/goserverbase/v5/db/mongo"
	"github.com/sabariramc/goserverbase/v5/utils/httputil"
)

type Tracer interface {
	aws.Tracer
	mongo.Tracer
	httputil.Tracer
	NewSpanFromContext(ctx context.Context, operationName string) (context.Context, Span)
}

type Span interface {
	SetTag(name string, value string)
	SetError(err error)
	Finish()
}
