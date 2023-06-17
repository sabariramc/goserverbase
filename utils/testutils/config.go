package testutils

import (
	"github.com/google/uuid"
	baseapp "github.com/sabariramc/goserverbase/v3/app"
	"github.com/sabariramc/goserverbase/v3/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v3/app/server/kafkaconsumer"
	"github.com/sabariramc/goserverbase/v3/db/mongo"
	"github.com/sabariramc/goserverbase/v3/kafka"
	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/sabariramc/goserverbase/v3/log/logwriter"
	"github.com/sabariramc/goserverbase/v3/utils"
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
	Logger            *log.Config
	App               *baseapp.ServerConfig
	Http              *httpserver.HttpServerConfig
	Kafka             *kafkaconsumer.KafkaConsumerServerConfig
	Mongo             *mongo.Config
	AWS               *AWSConfig
	KafkaConsumer     *kafka.KafkaConsumerConfig
	KafkaProducer     *kafka.KafkaProducerConfig
	KafkaTestTopic    string
	KafkaHTTPProxyURL string
	Graylog           logwriter.GraylogConfig
}

func (t *TestConfig) GetLoggerConfig() *log.Config {
	return t.Logger
}
func (t *TestConfig) GetAppConfig() *baseapp.ServerConfig {
	return t.App
}

func NewConfig() *TestConfig {
	serviceName := utils.GetEnv("SERVICE_NAME", "go-base")
	kafkaBaseConfig := kafka.KafkaCred{Brokers: utils.GetEnv("KAFKA_BROKER", ""),
		ClientID: serviceName + "-" + uuid.NewString(),
	}
	if utils.GetEnv("KAFKA_USERNAME", "") != "" {
		kafkaBaseConfig = kafka.KafkaCred{Brokers: utils.GetEnv("KAFKA_BROKER", ""),
			Username:      utils.GetEnv("KAFKA_USERNAME", ""),
			Password:      utils.GetEnv("KAFKA_PASSWORD", ""),
			SASLMechanism: "PLAIN",
			SASLProtocol:  "SASL_SSL",
			ClientID:      serviceName + "-" + uuid.NewString(),
		}
	}
	appConfig := &baseapp.ServerConfig{

		ServiceName: serviceName,
		Debug:       utils.GetEnvBool("DEBUG", false),
	}
	consumer := &kafka.KafkaConsumerConfig{
		KafkaCred:      &kafkaBaseConfig,
		GoEventChannel: false,
		GroupID:        utils.GetEnvMust("KAFKA_CONSUMER_ID"),
		OffsetReset:    "latest",
		MaxBuffer:      uint64(utils.GetEnvInt("KAFKA_CONSUMER_MAX_BUFFER", 1000)),
	}
	return &TestConfig{

		Logger: &log.Config{
			HostParams: log.HostParams{
				Version:     utils.GetEnv("LOG_VERSION", "1.1"),
				Host:        utils.GetEnv("HOST", utils.GetHostName()),
				ServiceName: serviceName,
			},
			LogLevel:   utils.GetEnvInt("LOG_LEVEL", 6),
			BufferSize: utils.GetEnvInt("LOG_BUFFER_SIZE", 1),
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
			KafkaCred:   &kafkaBaseConfig,
			Acknowledge: "all",
			MaxBuffer:   utils.GetEnvInt("KAFKA_PRODUCER_MAX_BUFFER", 1000),
		},
		KafkaConsumer:     consumer,
		KafkaTestTopic:    utils.GetEnvMust("KAFKA_TEST_TOPIC"),
		KafkaHTTPProxyURL: utils.GetEnvMust("KAFKA_HTTP_PROXY"),
		Graylog: logwriter.GraylogConfig{
			Address: utils.GetEnv("GRAYLOG_URL", ""),
			Port:    utils.GetEnvInt("GRAYLOG_PORT", 12001),
		},
	}
}
