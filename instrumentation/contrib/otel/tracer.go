// Package otel is the implementation of instrumentation.Tracer for opentelemetry
package otel

import (
	"context"

	"github.com/sabariramc/goserverbase/v5/instrumentation"
	"github.com/sabariramc/goserverbase/v5/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type tracerManager struct {
	*sdkTrace.TracerProvider
	env     string
	version string
}

type tracer struct {
	trace.Tracer
	env     string
	version string
}

func (t *tracerManager) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	tr := t.TracerProvider.Tracer(name, opts...)
	return &tracer{Tracer: tr, env: t.env, version: t.version}
}

func (t *tracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	opts = append(opts, trace.WithAttributes(attribute.String("env", t.env), attribute.String("version", t.version)))
	return t.Tracer.Start(ctx, spanName, opts...)
}

var global *tracerManager

func Init() (instrumentation.Tracer, error) {
	if global != nil {
		return global, nil
	}
	tp, err := newTraceProvider()
	if err != nil {
		return nil, err
	}
	global = &tracerManager{
		TracerProvider: tp,
		env:            utils.GetEnv("OTEL_ENV", ""),
		version:        utils.GetEnv("OTEL_SERVICE_VERSION", ""),
	}
	meterProvider, err := newMeterProvider()
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(global)
	otel.SetTextMapPropagator(newPropagator())
	otel.SetMeterProvider(meterProvider)
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

func newTraceProvider() (*sdkTrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(context.Background())
	if err != nil {
		return nil, err
	}
	bsp := sdkTrace.NewBatchSpanProcessor(exporter)
	tp := sdkTrace.NewTracerProvider(
		sdkTrace.WithSampler(sdkTrace.AlwaysSample()),
		sdkTrace.WithBatcher(exporter),
		sdkTrace.WithSpanProcessor(bsp),
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
