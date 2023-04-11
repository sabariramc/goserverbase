package csfle

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/aws"
	"github.com/sabariramc/goserverbase/db/mongo"
	"github.com/sabariramc/goserverbase/log"

	"github.com/aws/aws-sdk-go/aws/session"
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

func GetDefaultAWSProvider(ctx context.Context, logger *log.Logger) (*AWSKMSProvider, error) {
	session := aws.GetDefaultAWSSession()
	return GetAWSProvider(ctx, logger, session)
}

func GetAWSProvider(ctx context.Context, logger *log.Logger, session *session.Session) (provider *AWSKMSProvider, err error) {
	cred, err := session.Config.Credentials.Get()
	if err != nil {
		logger.Error(ctx, "AWS credential fetch", err)
		return
	}
	provider = CreateAWSProvider(cred.AccessKeyID, cred.SecretAccessKey, cred.SessionToken, *session.Config.Region)
	logger.Debug(ctx, "Mongo AWS KMS provider", provider)
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

func GetDefaultAWSKMSProvider(ctx context.Context, logger *log.Logger, KMSARN string) (MasterKeyProvider, error) {
	provider, err := GetDefaultAWSProvider(ctx, logger)
	if err != nil {
		return nil, fmt.Errorf("mongo.GetDefaultAWSKMSProvider : %w", err)
	}
	provider.setARN(KMSARN)
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
		logger.Error(ctx, "CSFLE Schema unmarshal error", err)
		return err
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
			return fmt.Errorf(errorMsg)
		}
		encryptMetadata["keyId"] = encryptionKeyID
	} else {
		schema["encryptMetadata"] = map[string]interface{}{
			"keyId": encryptionKeyID,
		}
	}
	blob, err := json.Marshal(schema)
	if err != nil {
		logger.Error(ctx, "CSFLE Schema marshal error", err)
		return err
	}
	*encryptionSchema = string(blob)
	return nil
}
