package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/sabariramc/goserverbase/log"
)

type SecretManager struct {
	_      struct{}
	client *secretsmanager.SecretsManager
	log    *log.Logger
}

type secretManagerCache struct {
	//TODO: maybe data field should be encrypted
	data secretsmanager.GetSecretValueOutput

	expireTime time.Time
}

var secretCache = make(map[string]secretManagerCache)

var defaultSecretManagerClient *secretsmanager.SecretsManager

func NewAWSSecretManagerClient(awsSession *session.Session) *secretsmanager.SecretsManager {
	client := secretsmanager.New(awsSession)
	return client
}

func GetDefaultSecretManagerClient(logger *log.Logger) *SecretManager {
	if defaultSecretManagerClient == nil {
		defaultSecretManagerClient = NewAWSSecretManagerClient(defaultAWSSession)
	}
	return NewSecretManagerClient(logger, defaultSecretManagerClient)
}

func NewSecretManagerClient(logger *log.Logger, client *secretsmanager.SecretsManager) *SecretManager {
	return &SecretManager{client: client, log: logger}
}

func (s *SecretManager) GetSecret(ctx context.Context, secretArn string) (map[string]interface{}, error) {
	secretCacheData, ok := secretCache[secretArn]
	if ok && time.Now().Before(secretCacheData.expireTime) {
		s.log.Info(ctx, "Secret fetched from cache", nil)
	} else {
		res, err := s.GetSecretNonCache(ctx, secretArn)
		if err != nil {
			return nil, fmt.Errorf("SecretManager.GetSecret: %w", err)
		}
		secretCacheData = secretManagerCache{expireTime: time.Now().Add(time.Minute * 15), data: *res}
		secretCache[secretArn] = secretCacheData
	}
	s.log.Debug(ctx, "Secret data", fmt.Sprint(secretCacheData))
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(*secretCacheData.data.SecretString), &data)
	if err != nil {
		s.log.Error(ctx, "Secret unmarshall error", err)
		s.log.Debug(ctx, "Secret data", secretCacheData.data.SecretString)
		return nil, fmt.Errorf("SecretManager.GetSecret: %w", err)
	}
	return data, nil
}

func (s *SecretManager) GetSecretNonCache(ctx context.Context, secretArn string) (*secretsmanager.GetSecretValueOutput, error) {
	req := &secretsmanager.GetSecretValueInput{SecretId: &secretArn}
	s.log.Debug(ctx, "Secret fetch request", req)
	res, err := s.client.GetSecretValueWithContext(ctx, req)
	if err != nil {
		s.log.Error(ctx, "Error in secret fetch", err)
		return nil, fmt.Errorf("SecretManager.GetSecretNonCache: %w", err)
	}
	s.log.Debug(ctx, "Secret fetch response", fmt.Sprint(res))
	return res, nil
}
