package ddtrace

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go-v2/aws"
)

func (t *tracer) AWS(cfg *aws.Config) {
	awstrace.AppendMiddleware(cfg)
}
