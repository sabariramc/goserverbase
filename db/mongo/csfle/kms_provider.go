package csfle

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	cuaws "github.com/sabariramc/goserverbase/v6/aws"
	"github.com/sabariramc/goserverbase/v6/log"
)

// MasterKeyProvider defines the interface for providers of master encryption keys.
type MasterKeyProvider interface {
	Name() string
	Credentials() map[string]map[string]interface{}
	DataKeyOpts() interface{}
}

// AWSDataKeyOpts holds the AWS-specific options for data key creation.
type AWSDataKeyOpts struct {
	Region   string `bson:"region"`             // AWS region where the KMS key is located.
	KeyARN   string `bson:"key"`                // Amazon Resource Name (ARN) of the KMS key.
	Endpoint string `bson:"endpoint,omitempty"` // Optional custom endpoint for the KMS service.
}

// AWSKMSProvider implements MasterKeyProvider for AWS KMS.
type AWSKMSProvider struct {
	credentials map[string]interface{}
	dataKeyOpts AWSDataKeyOpts
	name        string
}

// GetDefaultAWSProvider creates an AWSKMSProvider with default AWS configuration.
func GetDefaultAWSProvider(ctx context.Context, logger log.Log, kmsARN string) (*AWSKMSProvider, error) {
	awsConfig := cuaws.GetDefaultAWSConfig()
	return GetAWSProvider(ctx, logger, awsConfig, kmsARN)
}

// GetAWSProvider creates an AWSKMSProvider with the provided AWS configuration.
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

// NewAWSProvider initializes a new AWSKMSProvider with the given credentials and options.
func NewAWSProvider(credentials map[string]interface{}, opts AWSDataKeyOpts) *AWSKMSProvider {
	return &AWSKMSProvider{
		credentials: credentials,
		name:        "aws",
		dataKeyOpts: opts,
	}
}

// GetAWSMasterKeyProvider is a convenience function to get an AWS KMS provider with default settings.
func GetAWSMasterKeyProvider(ctx context.Context, logger log.Log, kmsARN string) (MasterKeyProvider, error) {
	provider, err := GetDefaultAWSProvider(ctx, logger, kmsARN)
	if err != nil {
		return nil, fmt.Errorf("csfle.GetDefaultAWSKMSProvider: %w", err)
	}
	return provider, nil
}

// Name returns the name of the AWS KMS provider.
func (a *AWSKMSProvider) Name() string {
	return a.name
}

// Credentials returns the AWS credentials needed for KMS operations.
func (a *AWSKMSProvider) Credentials() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{"aws": a.credentials}
}

// DataKeyOpts returns the options needed to create a data key using AWS KMS.
func (a *AWSKMSProvider) DataKeyOpts() interface{} {
	return a.dataKeyOpts
}
