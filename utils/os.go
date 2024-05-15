package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sabariramc/goserverbase/v6/errors"
)

// GetEnv looks for the key in env, if not found returns defaultVal
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

// GetEnvInt looks for the key in env, calls [strconv.Atoi] with the value if no error returns the value else returns defaultVal
func GetEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if iVal, err := strconv.Atoi(value); err == nil {
			return iVal
		}
	}
	return defaultVal
}

// GetEnvBool looks for the key in env, if the value is "1" or "true" returns true else false if the value is not found returns defaultVal
func GetEnvBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if value == "1" || strings.ToLower(value) == "true" {
			return true
		}
		return false
	}
	return defaultVal
}

// GetEnvAsSlice looks for the key in env, if found splits the value using the sep
func GetEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := GetEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}

// GetEnvMust looks for the key in env, if not found raises panic
func GetEnvMust(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(errors.NewCustomError("MANDATORY_KEY_MISSING", fmt.Sprintf("mandatory environment variable is not set %v", key), nil, nil, true, nil))
	}
	return value
}
