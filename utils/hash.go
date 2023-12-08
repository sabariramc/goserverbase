package utils

import (
	"crypto/sha256"
	b64 "encoding/base64"
)

func GetHash(val string) string {
	hashByte := sha256.Sum256([]byte(val))
	return b64.StdEncoding.EncodeToString(hashByte[:])
}
