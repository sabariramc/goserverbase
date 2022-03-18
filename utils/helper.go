package utils

import (
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"sabariram.com/goserverbase/log"
)

var IST, _ = time.LoadLocation("Asia/Kolkata")

func GetStringValue(data map[string]interface{}, key string) (string, bool) {
	logger := log.GetDefaultLogger()
	valIn, ok := data[key]
	if ok {
		logger.Info(fmt.Sprintf("key found - %v, value - %v", key, valIn), nil)
		return valIn.(string), ok
	}
	logger.Info(fmt.Sprintf("key not found - %v", key), nil)
	return "", ok
}

func GetMapValue(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	valIn, ok := data[key]
	logger := log.GetDefaultLogger()
	if ok {
		logger.Info(fmt.Sprintf("key found - %v", key), valIn)
		return valIn.(map[string]interface{}), ok
	}
	logger.Info(fmt.Sprintf("key not found - %v", key), nil)
	return nil, ok
}

func GetString(obj interface{}) (*string, error) {
	blob, err := json.Marshal(obj)
	logger := log.GetDefaultLogger()
	if err != nil {
		logger.Error("Error marshalling object", err)
		logger.Error("Marshall input object", obj)
		return nil, err
	}
	body := string(blob)
	return &body, nil
}

func GetHash(val string) string {
	hashByte := sha256.Sum256([]byte(val))
	return b64.StdEncoding.EncodeToString(hashByte[:])
}
