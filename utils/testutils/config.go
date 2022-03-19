package testutils

import (
	"sabariram.com/goserverbase/config"
	"sabariram.com/goserverbase/utils"
)

type TestConfig struct {
	Mysql         *config.MySqlConnectionConfig
	Logger        *config.LoggerConfig
	App           *config.ServerConfig
	Mongo         *config.MongoConfig
	S3            *config.AWSS3Config
	KMS           *config.AWSConfig
	SecretManager *config.AWSConfig
	SNS           *config.AWSConfig
	SQS           *config.AWSSQSConfig
	FIFOSQS       *config.AWSSQSConfig
}

func (t *TestConfig) GetLoggerConfig() *config.LoggerConfig {
	return t.Logger
}
func (t *TestConfig) GetAppConfig() *config.ServerConfig {
	return t.App
}

func NewConfig() *TestConfig {
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
			Version:     utils.GetEnv("LOG_VERSION", "1.1"),
			Host:        utils.GetEnv("HOST", "localhost"),
			ServiceName: utils.GetEnv("SERVICE_NAME", "API"),
			LogLevel:    utils.GetEnvInt("LOG_LEVEL", 6),
			BufferSize:  utils.GetEnvInt("LOG_BUFFER_SIZE", 1),
			GrayLog: &config.GraylogConfig{
				URL:     utils.GetEnv("GRAYLOG_URL", "http://localhost:12201/gelf"),
				Address: utils.GetEnv("GRAYLOG_ADD", "localhost"),
				Port:    uint(utils.GetEnvInt("GRAYLOG_PORT", 12201)),
			},
			AuthHeaderKeyList: utils.GetEnvAsSlice("AUTH_HEADER_LIST", []string{}, ";"),
		},
		App: &config.ServerConfig{
			Host:        utils.GetEnv("HOST", "localhost"),
			Port:        utils.GetEnv("APP_PORT", "8080"),
			ServiceName: utils.GetEnv("APP_SERVICE_NAME", "API"),
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
	}
}
