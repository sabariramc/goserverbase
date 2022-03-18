package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"sabariram.com/goserverbase/log"
)

type SecretManager struct {
	_      struct{}
	client *secretsmanager.SecretsManager
	log    *log.Log
	ctx    context.Context
}

type secretManagerCache struct {
	//TODO: maybe data field should be encrypted
	data secretsmanager.GetSecretValueOutput

	expireTime time.Time
}

var secretCache = make(map[string]secretManagerCache)

var defaultSecretManagerClient *secretsmanager.SecretsManager

func GetAWSSecretManagerClient(awsSession *session.Session) *secretsmanager.SecretsManager {
	client := secretsmanager.New(awsSession)
	return client
}

func GetDefaultSecretManagerClient(ctx context.Context) *SecretManager {
	if defaultSecretManagerClient == nil {
		defaultSecretManagerClient = GetAWSSecretManagerClient(defaultAWSSession)
	}
	return GetSecretManagerClient(ctx, defaultSecretManagerClient)
}

func GetSecretManagerClient(ctx context.Context, client *secretsmanager.SecretsManager) *SecretManager {
	return &SecretManager{client: client, log: log.GetDefaultLogger(), ctx: ctx}
}

func (s *SecretManager) GetSecret(secretArn string) (map[string]interface{}, error) {
	secretCacheData, ok := secretCache[secretArn]
	if ok && time.Now().Before(secretCacheData.expireTime) {
		s.log.Info("Secret fetched from cache", nil)
	} else {
		res, err := s.GetSecretNonCache(secretArn)
		if err != nil {
			return nil, err
		}
		secretCacheData = secretManagerCache{expireTime: time.Now().Add(time.Minute * 15), data: *res}
		secretCache[secretArn] = secretCacheData
	}
	s.log.Debug("Secret data", fmt.Sprint(secretCacheData))
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(*secretCacheData.data.SecretString), &data)
	if err != nil {
		s.log.Error("Secret unmarshall error", err)
		s.log.Debug("Secret data", secretCacheData.data.SecretString)
		return nil, err
	}
	return data, nil
}

func (s *SecretManager) GetSecretNonCache(secretArn string) (*secretsmanager.GetSecretValueOutput, error) {
	req := &secretsmanager.GetSecretValueInput{SecretId: &secretArn}
	s.log.Debug("Secret fetch request", req)
	res, err := s.client.GetSecretValueWithContext(s.ctx, req)
	if err != nil {
		s.log.Error("Error in secret fetch", err)
		return nil, err
	}
	s.log.Debug("Secret fetch response", fmt.Sprint(res))
	return res, nil
}
