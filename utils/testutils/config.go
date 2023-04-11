package testutils

import (
	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/config"
	"github.com/sabariramc/goserverbase/db/mongo"
	"github.com/sabariramc/goserverbase/kafka"
	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/utils"
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
	App                 *config.ServerConfig
	Mongo               *mongo.Config
	AWS                 *AWSConfig
	KafkaConsumerConfig *kafka.KafkaConsumerConfig
	KafkaProducerConfig *kafka.KafkaProducerConfig
	KafkaTestTopic      string
	KafkaHTTPProxyURL   string
}

func (t *TestConfig) GetLoggerConfig() *log.Config {
	return t.Logger
}
func (t *TestConfig) GetAppConfig() *config.ServerConfig {
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
	return &TestConfig{

		Logger: &log.Config{
			Version:           utils.GetEnv("LOG_VERSION", "1.1"),
			Host:              utils.GetEnv("HOST", utils.GetHostName()),
			ServiceName:       utils.GetEnv("SERVICE_NAME", "API"),
			LogLevel:          utils.GetEnvInt("LOG_LEVEL", 6),
			BufferSize:        utils.GetEnvInt("LOG_BUFFER_SIZE", 1),
			AuthHeaderKeyList: utils.GetEnvAsSlice("AUTH_HEADER_LIST", []string{}, ";"),
		},
		App: &config.ServerConfig{
			Host:        "0.0.0.0",
			Port:        utils.GetEnv("APP_PORT", "8080"),
			ServiceName: utils.GetEnv("SERVICE_NAME", "API"),
			Debug:       utils.GetEnvBool("DEBUG", false),
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
		KafkaProducerConfig: &kafka.KafkaProducerConfig{
			KafkaCred:   &kafkaBaseConfig,
			Acknowledge: "all",
		},
		KafkaConsumerConfig: &kafka.KafkaConsumerConfig{
			KafkaCred:      &kafkaBaseConfig,
			GoEventChannel: false,
			GroupID:        utils.GetEnv("KAFKA_CONSUMER_ID", serviceName),
			OffsetReset:    "latest",
		},
		KafkaTestTopic:    utils.GetEnv("KAFKA_TEST_TOPIC", serviceName),
		KafkaHTTPProxyURL: utils.GetEnv("KAFKA_HTTP_PROXY", serviceName),
	}
}
