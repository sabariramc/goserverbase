package testutils

import (
	"crypto/tls"

	baseapp "github.com/sabariramc/goserverbase/v4/app"
	"github.com/sabariramc/goserverbase/v4/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v4/app/server/kafkaconsumer"
	"github.com/sabariramc/goserverbase/v4/db/mongo"
	"github.com/sabariramc/goserverbase/v4/kafka"
	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/log/logwriter"
	"github.com/sabariramc/goserverbase/v4/utils"
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
	Logger              *log.Config
	App                 *baseapp.ServerConfig
	Http                *httpserver.HttpServerConfig
	Kafka               *kafkaconsumer.KafkaConsumerServerConfig
	KafkaSASLCredential *kafka.SASLCredential
	Mongo               *mongo.Config
	AWS                 *AWSConfig
	KafkaConsumer       *kafka.KafkaConsumerConfig
	KafkaProducer       *kafka.KafkaProducerConfig
	KafkaTestTopic      string
	KafkaTestTopic2     string
	KafkaHTTPProxyURL   string
	Graylog             logwriter.GraylogConfig
}

func (t *TestConfig) GetLoggerConfig() *log.Config {
	return t.Logger
}
func (t *TestConfig) GetAppConfig() *baseapp.ServerConfig {
	return t.App
}

func NewConfig() *TestConfig {
	serviceName := utils.GetEnv("SERVICE_NAME", "go-base")
	kafkaBaseConfig := kafka.KafkaCredConfig{Brokers: []string{utils.GetEnv("KAFKA_BROKER", "")},
		ServiceName: serviceName,
	}
	appConfig := &baseapp.ServerConfig{
		ServiceName: serviceName,
	}
	consumer := &kafka.KafkaConsumerConfig{
		KafkaCredConfig: &kafkaBaseConfig,
		GroupID:         utils.GetEnvMust("KAFKA_CONSUMER_ID"),
		MaxBuffer:       uint64(utils.GetEnvInt("KAFKA_CONSUMER_MAX_BUFFER", 1000)),
		AutoCommit:      true,
	}
	saslConfig := &kafka.SASLCredential{
		SASLMechanism: utils.GetEnvMust("SASL_MECHANISM"),
	}
	if saslConfig.SASLMechanism == "PLAIN" {
		kafkaBaseConfig.SASLMechanism = &plain.Mechanism{
			Username: utils.GetEnv("KAFKA_USERNAME", ""),
			Password: utils.GetEnv("KAFKA_PASSWORD", ""),
		}
		kafkaBaseConfig.TLSConfig = &tls.Config{ //For Confluent kafka
			MinVersion: tls.VersionTLS12,
		}
	}

	return &TestConfig{
		Logger: &log.Config{
			HostParams: log.HostParams{
				Version:     utils.GetEnv("LOG_VERSION", "1.1"),
				Host:        utils.GetEnv("HOST", utils.GetHostName()),
				ServiceName: serviceName,
			},
			LogLevelName: utils.GetEnv("LOG_LEVEL", "INFO"),
			BufferSize:   utils.GetEnvInt("LOG_BUFFER_SIZE", 1),
		},
		App: appConfig,
		Http: &httpserver.HttpServerConfig{
			ServerConfig: appConfig,
			Log:          &httpserver.LogConfig{AuthHeaderKeyList: utils.GetEnvAsSlice("AUTH_HEADER_LIST", []string{}, ";")},
			Host:         "0.0.0.0",
			Port:         utils.GetEnv("APP_PORT", "8080"),
		},
		Kafka: &kafkaconsumer.KafkaConsumerServerConfig{
			ServerConfig:        appConfig,
			KafkaConsumerConfig: consumer,
		},
		KafkaSASLCredential: saslConfig,
		Mongo: &mongo.Config{
			ConnectionString:  utils.GetEnv("MONGO_URL", "mongodb://localhost:60001"),
			MinConnectionPool: uint64(utils.GetEnvInt("MONGO_MIN_CONNECTION_POOL", 10)),
			MaxConnectionPool: uint64(utils.GetEnvInt("MONGO_MAX_CONNECTION_POOL", 50)),
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
			MaxBuffer:       utils.GetEnvInt("KAFKA_PRODUCER_MAX_BUFFER", 1000),
		},
		KafkaConsumer:     consumer,
		KafkaTestTopic:    utils.GetEnvMust("KAFKA_TEST_TOPIC"),
		KafkaTestTopic2:   utils.GetEnvMust("KAFKA_TEST_TOPIC_2"),
		KafkaHTTPProxyURL: utils.GetEnvMust("KAFKA_HTTP_PROXY"),
		Graylog: logwriter.GraylogConfig{
			Address: utils.GetEnv("GRAYLOG_URL", ""),
			Port:    utils.GetEnvInt("GRAYLOG_PORT", 12001),
		},
	}
}
