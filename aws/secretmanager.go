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

// SecretManager provides methods to interact with AWS Secrets Manager service.
type SecretManager struct {
	_ struct{}
	*secretsmanager.Client
	log log.Log
}

// secretManagerCache stores cached secret data along with expiration time.
type secretManagerCache struct {
	data       secretsmanager.GetSecretValueOutput
	expireTime time.Time
}

// secretCache is a map to store cached secret data.
var secretCache = make(map[string]secretManagerCache)

// defaultSecretManagerClient is the default AWS Secrets Manager client.
var defaultSecretManagerClient *secretsmanager.Client

// NewSecretManagerClientWithSession creates a new Secrets Manager client with the provided AWS configuration.
func NewSecretManagerClientWithSession(awsConfig aws.Config) *secretsmanager.Client {
	client := secretsmanager.NewFromConfig(awsConfig)
	return client
}

// GetDefaultSecretManagerClient returns the default Secrets Manager client using the provided logger.
func GetDefaultSecretManagerClient(logger log.Log) *SecretManager {
	if defaultSecretManagerClient == nil {
		defaultSecretManagerClient = NewSecretManagerClientWithSession(*defaultAWSConfig)
	}
	return NewSecretManagerClient(logger, defaultSecretManagerClient)
}

// NewSecretManagerClient creates a new SecretManager instance with the provided logger and Secrets Manager client.
func NewSecretManagerClient(logger log.Log, client *secretsmanager.Client) *SecretManager {
	return &SecretManager{Client: client, log: logger.NewResourceLogger("SecretManager")}
}

// GetSecret retrieves the secret value associated with the provided secret ARN.
// It returns the secret data as a map[string]interface{}.
// If an error occurs during retrieval or unmarshalling, it returns an error.
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
