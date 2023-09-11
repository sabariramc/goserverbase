package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/sabariramc/goserverbase/v3/log"
)

type SecretManager struct {
	_ struct{}
	*secretsmanager.Client
	log *log.Logger
}

type secretManagerCache struct {
	data       secretsmanager.GetSecretValueOutput
	expireTime time.Time
}

var secretCache = make(map[string]secretManagerCache)

var defaultSecretManagerClient *secretsmanager.Client

func NewSecretManagerClientWithSession(awsConfig aws.Config) *secretsmanager.Client {
	client := secretsmanager.NewFromConfig(awsConfig)
	return client
}

func GetDefaultSecretManagerClient(logger *log.Logger) *SecretManager {
	if defaultSecretManagerClient == nil {
		defaultSecretManagerClient = NewSecretManagerClientWithSession(*defaultAWSConfig)
	}
	return NewSecretManagerClient(logger, defaultSecretManagerClient)
}

func NewSecretManagerClient(logger *log.Logger, client *secretsmanager.Client) *SecretManager {
	return &SecretManager{Client: client, log: logger.NewResourceLogger("SecretManager")}
}

func (s *SecretManager) GetSecret(ctx context.Context, secretArn string) (map[string]interface{}, error) {
	secretCacheData, ok := secretCache[secretArn]
	if ok && time.Now().Before(secretCacheData.expireTime) {
		s.log.Info(ctx, "Secret fetched from cache", nil)
	} else {
		req := &secretsmanager.GetSecretValueInput{SecretId: &secretArn}
		s.log.Debug(ctx, "Secret fetch request", req)
		res, err := s.Client.GetSecretValue(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("SecretManager.GetSecretNonCache: error in fetching secret: %w", err)
		}
		s.log.Debug(ctx, "Secret fetch response", res)
		if err != nil {
			return nil, fmt.Errorf("SecretManager.GetSecret: %w", err)
		}
		secretCacheData = secretManagerCache{expireTime: time.Now().Add(time.Minute * 15), data: *res}
		secretCache[secretArn] = secretCacheData
	}
	s.log.Debug(ctx, "Secret data", secretCacheData)
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(*secretCacheData.data.SecretString), &data)
	if err != nil {
		s.log.Error(ctx, "Secret un-marshall error", err)
		s.log.Debug(ctx, "Secret data", secretCacheData.data.SecretString)
		return nil, fmt.Errorf("SecretManager.GetSecret: error un-marshalling secret data: %w", err)
	}
	return data, nil
}
