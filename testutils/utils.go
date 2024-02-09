package testutils

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	base "github.com/sabariramc/goserverbase/v5/aws"
	"github.com/sabariramc/goserverbase/v5/utils"

	"github.com/joho/godotenv"
)

func setAWSSession() {
	cnf, _ := config.LoadDefaultConfig(context.TODO())
	if os.Getenv("AWS_PROVIDER") == "local" {
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
		cnf, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(awsRegion),
			config.WithEndpointResolverWithOptions(customResolver),
		)
		if err != nil {
			log.Fatalf("Cannot load the AWS configs: %s", err)
		}
		os.Setenv("AWS_PROVIDER", "local")
	}
	base.SetDefaultAWSConfig(cnf, nil)
}

func LoadEnv(path string) {
	if err := godotenv.Load(path); err != nil {
		fmt.Printf("Env file error - %v\n", err)
	}
}

func Initialize() {
	setAWSSession()
}
