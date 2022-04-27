package aes

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/crypto"
	"github.com/sabariramc/goserverbase/crypto/padding"
	"github.com/sabariramc/goserverbase/log"
)

var ErrBlockError = fmt.Errorf("ciphertext is not a multiple of the block size")
var ErrInvalidKeyLength = fmt.Errorf("Invalid key length")

type AESCBC struct {
	padder crypto.Padder
	key    []byte
	log    *log.Logger
}

func NewAESCBCPKCS7(ctx context.Context, log *log.Logger, key string) (*AESCBC, error) {
	keyByte, err := getKey(key)
	if err != nil {
		log.Error(ctx, "Error creating AES CBC", err)
		return nil, fmt.Errorf("crypto.aes.NewAESCBCPKCS7: %w", err)
	}
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return nil, fmt.Errorf("crypto.aes.Chiper: %w", err)
	}
	return NewAESCBC(ctx, log, key, padding.NewPKCS7(block.BlockSize()))
}

func NewAESCBC(ctx context.Context, log *log.Logger, key string, padder crypto.Padder) (*AESCBC, error) {
	keyByte, err := getKey(key)
	if err != nil {
		log.Error(ctx, "Error creating AES CBC", err)
		return nil, fmt.Errorf("crypto.aes.NewAESCBC: %w", err)
	}
	return &AESCBC{key: keyByte, padder: padder, log: log}, nil
}

func getKey(key string) ([]byte, error) {
	keyLen := len(key)
	if keyLen != 16 || keyLen != 24 || keyLen != 32 {
		return nil, ErrInvalidKeyLength
	}
	return []byte(key), nil
}

func (a *AESCBC) Encrypt(plainBlob []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, fmt.Errorf("AESCBC.Encrypt: %w", err)
	}
	paddedData := a.padder.Pad(plainBlob)
	iv := []byte(strings.Replace(uuid.New().String(), "-", "", -1))
	blockModel := cipher.NewCBCEncrypter(block, iv[:block.BlockSize()])
	cipherBlob := make([]byte, len(paddedData))
	blockModel.CryptBlocks(cipherBlob, paddedData)
	return append(iv[:block.BlockSize()], cipherBlob...), nil
}

func (a *AESCBC) EncryptString(plainText string) (string, error) {
	res, err := a.Encrypt([]byte(plainText))
	if err != nil {
		return "", fmt.Errorf("AESCBC.EncryptString: %w", err)
	}
	return base64.StdEncoding.EncodeToString(res), nil
}

func (a *AESCBC) Decrypt(encryptedData []byte) (plainData []byte, err error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, fmt.Errorf("AESCBC.Decrypt: %w", err)
	}
	blockSize := block.BlockSize()
	iv := encryptedData[:blockSize]
	encryptedData = encryptedData[blockSize:]
	if len(encryptedData)%blockSize != 0 {
		return nil, fmt.Errorf("AESCBC.EncryptString: %w", ErrBlockError)
	}
	blockModel := cipher.NewCBCDecrypter(block, iv)
	plainData = make([]byte, len(encryptedData))
	blockModel.CryptBlocks(plainData, encryptedData)
	plainData = a.padder.UnPad(plainData)
	return plainData, nil
}

func (a *AESCBC) DecryptString(plainText string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(plainText)
	if err != nil {
		return "", fmt.Errorf("AESCBC.DecryptString.B64Decode: %w", err)
	}
	res, err := a.Decrypt([]byte(decoded))
	if err != nil {
		return "", fmt.Errorf("AESCBC.EncryptString: %w", err)
	}
	return string(res), nil
}
