package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sabariramc/goserverbase/v3/errors"
)

func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func GetHostName() string {
	nodeName, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return nodeName
}

func GetEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if iVal, err := strconv.Atoi(value); err == nil {
			return iVal
		}
	}
	return defaultVal
}

func GetEnvBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if value == "1" || strings.ToLower(value) == "true" {
			return true
		}
		return false
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

func GetEnvMust(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(errors.NewCustomError("MANDATORY_KEY_MISSING", fmt.Sprintf("mandatory environment variable is not set %v", key), nil, nil, true))
	}
	return value
}
