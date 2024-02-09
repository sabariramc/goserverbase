package opentelemetry

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/sabariramc/goserverbase/v5/instrumentation"
	"go.mongodb.org/mongo-driver/event"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

type tracer struct {
	*trace.TracerProvider
}

var global *tracer

func Init() (instrumentation.Tracer, error) {
	if global != nil {
		return global, nil
	}
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	bsp := trace.NewBatchSpanProcessor(exporter)
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	global = &tracer{
		TracerProvider: tp,
	}
	return global, nil
}

func ShutDown() {
	if global == nil {
		return
	}
	global.Shutdown(context.TODO())
}

func (t *tracer) AWS(cfg *aws.Config) {
	otelaws.AppendMiddlewares(&cfg.APIOptions)
}

func (t *tracer) MongoDB() *event.CommandMonitor {
	return otelmongo.NewMonitor()
}
