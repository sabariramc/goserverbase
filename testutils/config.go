package testutils

import (
	"crypto/tls"

	baseapp "github.com/sabariramc/goserverbase/v6/app"
	"github.com/sabariramc/goserverbase/v6/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v6/app/server/kafkaconsumer"
	"github.com/sabariramc/goserverbase/v6/db/mongo"
	"github.com/sabariramc/goserverbase/v6/db/mongo/csfle"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/utils"
	"github.com/segmentio/kafka-go/sasl/plain"
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
	KafkaConsumer   kafka.ConsumerConfig
	KafkaProducer   *kafka.ProducerConfig
	KafkaTestTopic  string
	KafkaTestTopic2 string
	TestURL1        string
	TestURL2        string
}

func NewConfig() *TestConfig {
	serviceName := utils.GetEnv("SERVICE_NAME", "go-base")
	kafkaBaseConfig := kafka.CredConfig{Brokers: []string{utils.GetEnv("KAFKA_BROKER", "")},
		SASLType: utils.GetEnv("SASL_TYPE", "NONE"),
	}
	appConfig := &baseapp.ServerConfig{
		ServiceName: serviceName,
	}
	consumer := kafka.ConsumerConfig{
		CredConfig: &kafkaBaseConfig,
		GroupID:    utils.GetEnvMust("KAFKA_CONSUMER_ID"),
		MaxBuffer:  uint(utils.GetEnvInt("KAFKA_CONSUMER_MAX_BUFFER", 1000)),
		AutoCommit: true,
	}
	if kafkaBaseConfig.SASLType == "PLAIN" {
		kafkaBaseConfig.SASLMechanism = &plain.Mechanism{
			Username: utils.GetEnv("KAFKA_USERNAME", ""),
			Password: utils.GetEnv("KAFKA_PASSWORD", ""),
		}
		kafkaBaseConfig.TLSConfig = &tls.Config{ //For Confluent kafka
			MinVersion: tls.VersionTLS12,
		}
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
			ServerConfig:   *appConfig,
			ConsumerConfig: consumer,
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
		KafkaProducer: &kafka.ProducerConfig{
			CredConfig:   &kafkaBaseConfig,
			RequiredAcks: -1,
			MaxBuffer:    utils.GetEnvInt("KAFKA_PRODUCER_MAX_BUFFER", 1000),
			Async:        true,
		},
		KafkaConsumer:   consumer,
		KafkaTestTopic:  utils.GetEnvMust("KAFKA_TEST_TOPIC"),
		KafkaTestTopic2: utils.GetEnvMust("KAFKA_TEST_TOPIC_2"),
		TestURL1:        utils.GetEnv("TEST_URL_1", ""),
		TestURL2:        utils.GetEnv("TEST_URL_2", ""),
	}
}
