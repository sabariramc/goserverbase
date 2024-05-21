// Package testutils has common utility for testing the components
package testutils

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	base "github.com/sabariramc/goserverbase/v6/aws"
	"github.com/sabariramc/goserverbase/v6/utils"
	"github.com/sabariramc/snowflake"

	"github.com/joho/godotenv"
)

// SetAWSConfig sets the global AWS config and enable support for using localstack as proxy AWS services
func SetAWSConfig(tr base.Tracer) {
	cnf, _ := config.LoadDefaultConfig(context.TODO())
	if utils.GetEnv("AWS_PROVIDER", "") == "local" {
		var err error
		cnf, err = GetLocalStackConfig()
		if err != nil {
			log.Fatal(err)
		}
	}
	s, _ := snowflake.New()
	s.GenerateID()
	base.SetDefaultAWSConfig(cnf, tr)
}

func LoadEnv(path string) {
	if err := godotenv.Load(path); err != nil {
		fmt.Printf("Env file error - %v\n", err)
	}
}

func Initialize() {
	SetAWSConfig(nil)
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
