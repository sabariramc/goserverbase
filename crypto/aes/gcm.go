package aes

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"

	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/sabariramc/goserverbase/v3/utils"
)

type AESGCM struct {
	key []byte
	log *log.Logger
}

func NewAESGCM(ctx context.Context, log *log.Logger, key string) (*AESGCM, error) {
	keyByte, err := getKey(key)
	if err != nil {
		log.Error(ctx, "Error creating AES CBC", err)
		return nil, fmt.Errorf("crypto.aes.NewAESGCM: %w", err)
	}
	return &AESGCM{key: keyByte, log: log.NewResourceLogger("AESGCM")}, nil
}

func (a *AESGCM) Encrypt(ctx context.Context, plainBlob []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, fmt.Errorf("AESGCM.Encrypt: %w", err)
	}
	cipher, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("AESGCM.Encrypt: %w", err)
	}
	nonce := []byte(utils.GenerateRandomString(cipher.NonceSize()))
	cipherBlob := cipher.Seal(nil, nonce, plainBlob, nil)
	return append(nonce, cipherBlob...), nil
}

func (a *AESGCM) EncryptString(ctx context.Context, plainText string) (string, error) {
	a.log.Debug(ctx, "Plain Text", plainText)
	blobRes, err := a.Encrypt(ctx, []byte(plainText))
	if err != nil {
		return "", fmt.Errorf("AESGCM.EncryptString: %w", err)
	}
	res := base64.StdEncoding.EncodeToString(blobRes)
	a.log.Debug(ctx, "EncryptedString", res)
	return res, nil
}

func (a *AESGCM) Decrypt(ctx context.Context, encryptedData []byte) (plainData []byte, err error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, fmt.Errorf("AESGCM.Decrypt: %w", err)
	}
	cipher, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("AESGCM.Decrypt: %w", err)
	}
	nonce := encryptedData[:cipher.NonceSize()]
	encryptedData = encryptedData[cipher.NonceSize():]
	plainData, err = cipher.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("AESGCM.Decrypt: %w", err)
	}
	return plainData, nil
}

func (a *AESGCM) DecryptString(ctx context.Context, encryptedText string) (string, error) {
	a.log.Debug(ctx, "EncryptedString", encryptedText)
	decoded, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("AESGCM.DecryptString.B64Decode: %w", err)
	}
	blobRes, err := a.Decrypt(ctx, []byte(decoded))
	if err != nil {
		return "", fmt.Errorf("AESGCM.DecryptString: %w", err)
	}
	res := string(blobRes)
	a.log.Debug(ctx, "DecryptedString", res)
	return res, nil
}
