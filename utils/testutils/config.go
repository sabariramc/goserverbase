package testutils

import (
	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/config"
	"github.com/sabariramc/goserverbase/kafka"
	"github.com/sabariramc/goserverbase/utils"
)

type TestConfig struct {
	Mysql               *config.MySqlConnectionConfig
	Logger              *config.LoggerConfig
	App                 *config.ServerConfig
	Mongo               *config.MongoConfig
	MongoCSFLE          *config.MongoCFLEConfig
	S3                  *config.AWSS3Config
	KMS                 *config.AWSConfig
	SecretManager       *config.AWSConfig
	SNS                 *config.AWSConfig
	SQS                 *config.AWSSQSConfig
	FIFOSQS             *config.AWSSQSConfig
	KafkaConsumerConfig *kafka.KafkaConsumerConfig
	KafkaProducerConfig *kafka.KafkaProducerConfig
	KafkaTestTopic      string
}

func (t *TestConfig) GetLoggerConfig() *config.LoggerConfig {
	return t.Logger
}
func (t *TestConfig) GetAppConfig() *config.ServerConfig {
	return t.App
}

func NewConfig() *TestConfig {
	serviceName := utils.GetEnv("SERVICE_NAME", "lending-errornotifier")
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
		Mysql: &config.MySqlConnectionConfig{
			Host:         utils.GetEnv("MYSQL_HOST", "localhost"),
			Port:         utils.GetEnv("MYSQL_PORT", "3306"),
			DatabaseName: utils.GetEnv("MYSQL_DATABASE", ""),
			Username:     utils.GetEnv("MYSQL_USERNAME", "root"),
			Password:     utils.GetEnv("MYSQL_PASSWORD", ""),
			Timezone:     utils.GetEnv("MYSQL_TIMEZONE", "Local"),
			Charset:      utils.GetEnv("MYSQL_CHARSET", "utf8"),
		},
		Logger: &config.LoggerConfig{
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
		Mongo: &config.MongoConfig{
			ConnectionString:  utils.GetEnv("MONGO_URL", "mongodb://localhost:60001"),
			DatabaseName:      utils.GetEnv("MONGO_DATABASE", "GOLANGTEST"),
			MinConnectionPool: uint64(utils.GetEnvInt("MONGO_MIN_CONNECTION_POOL", 10)),
			MaxConnectionPool: uint64(utils.GetEnvInt("MONGO_MAX_CONNECTION_POOL", 50)),
		},
		MongoCSFLE: &config.MongoCFLEConfig{
			KeyVaultNamespace: utils.GetEnv("MONGO_KEY_VAULT", "encryption.__keyVault"),
			MasterKeyARN: &config.AWSConfig{
				Arn: utils.GetEnv("KMS_ARN", ""),
			},
		},
		S3: &config.AWSS3Config{
			BucketName: utils.GetEnv("S3_BUCKET", ""),
		},
		KMS: &config.AWSConfig{
			Arn: utils.GetEnv("KMS_ARN", ""),
		},
		SecretManager: &config.AWSConfig{
			Arn: utils.GetEnv("SECRET_ARN", ""),
		},
		SQS: &config.AWSSQSConfig{
			QueueURL: utils.GetEnv("SQS_URL", ""),
		},
		FIFOSQS: &config.AWSSQSConfig{
			QueueURL: utils.GetEnv("FIFO_SQS_URL", ""),
		},
		SNS: &config.AWSConfig{
			Arn: utils.GetEnv("SNS_ARN", ""),
		},
		KafkaProducerConfig: &kafka.KafkaProducerConfig{
			KafkaCred:   &kafkaBaseConfig,
			Acknowledge: "all",
		},
		KafkaConsumerConfig: &kafka.KafkaConsumerConfig{
			KafkaCred:      &kafkaBaseConfig,
			GoEventChannel: false,
			GroupID:        serviceName,
			OffsetReset:    "latest",
		},
		KafkaTestTopic: "com.sabariram.test",
	}
}
