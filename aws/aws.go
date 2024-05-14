// Package aws provides simplified interface for different AWS services
package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
)

var defaultAWSConfig *aws.Config

type Tracer interface {
	AWS(*aws.Config)
}

func SetDefaultAWSConfig(defaultConfig aws.Config, t Tracer) {
	if t != nil {
		t.AWS(&defaultConfig)
	}
	defaultAWSConfig = &defaultConfig
}

func GetDefaultAWSConfig() *aws.Config {
	return defaultAWSConfig
}
