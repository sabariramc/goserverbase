package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type MySqlConnectionConfig struct {
	Host         string
	Port         string
	DatabaseName string
	Username     string
	Password     string
	Timezone     string
	Charset      string
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

type Config struct {
	Mysql  *MySqlConnectionConfig
	Logger *LoggerConfig
	App    *ServerConfig
}

func NewConfig() *Config {
	return &Config{
		Mysql: &MySqlConnectionConfig{
			Host:         getEnv("MYSQL_HOST", "localhost"),
			Port:         getEnv("MYSQL_PORT", "3306"),
			DatabaseName: getEnv("MYSQL_DATABASE", ""),
			Username:     getEnv("MYSQL_USERNAME", "root"),
			Password:     getEnv("MYSQL_PASSWORD", ""),
			Timezone:     getEnv("MYSQL_TIMEZONE", "Local"),
			Charset:      getEnv("MYSQL_CHARSET", "utf8"),
		},
		Logger: &LoggerConfig{
			Version:     getEnv("LOG_VERSION", "1.1"),
			Host:        getEnv("HOST", "localhost"),
			ServiceName: getEnv("SERVICE_NAME", "API"),
			LogLevel:    getEnvInt("LOG_LEVEL", 6),
			BufferSize:  getEnvInt("LOG_BUFFER_SIZE", 1),
			GrayLog: &GraylogConfig{
				URL:     getEnv("GRAYLOG_URL", "http://localhost:12201/gelf"),
				Address: getEnv("GRAYLOG_ADD", "localhost"),
				Port:    uint(getEnvInt("GRAYLOG_PORT", 12201)),
			},
			AuthHeaderKeyList: getEnvAsSlice("AUTH_HEADER_LIST", []string{}, ";"),
		},
		App: &ServerConfig{
			Host:        getEnv("HOST", "localhost"),
			Port:        getEnv("APP_PORT", "8080"),
			ServiceName: getEnv("APP_SERVICE_NAME", "API"),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if iVal, err := strconv.Atoi(value); err == nil {
			return iVal
		}
	}
	return defaultVal
}

func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}
