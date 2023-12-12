package csfle

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	cuaws "github.com/sabariramc/goserverbase/v4/aws"
	"github.com/sabariramc/goserverbase/v4/db/mongo"
	"github.com/sabariramc/goserverbase/v4/log"
)

type MasterKeyProvider interface {
	Name() string
	Credentials() map[string]map[string]interface{}
	DataKeyOpts() interface{}
}

type awsKMSDataKeyOpts struct {
	Region   string `bson:"region"`
	KeyARN   string `bson:"key"`
	Endpoint string `bson:"endpoint,omitempty"`
}

type AWSKMSProvider struct {
	credentials map[string]interface{}
	dataKeyOpts awsKMSDataKeyOpts
	name        string
}

func GetDefaultAWSProvider(ctx context.Context, logger *log.Logger, kmsARN string) (*AWSKMSProvider, error) {
	awsConfig := cuaws.GetDefaultAWSConfig()
	return GetAWSProvider(ctx, logger, awsConfig, kmsARN)
}

func GetAWSProvider(ctx context.Context, logger *log.Logger, awsConfig *aws.Config, kmsARN string) (provider *AWSKMSProvider, err error) {
	kms := cuaws.NewKMSClient(logger, cuaws.NewAWSKMSClientWithConfig(*awsConfig), kmsARN)
	_, err = kms.Encrypt(ctx, []byte("test"))
	if err != nil {
		return nil, fmt.Errorf("csfle.GetAWSProvider: error in kms test flight: %w", err)
	}
	cred, err := awsConfig.Copy().Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("CSFLE.GetAWSProvider: error in aws credential fetch: %w", err)
	}
	provider = CreateAWSProvider(cred.AccessKeyID, cred.SecretAccessKey, cred.SessionToken, awsConfig.Region)
	return provider, nil
}

func CreateAWSProvider(awsAccessKeyID, awsSecretAccessKey, sessionToken, awsKeyRegion string) *AWSKMSProvider {
	credentials := map[string]interface{}{
		"accessKeyId":     awsAccessKeyID,
		"secretAccessKey": awsSecretAccessKey,
	}
	if sessionToken != "" {
		credentials["sessionToken"] = sessionToken
	}
	return &AWSKMSProvider{
		credentials: credentials,
		name:        "aws",
		dataKeyOpts: awsKMSDataKeyOpts{
			Region: awsKeyRegion,
		},
	}
}

func GetDefaultAWSKMSProvider(ctx context.Context, logger *log.Logger, kmsARN string) (MasterKeyProvider, error) {
	provider, err := GetDefaultAWSProvider(ctx, logger, kmsARN)
	if err != nil {
		return nil, fmt.Errorf("csfle.GetDefaultAWSKMSProvider: %w", err)
	}
	provider.setARN(kmsARN)
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

func (a *AWSKMSProvider) setARN(awsKeyARN string) {
	a.dataKeyOpts.KeyARN = awsKeyARN
}

func SetEncryptionKey(ctx context.Context, logger *log.Logger, encryptionSchema *string, c mongo.Config, keyVaultNamespace, keyAltName string, kmsProvider MasterKeyProvider) error {
	schema := make(map[string]interface{})
	err := json.Unmarshal([]byte(*encryptionSchema), &schema)
	if err != nil {
		return fmt.Errorf("csfle.SetEncryptionKey: csfle schema unmarshal error: %w", err)
	}
	client, err := mongo.New(ctx, logger, c)
	if err != nil {
		return err
	}
	keyID, err := GetDataKey(ctx, client, keyVaultNamespace, keyAltName, kmsProvider)
	if err != nil {
		return err
	}
	encryptMetadataIn, ok := schema["encryptMetadata"]
	encryptionKeyID := []interface{}{
		map[string]interface{}{"$binary": map[string]interface{}{"base64": keyID, "subType": "04"}},
	}
	if ok {
		encryptMetadata, ok := encryptMetadataIn.(map[string]interface{})
		if !ok {
			errorMsg := "key `encryptMetadata` should be a compatible with `map[string]interface{}` in param `encryptionSchema`"
			logger.Error(ctx, errorMsg, schema)
			return fmt.Errorf("csfle.SetEncryptionKey: %v", errorMsg)
		}
		encryptMetadata["keyId"] = encryptionKeyID
	} else {
		schema["encryptMetadata"] = map[string]interface{}{
			"keyId": encryptionKeyID,
		}
	}
	blob, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("csfle.SetEncryptionKey: error in marshaling csfle scheme: %w", err)
	}
	*encryptionSchema = string(blob)
	return nil
}
