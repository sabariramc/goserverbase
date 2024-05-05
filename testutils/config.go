package testutils

import (
	"crypto/tls"

	baseapp "github.com/sabariramc/goserverbase/v5/app"
	"github.com/sabariramc/goserverbase/v5/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v5/app/server/kafkaconsumer"
	"github.com/sabariramc/goserverbase/v5/db/mongo"
	"github.com/sabariramc/goserverbase/v5/db/mongo/csfle"
	"github.com/sabariramc/goserverbase/v5/kafka"
	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/log/logwriter"
	"github.com/sabariramc/goserverbase/v5/utils"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type AWSConfig struct {
	KMS_ARN      string
	SNS_ARN      string
	S3_BUCKET    string
	SECRET_ARN   string
	SQS_URL      string
	FIFO_SQS_URL string
}

type TestConfig struct {
	Logger          log.Config
	App             *baseapp.ServerConfig
	HTTP            *httpserver.HTTPServerConfig
	Kafka           *kafkaconsumer.KafkaConsumerServerConfig
	Mongo           *mongo.Config
	CSFLE           *csfle.Config
	AWS             *AWSConfig
	KafkaConsumer   kafka.KafkaConsumerConfig
	KafkaProducer   *kafka.KafkaProducerConfig
	KafkaTestTopic  string
	KafkaTestTopic2 string
	Graylog         logwriter.GraylogConfig
	TestURL1        string
	TestURL2        string
}

func NewConfig() *TestConfig {
	serviceName := utils.GetEnv("SERVICE_NAME", "go-base")
	kafkaBaseConfig := kafka.KafkaCredConfig{Brokers: []string{utils.GetEnv("KAFKA_BROKER", "")},
		SASLType: utils.GetEnv("SASL_TYPE", "NONE"),
	}
	appConfig := &baseapp.ServerConfig{
		ServiceName: serviceName,
	}
	consumer := kafka.KafkaConsumerConfig{
		KafkaCredConfig: &kafkaBaseConfig,
		GroupID:         utils.GetEnvMust("KAFKA_CONSUMER_ID"),
		MaxBuffer:       utils.GetEnvInt("KAFKA_CONSUMER_MAX_BUFFER", 1000),
		AutoCommit:      true,
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
			ServiceName:  serviceName,
			LogLevelName: utils.GetEnv("LOG_LEVEL", "INFO"),
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
		Kafka: &kafkaconsumer.KafkaConsumerServerConfig{
			ServerConfig:        *appConfig,
			KafkaConsumerConfig: consumer,
		},
		Mongo: mongo,
		CSFLE: &csfle.Config{
			Config:             mongo,
			CryptSharedLibPath: utils.GetEnv("CSFLE_CRYPT_SHARED_LIB_PATH", ""),
			KeyVaultNamespace:  utils.GetEnv("CSFLE_KEY_VAULT_NAMESPACE", ""),
		},
		AWS: &AWSConfig{
			KMS_ARN:      utils.GetEnv("KMS_ARN", ""),
			SNS_ARN:      utils.GetEnv("SNS_ARN", ""),
			S3_BUCKET:    utils.GetEnv("S3_BUCKET", ""),
			SECRET_ARN:   utils.GetEnv("SECRET_ARN", ""),
			SQS_URL:      utils.GetEnv("SQS_URL", ""),
			FIFO_SQS_URL: utils.GetEnv("FIFO_SQS_URL", ""),
		},
		KafkaProducer: &kafka.KafkaProducerConfig{
			KafkaCredConfig: &kafkaBaseConfig,
			Acknowledge:     -1,
			BatchMaxBuffer:  utils.GetEnvInt("KAFKA_PRODUCER_MAX_BUFFER", 1000),
			Async:           true,
		},
		KafkaConsumer:   consumer,
		KafkaTestTopic:  utils.GetEnvMust("KAFKA_TEST_TOPIC"),
		KafkaTestTopic2: utils.GetEnvMust("KAFKA_TEST_TOPIC_2"),
		Graylog: logwriter.GraylogConfig{
			Address: utils.GetEnv("GRAYLOG_URL", ""),
			Port:    utils.GetEnvInt("GRAYLOG_PORT", 12001),
		},
		TestURL1: utils.GetEnv("TEST_URL_1", ""),
		TestURL2: utils.GetEnv("TEST_URL_2", ""),
	}
}
