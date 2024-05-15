package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/sabariramc/goserverbase/v6/log"
)

type SecretManager struct {
	_ struct{}
	*secretsmanager.Client
	log log.Log
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

func GetDefaultSecretManagerClient(logger log.Log) *SecretManager {
	if defaultSecretManagerClient == nil {
		defaultSecretManagerClient = NewSecretManagerClientWithSession(*defaultAWSConfig)
	}
	return NewSecretManagerClient(logger, defaultSecretManagerClient)
}

func NewSecretManagerClient(logger log.Log, client *secretsmanager.Client) *SecretManager {
	return &SecretManager{Client: client, log: logger.NewResourceLogger("SecretManager")}
}

func (s *SecretManager) GetSecret(ctx context.Context, secretArn string) (map[string]interface{}, error) {
	secretCacheData, ok := secretCache[secretArn]
	if ok && time.Now().Before(secretCacheData.expireTime) {
		s.log.Notice(ctx, "Secret fetched from cache", nil)
	} else {
		req := &secretsmanager.GetSecretValueInput{SecretId: &secretArn}
		res, err := s.Client.GetSecretValue(ctx, req)
		if err != nil {
			s.log.Error(ctx, "error fetching secret", err)
			return nil, fmt.Errorf("SecretManager.GetSecret: error fetching secret: %w", err)
		}
		secretCacheData = secretManagerCache{expireTime: time.Now().Add(time.Minute * 15), data: *res}
		secretCache[secretArn] = secretCacheData
	}
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(*secretCacheData.data.SecretString), &data)
	if err != nil {
		s.log.Error(ctx, "Secret un-marshall error", err)
		return nil, fmt.Errorf("SecretManager.GetSecret: error un-marshalling secret data: %w", err)
	}
	return data, nil
}
