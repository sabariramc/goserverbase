package mongo

import (
	"context"

	"sabariram.com/goserverbase/aws"
	"sabariram.com/goserverbase/log"

	"github.com/aws/aws-sdk-go/aws/session"
)

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

func GetDefaultAWSKMSProvider(ctx context.Context, logger *log.Logger, awsKeyARN string) (MasterKeyProvider, error) {
	provider, err := GetDefaultAWSProvider(ctx, logger)
	if err != nil {
		return nil, err
	}
	provider.updateKMSARN(awsKeyARN)
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

func (a *AWSKMSProvider) updateKMSARN(awsKeyARN string) {
	a.dataKeyOpts.KeyARN = awsKeyARN
}
