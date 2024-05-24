// Package aes provides encryption and decryption functionalities using the AES algorithm.
package aes

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"

	"github.com/sabariramc/goserverbase/v6/crypto"
	"github.com/sabariramc/goserverbase/v6/crypto/padding"
	"github.com/sabariramc/goserverbase/v6/log"
)

// ECB represents the Electronic Codebook (ECB) mode of AES encryption.
type ECB struct {
	b         cipher.Block
	key       []byte
	blockSize int
	padder    crypto.Padder
	log       log.Log
}

// NewECBPKC5 creates a new ECB instance with PKCS7 padding.
func NewECBPKC5(ctx context.Context, key []byte, log log.Log) (*ECB, error) {
	cipher, err := NewECB(ctx, key, log, padding.NewPKCS7(len(key)))
	if err != nil {
		log.Error(ctx, "error creating ECB", err)
		return nil, fmt.Errorf("NewECBPKC5: error creating ECB: %w", err)
	}
	return cipher, nil
}

// NewECB creates a new ECB instance with custom padding.
func NewECB(ctx context.Context, key []byte, log log.Log, padder crypto.Padder) (*ECB, error) {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		log.Error(ctx, "error creating cipher", err)
		return nil, fmt.Errorf("NewECB: error creating cipher: %w", err)
	}
	return &ECB{
		b:         cipher,
		key:       key,
		blockSize: cipher.BlockSize(),
		padder:    padder,
		log:       log.NewResourceLogger("ECB"),
	}, nil
}

// Encrypt encrypts the given data using ECB mode and returns the encrypted data.
func (a *ECB) Encrypt(ctx context.Context, data []byte) ([]byte, error) {
	data = a.padder.Pad(data)
	encrypted := make([]byte, len(data))
	size := a.blockSize
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		a.b.Encrypt(encrypted[bs:be], data[bs:be])
	}
	return encrypted, nil
}

// EncryptString encrypts the given plaintext string using ECB mode and returns the encrypted string.
func (a *ECB) EncryptString(ctx context.Context, plainText string) (string, error) {
	blobRes, err := a.Encrypt(ctx, []byte(plainText))
	if err != nil {
		a.log.Error(ctx, "error encryption content", err)
		return "", fmt.Errorf("EncryptString: error encryption content: %w", err)
	}
	res := base64.StdEncoding.EncodeToString(blobRes)
	return res, nil
}

// Decrypt decrypts the given data using ECB mode and returns the decrypted data.
func (a *ECB) Decrypt(ctx context.Context, data []byte) ([]byte, error) {
	decrypted := make([]byte, len(data))
	size := a.blockSize
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		a.b.Decrypt(decrypted[bs:be], data[bs:be])
	}
	decrypted = a.padder.UnPad(decrypted)
	return decrypted, nil
}
