package utils

import (
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/json"
)

func GetString(val interface{}) (*string, error) {
	blob, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	str := string(blob)
	return &str, nil
}

func GetHash(val string) string {
	hashByte := sha256.Sum256([]byte(val))
	return b64.StdEncoding.EncodeToString(hashByte[:])
}
