package aes

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/utils"
)

type GCM struct {
	key []byte
	log *log.Logger
}

func NewGCM(ctx context.Context, log *log.Logger, key string) (*GCM, error) {
	keyByte, err := getKeyBytes(key)
	if err != nil {
		log.Error(ctx, "error in key", err)
		return nil, fmt.Errorf("NewGCM: error in key: %w", err)
	}
	return &GCM{key: keyByte, log: log.NewResourceLogger("GCM")}, nil
}

func (a *GCM) Encrypt(ctx context.Context, plainBlob []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		a.log.Error(ctx, "error in creating cipher block", err)
		return nil, fmt.Errorf("GCM.Encrypt: error in creating cipher block: %w", err)
	}
	cipher, err := cipher.NewGCM(block)
	if err != nil {
		a.log.Error(ctx, "error in creating gcm cipher", err)
		return nil, fmt.Errorf("GCM.Encrypt: error in creating gcm cipher: %w", err)
	}
	nonce := []byte(utils.GenerateRandomString(cipher.NonceSize()))
	cipherBlob := cipher.Seal(nil, nonce, plainBlob, nil)
	return append(nonce, cipherBlob...), nil
}

func (a *GCM) EncryptString(ctx context.Context, plainText string) (string, error) {
	blobRes, err := a.Encrypt(ctx, []byte(plainText))
	if err != nil {
		a.log.Error(ctx, "error in encrypting data", err)
		return "", fmt.Errorf("GCM.EncryptString: error in encrypting data: %w", err)
	}
	res := base64.StdEncoding.EncodeToString(blobRes)
	return res, nil
}

func (a *GCM) Decrypt(ctx context.Context, encryptedData []byte) (plainData []byte, err error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		a.log.Error(ctx, "error in creating cipher block", err)
		return nil, fmt.Errorf("GCM.Decrypt: error in creating cipher block: %w", err)
	}
	cipher, err := cipher.NewGCM(block)
	if err != nil {
		a.log.Error(ctx, "error in creating gcm cipher", err)
		return nil, fmt.Errorf("GCM.Decrypt: error in creating gcm cipher: %w", err)
	}
	nonce := encryptedData[:cipher.NonceSize()]
	encryptedData = encryptedData[cipher.NonceSize():]
	plainData, err = cipher.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		a.log.Error(ctx, "error during decrypting data", err)
		return nil, fmt.Errorf("GCM.Decrypt: error during decrypting data: %w", err)
	}
	return plainData, nil
}

func (a *GCM) DecryptString(ctx context.Context, encryptedText string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		a.log.Error(ctx, "error in decoding encryptedData", err)
		return "", fmt.Errorf("GCM.DecryptString: error in decoding encryptedData: %w", err)
	}
	blobRes, err := a.Decrypt(ctx, []byte(decoded))
	if err != nil {
		a.log.Error(ctx, "error in decrypting", err)
		return "", fmt.Errorf("GCM.DecryptString: error in decrypting: %w", err)
	}
	res := string(blobRes)
	return res, nil
}
