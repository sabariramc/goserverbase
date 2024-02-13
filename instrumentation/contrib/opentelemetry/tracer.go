package opentelemetry

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/instrumentation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
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
	tp, err := newTraceProvider()
	if err != nil {
		return nil, err
	}
	meterProvider, err := newMeterProvider()
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(newPropagator())
	otel.SetMeterProvider(meterProvider)
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

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider() (*trace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(context.Background())
	if err != nil {
		return nil, err
	}
	bsp := trace.NewBatchSpanProcessor(exporter)
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithSpanProcessor(bsp),
	)
	return tp, nil
}

func newMeterProvider() (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(context.Background())
	if err != nil {
		return nil, err
	}
	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(metricExporter)))
	return meterProvider, nil
}
