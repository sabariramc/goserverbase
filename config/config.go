package config

type MySqlConnectionConfig struct {
	Host         string
	Port         string
	DatabaseName string
	Username     string
	Password     string
	Timezone     string
	Charset      string
}

type AWSConfig struct {
	Arn string
}
type AWSSQSConfig struct {
	QueueURL string
}
type AWSS3Config struct {
	BucketName string
}

type MongoConfig struct {
	ConnectionString string
}

type ServerConfig struct {
	Host        string
	Port        string
	ServiceName string
}

type GraylogConfig struct {
	URL     string
	Address string
	Port    uint
}

type LoggerConfig struct {
	Version           string
	Host              string
	ServiceName       string
	LogLevel          int
	BufferSize        int
	GrayLog           *GraylogConfig
	AuthHeaderKeyList []string
}

type RuntimeConfig struct {
	GoMaxProcs int
}
