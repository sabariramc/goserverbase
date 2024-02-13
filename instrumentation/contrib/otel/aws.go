package otel

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

func (t *tracerManager) AWS(cfg *aws.Config) {
	otelaws.AppendMiddlewares(&cfg.APIOptions)
}
