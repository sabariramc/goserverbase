package mongo

import (
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

func GetDefaultAWSProvider() (*AWSKMSProvider, error) {
	session := aws.GetDefaultAWSSession()
	return GetAWSProvider(session, log.GetDefaultLogger())
}

func GetAWSProvider(session *session.Session, log *log.Log) (provider *AWSKMSProvider, err error) {
	cred, err := session.Config.Credentials.Get()
	if err != nil {
		log.Error("AWS credential fetch", err)
		return
	}
	provider = CreateAWSProvider(cred.AccessKeyID, cred.SecretAccessKey, cred.SessionToken, *session.Config.Region)
	log.Debug("Mongo AWS KMS provider", provider)
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

func GetDefaultAWSKMSProvider(awsKeyARN string) (MasterKeyProvider, error) {
	provider, err := GetDefaultAWSProvider()
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
