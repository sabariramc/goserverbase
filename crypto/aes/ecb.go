/*
Package aes is a wrapped for AES/ECB/PKCS5PADDING
*/
package aes

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"

	"github.com/sabariramc/goserverbase/v4/crypto"
	"github.com/sabariramc/goserverbase/v4/crypto/padding"
	"github.com/sabariramc/goserverbase/v4/log"
)

type ECB struct {
	b         cipher.Block
	key       []byte
	blockSize int
	padder    crypto.Padder
	log       *log.Logger
}

func NewECBPKC5(ctx context.Context, key []byte, log *log.Logger) (*ECB, error) {
	cipher, err := NewECB(key, log, padding.NewPKCS7(len(key)))
	if err != nil {
		return nil, fmt.Errorf("crypto.aes.NewAESGCM: error creating aes ecb: %w", err)
	}
	return cipher, nil
}

func NewECB(key []byte, log *log.Logger, padder crypto.Padder) (*ECB, error) {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("AESECBCipher: key error: key - %v, error - %w", key, err)
	}
	return &ECB{
		b:         cipher,
		key:       key,
		blockSize: cipher.BlockSize(),
		padder:    padder,
		log:       log.NewResourceLogger("AESECB"),
	}, nil
}

func (a *ECB) Encrypt(ctx context.Context, data []byte) ([]byte, error) {
	data = a.padder.Pad(data)
	encrypted := make([]byte, len(data))
	size := a.blockSize
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		a.b.Encrypt(encrypted[bs:be], data[bs:be])
	}
	return encrypted, nil
}

func (a *ECB) EncryptString(ctx context.Context, plainText string) (string, error) {
	a.log.Debug(ctx, "Plain Text", plainText)
	blobRes, err := a.Encrypt(ctx, []byte(plainText))
	if err != nil {
		return "", fmt.Errorf("AESGCM.EncryptString: %w", err)
	}
	res := base64.StdEncoding.EncodeToString(blobRes)
	a.log.Debug(ctx, "EncryptedString", res)
	return res, nil
}

func (a *ECB) Decrypt(ctx context.Context, data []byte) ([]byte, error) {
	decrypted := make([]byte, len(data))
	size := a.blockSize
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		a.b.Decrypt(decrypted[bs:be], data[bs:be])
	}
	decrypted = a.padder.UnPad(decrypted)
	return decrypted, nil
}
