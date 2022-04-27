package aes

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"

	"github.com/sabariramc/goserverbase/crypto"
	"github.com/sabariramc/goserverbase/crypto/padding"
	"github.com/sabariramc/goserverbase/log"
)

var ErrIVLengthMismatch = fmt.Errorf("IV lenght is not matching with block size")

type AESCBCV2 struct {
	padder crypto.Padder
	key    []byte
	log    *log.Logger
	iv     []byte
}

func NewAESCBCV2PKCS7(ctx context.Context, log *log.Logger, key string, iv []byte) (*AESCBCV2, error) {
	keyByte, err := getKey(key)
	if err != nil {
		log.Error(ctx, "Error creating AES CBC V2", err)
		return nil, fmt.Errorf("crypto.aes.NewAESCBCV2PKCS7ConstantIV: %w", err)
	}
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return nil, fmt.Errorf("crypto.aes.ChiperV2: %w", err)
	}
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("crypto.aes.ChiperV2: %w", ErrIVLengthMismatch)
	}
	return NewAESCBCV2(ctx, log, key, dup(iv), padding.NewPKCS7(block.BlockSize()))
}

func NewAESCBCV2(ctx context.Context, log *log.Logger, key string, iv []byte, padder crypto.Padder) (*AESCBCV2, error) {
	keyByte, err := getKey(key)
	if err != nil {
		log.Error(ctx, "Error creating AES CBC V2", err)
		return nil, fmt.Errorf("crypto.aes.NewAESCBCV2ConstantIV: %w", err)
	}
	return &AESCBCV2{key: keyByte, padder: padder, log: log, iv: iv}, nil
}

func (a *AESCBCV2) Encrypt(ctx context.Context, plainBlob []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, fmt.Errorf("AESCBCV2.Encrypt: %w", err)
	}
	paddedData := a.padder.Pad(plainBlob)
	blockModel := cipher.NewCBCEncrypter(block, a.iv)
	cipherBlob := make([]byte, len(paddedData))
	blockModel.CryptBlocks(cipherBlob, paddedData)
	return cipherBlob, nil
}

func (a *AESCBCV2) EncryptString(ctx context.Context, plainText string) (string, error) {
	a.log.Debug(ctx, "Plain Text", plainText)
	res, err := a.Encrypt(ctx, []byte(plainText))
	if err != nil {
		return "", fmt.Errorf("AESCBCV2.EncryptString: %w", err)
	}
	strres := base64.StdEncoding.EncodeToString(res)
	a.log.Debug(ctx, "EncryptedString", strres)
	return strres, nil
}

func (a *AESCBCV2) Decrypt(ctx context.Context, encryptedData []byte) (plainData []byte, err error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, fmt.Errorf("AESCBCV2.Decrypt: %w", err)
	}
	blockModel := cipher.NewCBCDecrypter(block, a.iv)
	plainData = make([]byte, len(encryptedData))
	blockModel.CryptBlocks(plainData, encryptedData)
	plainData = a.padder.UnPad(plainData)
	return plainData, nil
}

func (a *AESCBCV2) DecryptString(ctx context.Context, encryptedText string) (string, error) {
	a.log.Debug(ctx, "EncryptedString", encryptedText)
	decoded, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("AESCBCV2.DecryptString.B64Decode: %w", err)
	}
	res, err := a.Decrypt(ctx, []byte(decoded))
	if err != nil {
		return "", fmt.Errorf("AESCBCV2.DecryptString: %w", err)
	}
	strres := string(res)
	a.log.Debug(ctx, "DecryptedString", strres)
	return strres, nil
}

func dup(p []byte) []byte {
	q := make([]byte, len(p))
	copy(q, p)
	return q
}
