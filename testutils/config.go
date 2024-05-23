package testutils

import (
	baseapp "github.com/sabariramc/goserverbase/v6/app"
	"github.com/sabariramc/goserverbase/v6/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v6/app/server/kafkaconsumer"
	"github.com/sabariramc/goserverbase/v6/db/mongo"
	"github.com/sabariramc/goserverbase/v6/db/mongo/csfle"
	"github.com/sabariramc/goserverbase/v6/log"
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
	Logger          log.Config
	App             *baseapp.ServerConfig
	HTTP            *httpserver.HTTPServerConfig
	Kafka           *kafkaconsumer.Config
	Mongo           *mongo.Config
	CSFLE           *csfle.Config
	AWS             *AWSResources
	KafkaTestTopic  string
	KafkaTestTopic2 string
	TestURL1        string
	TestURL2        string
}

func NewConfig() *TestConfig {
	serviceName := utils.GetEnv("SERVICE_NAME", "go-base")
	appConfig := &baseapp.ServerConfig{
		ServiceName: serviceName,
	}

	mongo := &mongo.Config{
		ConnectionString: utils.GetEnv("MONGO_URL", "mongodb://localhost:60001"),
	}
	return &TestConfig{
		Logger: log.Config{
			ServiceName: serviceName,
		},
		App: appConfig,
		HTTP: &httpserver.HTTPServerConfig{
			ServerConfig: *appConfig,
			Log:          &httpserver.LogConfig{AuthHeaderKeyList: utils.GetEnvAsSlice("AUTH_HEADER_LIST", []string{}, ";")},
			Host:         "0.0.0.0",
			Port:         utils.GetEnv("APP_PORT", "8080"),
			DocumentationConfig: httpserver.DocumentationConfig{
				DocHost:           utils.GetEnv("DOC_HOST", "localhost:8080"),
				SwaggerRootFolder: utils.GetEnv("DOC_ROOT_FOLDER", ""),
			},
			HTTP2Config: &httpserver.HTTP2Config{
				PublicKeyPath:  utils.GetEnv("HTTP2_PUBLIC_KEY", ""),
				PrivateKeyPath: utils.GetEnv("HTTP2_PRIVATE_KEY", ""),
			},
		},
		Kafka: &kafkaconsumer.Config{
			ServerConfig: *appConfig,
		},
		Mongo: mongo,
		CSFLE: &csfle.Config{
			Config:             mongo,
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
