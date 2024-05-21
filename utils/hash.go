package utils

import (
	"crypto/sha256"
	b64 "encoding/base64"
)

// GetHash computes the SHA-256 hash of the input value and returns it as a base64-encoded string.
func GetHash(val string) string {
	hashByte := sha256.Sum256([]byte(val))
	return b64.StdEncoding.EncodeToString(hashByte[:])
}
