// Package aes wraps crypto/aes package cipher to be compatible with Cipher interface defined under Crypto package
package aes

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v5/crypto"
	"github.com/sabariramc/goserverbase/v5/crypto/padding"
	"github.com/sabariramc/goserverbase/v5/log"
)

var ErrBlockError = fmt.Errorf("cipher text is not a multiple of the block size")
var ErrInvalidKeyLength = fmt.Errorf("invalid key length")
var ErrIVLengthMismatch = fmt.Errorf("IV length is not matching with block size")

type CBC struct {
	padder crypto.Padder
	key    []byte
	log    log.Log
	iv     []byte
}

func NewCBCPKCS7(ctx context.Context, log log.Log, key string, iv []byte) (*CBC, error) {
	keyByte, err := getKeyBytes(key)
	if err != nil {
		log.Error(ctx, "erroneous in key", err)
		return nil, fmt.Errorf("NewAESCBCPKCS7: %w", err)
	}
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return nil, fmt.Errorf("NewAESCBCPKCS7: error creating CBCPKCS7: %w", err)
	}
	return NewCBC(ctx, log, key, padding.NewPKCS7(block.BlockSize()), iv)
}

func NewCBC(ctx context.Context, log log.Log, key string, padder crypto.Padder, iv []byte) (*CBC, error) {
	keyByte, err := getKeyBytes(key)
	if err != nil {
		log.Error(ctx, "error in key", err)
		return nil, fmt.Errorf("crypto.aes.NewCBC: %w", err)
	}
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		log.Error(ctx, "error creating cipher", err)
		return nil, fmt.Errorf("crypto.aes.NewCBC: error creating cipher: %w", err)
	}
	if iv != nil && len(iv) != block.BlockSize() {
		log.Error(ctx, "invalid iv length", ErrIVLengthMismatch)
		return nil, fmt.Errorf("crypto.aes.NewCBC: invalid iv length: %w", ErrIVLengthMismatch)
	}
	return &CBC{key: keyByte, padder: padder, log: log.NewResourceLogger("CBC"), iv: iv}, nil
}

func (a *CBC) Encrypt(ctx context.Context, plainBlob []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		a.log.Error(ctx, "error creating cipher", err)
		return nil, fmt.Errorf("CBC.Encrypt: error creating cipher: %w", err)
	}
	paddedData := a.padder.Pad(plainBlob)
	iv := a.getIv()
	blockModel := cipher.NewCBCEncrypter(block, iv[:block.BlockSize()])
	cipherBlob := make([]byte, len(paddedData))
	blockModel.CryptBlocks(cipherBlob, paddedData)
	return append(iv[:block.BlockSize()], cipherBlob...), nil
}

func (a *CBC) EncryptString(ctx context.Context, data string) (string, error) {
	blobRes, err := a.Encrypt(ctx, []byte(data))
	if err != nil {
		a.log.Error(ctx, "error encrypting data", err)
		return "", fmt.Errorf("CBC.EncryptString: error encrypting data: %w", err)
	}
	res := base64.StdEncoding.EncodeToString(blobRes)
	return res, nil
}

func (a *CBC) Decrypt(ctx context.Context, encryptedData []byte) (plainData []byte, err error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		a.log.Error(ctx, "error creating cipher", err)
		return nil, fmt.Errorf("CBC.Decrypt: error creating cipher: %w", err)
	}
	blockSize := block.BlockSize()
	iv := encryptedData[:blockSize]
	encryptedData = encryptedData[blockSize:]
	if len(encryptedData)%blockSize != 0 {
		a.log.Error(ctx, "invalid block size", ErrBlockError)
		return nil, fmt.Errorf("CBC.Decrypt: invalid block size: %w", ErrBlockError)
	}
	blockModel := cipher.NewCBCDecrypter(block, iv)
	plainData = make([]byte, len(encryptedData))
	blockModel.CryptBlocks(plainData, encryptedData)
	plainData = a.padder.UnPad(plainData)
	return plainData, nil
}

func (a *CBC) DecryptString(ctx context.Context, encryptedData string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		a.log.Error(ctx, "error decoding encryptedData", err)
		return "", fmt.Errorf("CBC.DecryptString: error decoding content: %w", err)
	}
	blobRes, err := a.Decrypt(ctx, []byte(decoded))
	if err != nil {
		a.log.Error(ctx, "error decrypting content", err)
		return "", fmt.Errorf("CBC.DecryptString: error decrypting content: %w", err)
	}
	res := string(blobRes)
	return res, nil
}

func (a *CBC) getIv() []byte {
	iv := a.iv
	if a.iv == nil {
		iv = []byte(strings.Replace(uuid.New().String(), "-", "", -1))
	}
	return iv
}
