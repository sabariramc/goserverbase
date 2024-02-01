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

	"github.com/sabariramc/goserverbase/v5/crypto"
	"github.com/sabariramc/goserverbase/v5/crypto/padding"
	"github.com/sabariramc/goserverbase/v5/log"
)

type ECB struct {
	b         cipher.Block
	key       []byte
	blockSize int
	padder    crypto.Padder
	log       log.Log
}

func NewECBPKC5(ctx context.Context, key []byte, log log.Log) (*ECB, error) {
	cipher, err := NewECB(ctx, key, log, padding.NewPKCS7(len(key)))
	if err != nil {
		log.Error(ctx, "error creating ECB", err)
		return nil, fmt.Errorf("NewECBPKC5: error creating ECB: %w", err)
	}
	return cipher, nil
}

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
	blobRes, err := a.Encrypt(ctx, []byte(plainText))
	if err != nil {
		a.log.Error(ctx, "error encryption content", err)
		return "", fmt.Errorf("ECB.EncryptString: error encryption content: %w", err)
	}
	res := base64.StdEncoding.EncodeToString(blobRes)
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
