package otel

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

// AWS adds OpenTelemetry instrumentation middlewares to the given AWS SDK configuration.
// This allows tracing of AWS SDK calls.
func (t *tracerManager) AWS(cfg *aws.Config) {
	otelaws.AppendMiddlewares(&cfg.APIOptions)
}
