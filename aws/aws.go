package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
)

var defaultAWSConfig *aws.Config

func SetDefaultAWSConfig(defaultConfig aws.Config) {
	defaultAWSConfig = &defaultConfig
}

func GetDefaultAWSConfig() *aws.Config {
	return defaultAWSConfig
}
