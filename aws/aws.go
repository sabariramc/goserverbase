package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go-v2/aws"
)

var defaultAWSConfig *aws.Config

func SetDefaultAWSConfig(defaultConfig aws.Config) {
	awstrace.AppendMiddleware(&defaultConfig)
	defaultAWSConfig = &defaultConfig
}

func GetDefaultAWSConfig() *aws.Config {
	return defaultAWSConfig
}
