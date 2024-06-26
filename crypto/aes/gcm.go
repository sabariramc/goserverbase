// Package aes provides encryption and decryption functionalities using the AES algorithm.
package aes

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"

	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/randomstring"
)

// GCM represents the Galois/Counter Mode (GCM) of AES encryption.
type GCM struct {
	key []byte
	log log.Log
}

// NewGCM creates a new GCM instance with the given key.
func NewGCM(ctx context.Context, log log.Log, key string) (*GCM, error) {
	keyByte, err := getKeyBytes(key)
	if err != nil {
		log.Error(ctx, "error in key", err)
		return nil, fmt.Errorf("NewGCM: %w", err)
	}
	return &GCM{key: keyByte, log: log.NewResourceLogger("GCM")}, nil
}

// Encrypt encrypts the given data using GCM mode and returns the encrypted data.
func (a *GCM) Encrypt(ctx context.Context, plainBlob []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		a.log.Error(ctx, "error creating cipher block", err)
		return nil, fmt.Errorf("GCM.Encrypt: error creating cipher block: %w", err)
	}
	cipher, err := cipher.NewGCM(block)
	if err != nil {
		a.log.Error(ctx, "error creating gcm cipher", err)
		return nil, fmt.Errorf("GCM.Encrypt: error creating gcm cipher: %w", err)
	}
	nonce := []byte(randomstring.Generate(cipher.NonceSize()))
	cipherBlob := cipher.Seal(nil, nonce, plainBlob, nil)
	return append(nonce, cipherBlob...), nil
}

// EncryptString encrypts the given plaintext string using GCM mode and returns the encrypted string.
func (a *GCM) EncryptString(ctx context.Context, plainText string) (string, error) {
	blobRes, err := a.Encrypt(ctx, []byte(plainText))
	if err != nil {
		a.log.Error(ctx, "error encrypting data", err)
		return "", fmt.Errorf("GCM.EncryptString: error encrypting content: %w", err)
	}
	res := base64.StdEncoding.EncodeToString(blobRes)
	return res, nil
}

// Decrypt decrypts the given data using GCM mode and returns the decrypted data.
func (a *GCM) Decrypt(ctx context.Context, encryptedData []byte) (plainData []byte, err error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		a.log.Error(ctx, "error creating cipher block", err)
		return nil, fmt.Errorf("GCM.Decrypt: error creating cipher block: %w", err)
	}
	cipher, err := cipher.NewGCM(block)
	if err != nil {
		a.log.Error(ctx, "error creating gcm cipher", err)
		return nil, fmt.Errorf("GCM.Decrypt: error creating gcm cipher: %w", err)
	}
	nonce := encryptedData[:cipher.NonceSize()]
	encryptedData = encryptedData[cipher.NonceSize():]
	plainData, err = cipher.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		a.log.Error(ctx, "error during decrypting data", err)
		return nil, fmt.Errorf("GCM.Decrypt: error decrypting content: %w", err)
	}
	return plainData, nil
}

// DecryptString decrypts the given base64-encoded encrypted text using GCM mode and returns the decrypted string.
func (a *GCM) DecryptString(ctx context.Context, encryptedText string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		a.log.Error(ctx, "error decoding encryptedData", err)
		return "", fmt.Errorf("GCM.DecryptString: error decoding content: %w", err)
	}
	blobRes, err := a.Decrypt(ctx, []byte(decoded))
	if err != nil {
		a.log.Error(ctx, "error decrypting", err)
		return "", fmt.Errorf("GCM.DecryptString: error decrypting content: %w", err)
	}
	res := string(blobRes)
	return res, nil
}
