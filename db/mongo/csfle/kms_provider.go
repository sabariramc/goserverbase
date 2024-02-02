package csfle

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	cuaws "github.com/sabariramc/goserverbase/v5/aws"
	"github.com/sabariramc/goserverbase/v5/log"
)

type MasterKeyProvider interface {
	Name() string
	Credentials() map[string]map[string]interface{}
	DataKeyOpts() interface{}
}

type AWSDataKeyOpts struct {
	Region   string `bson:"region"`
	KeyARN   string `bson:"key"`
	Endpoint string `bson:"endpoint,omitempty"`
}

type AWSKMSProvider struct {
	credentials map[string]interface{}
	dataKeyOpts AWSDataKeyOpts
	name        string
}

func GetDefaultAWSProvider(ctx context.Context, logger log.Log, kmsARN string) (*AWSKMSProvider, error) {
	awsConfig := cuaws.GetDefaultAWSConfig()
	return GetAWSProvider(ctx, logger, awsConfig, kmsARN)
}

func GetAWSProvider(ctx context.Context, logger log.Log, awsConfig *aws.Config, kmsARN string) (provider *AWSKMSProvider, err error) {
	kms := cuaws.NewKMSClient(logger, cuaws.NewKMSClientWithConfig(*awsConfig), kmsARN)
	_, err = kms.Encrypt(ctx, []byte("test"))
	if err != nil {
		return nil, fmt.Errorf("csfle.GetAWSProvider: error test flight of kms: %w", err)
	}
	cred, err := awsConfig.Copy().Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("CSFLE.GetAWSProvider: error fetching aws credential: %w", err)
	}
	credentials := map[string]interface{}{
		"accessKeyId":     cred.AccessKeyID,
		"secretAccessKey": cred.SecretAccessKey,
	}
	if cred.SessionToken != "" {
		credentials["sessionToken"] = cred.SessionToken
	}
	provider = NewAWSProvider(credentials, AWSDataKeyOpts{
		Region: awsConfig.Region,
		KeyARN: kmsARN,
	})
	return provider, nil
}

func NewAWSProvider(credentials map[string]interface{}, opts AWSDataKeyOpts) *AWSKMSProvider {
	return &AWSKMSProvider{
		credentials: credentials,
		name:        "aws",
		dataKeyOpts: opts,
	}
}

func GetAWSMasterKeyProvider(ctx context.Context, logger log.Log, kmsARN string) (MasterKeyProvider, error) {
	provider, err := GetDefaultAWSProvider(ctx, logger, kmsARN)
	if err != nil {
		return nil, fmt.Errorf("csfle.GetDefaultAWSKMSProvider: %w", err)
	}
	return provider, nil
}

func (a *AWSKMSProvider) Name() string {
	return a.name
}

func (a *AWSKMSProvider) Credentials() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{"aws": a.credentials}
}

func (a *AWSKMSProvider) DataKeyOpts() interface{} {
	return a.dataKeyOpts
}
