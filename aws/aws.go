package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/sabariramc/goserverbase/v5/utils"
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

func GetLocalStackConfig() (aws.Config, error) {
	awsEndpoint := utils.GetEnvMust("AWS_ENDPOINT")
	awsRegion := utils.GetEnvMust("AWS_REGION")
	var err error
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if awsEndpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsEndpoint,
				SigningRegion: awsRegion,
			}, nil
		}
		// returning EndpointNotFoundError will allow the service to fallback to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
	cnf, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("GetLocalStackConfig: %w", err)
	}
	return cnf, nil
}
