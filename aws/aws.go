// Package aws provides a simplified interface for various AWS services.
package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
)

var defaultAWSConfig *aws.Config

// Tracer is an interface for tracing AWS configurations.
type Tracer interface {
	AWS(*aws.Config)
}

// SetDefaultAWSConfig sets the default AWS configuration and applies tracing if provided.
func SetDefaultAWSConfig(defaultConfig aws.Config, t Tracer) {
	if t != nil {
		t.AWS(&defaultConfig)
	}
	defaultAWSConfig = &defaultConfig
}

// GetDefaultAWSConfig retrieves the default AWS configuration.
func GetDefaultAWSConfig() *aws.Config {
	return defaultAWSConfig
}
