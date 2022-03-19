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

type Config interface {
	GetLoggerConfig() *LoggerConfig
	GetAppConfig() *ServerConfig
}

func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func GetEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if iVal, err := strconv.Atoi(value); err == nil {
			return iVal
		}
	}
	return defaultVal
}

func GetEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := GetEnv(name, "")

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
