package testutils

import (
	"github.com/sabariramc/goserverbase/v6/db/mongo/csfle"
	"github.com/sabariramc/goserverbase/v6/utils"
)

type AWSResources struct {
	KMS     string
	SNS     string
	S3      string
	SECRET  string
	SQS     string
	FIFOSQS string
}

type TestConfig struct {
	CSFLE           *csfle.Config
	AWS             *AWSResources
	KafkaTestTopic  string
	KafkaTestTopic2 string
	TestURL1        string
	TestURL2        string
}

func NewConfig() *TestConfig {
	return &TestConfig{
		CSFLE: &csfle.Config{
			CryptSharedLibPath: utils.GetEnv("CSFLE_CRYPT_SHARED_LIB_PATH", ""),
			KeyVaultNamespace:  utils.GetEnv("CSFLE_KEY_VAULT_NAMESPACE", ""),
		},
		AWS: &AWSResources{
			KMS:     utils.GetEnv("KMS_ARN", ""),
			SNS:     utils.GetEnv("SNS_ARN", ""),
			S3:      utils.GetEnv("S3_BUCKET", ""),
			SECRET:  utils.GetEnv("SECRET_ARN", ""),
			SQS:     utils.GetEnv("SQS_URL", ""),
			FIFOSQS: utils.GetEnv("FIFO_SQS_URL", ""),
		},
		KafkaTestTopic:  utils.GetEnvMust("KAFKA__TOPIC"),
		KafkaTestTopic2: utils.GetEnvMust("KAFKA__TOPIC_2"),
		TestURL1:        utils.GetEnv("TEST_URL_1", ""),
		TestURL2:        utils.GetEnv("TEST_URL_2", ""),
	}
}
