package utils

import (
	"bytes"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
)

func GetString(val interface{}) (*string, error) {
	blob, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("utils.GetString : %w", err)
	}
	str := string(blob)
	return &str, nil
}

func GetHash(val string) string {
	hashByte := sha256.Sum256([]byte(val))
	return b64.StdEncoding.EncodeToString(hashByte[:])
}

func JsonTransformer(src interface{}, dest interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(src)
	if err != nil {
		return fmt.Errorf("JsonTransformer encoding: %w", err)
	}
	decoder := json.NewDecoder(&buf)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(dest)
	if err != nil {
		return fmt.Errorf("JsonTransformer decoding: %w", err)
	}
	return nil
}

func JsonTransformerLoosy(src interface{}, dest interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(src)
	if err != nil {
		return fmt.Errorf("JsonTransformer encoding: %w", err)
	}
	decoder := json.NewDecoder(&buf)
	err = decoder.Decode(dest)
	if err != nil {
		return fmt.Errorf("JsonTransformer decoding: %w", err)
	}
	return nil
}
