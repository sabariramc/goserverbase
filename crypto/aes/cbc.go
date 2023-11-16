package aes

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v4/crypto"
	"github.com/sabariramc/goserverbase/v4/crypto/padding"
	"github.com/sabariramc/goserverbase/v4/log"
)

var ErrBlockError = fmt.Errorf("cipher text is not a multiple of the block size")
var ErrInvalidKeyLength = fmt.Errorf("invalid key length")
var ErrIVLengthMismatch = fmt.Errorf("IV length is not matching with block size")

type CBC struct {
	padder crypto.Padder
	key    []byte
	log    *log.Logger
	iv     []byte
}

func NewCBCPKCS7(ctx context.Context, log *log.Logger, key string, iv []byte) (*CBC, error) {
	keyByte, err := getKeyBytes(key)
	if err != nil {
		return nil, fmt.Errorf("crypto.aes.NewAESCBCPKCS7: error creating key: %w", err)
	}
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return nil, fmt.Errorf("crypto.aes.Cipher: %w", err)
	}
	return NewCBC(ctx, log, key, padding.NewPKCS7(block.BlockSize()), iv)
}

func NewCBC(ctx context.Context, log *log.Logger, key string, padder crypto.Padder, iv []byte) (*CBC, error) {
	keyByte, err := getKeyBytes(key)
	if err != nil {
		return nil, fmt.Errorf("crypto.aes.NewCBC: error in key: %w", err)
	}
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return nil, fmt.Errorf("crypto.aes.NewCBC: error in creating a block: %w", err)
	}
	if iv != nil && len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("crypto.aes.NewCBC: %w", ErrIVLengthMismatch)
	}
	return &CBC{key: keyByte, padder: padder, log: log.NewResourceLogger("AESCBC"), iv: iv}, nil
}

func (a *CBC) Encrypt(ctx context.Context, plainBlob []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, fmt.Errorf("AESCBC.Encrypt: %w", err)
	}
	paddedData := a.padder.Pad(plainBlob)
	iv := a.getIv()
	blockModel := cipher.NewCBCEncrypter(block, iv[:block.BlockSize()])
	cipherBlob := make([]byte, len(paddedData))
	blockModel.CryptBlocks(cipherBlob, paddedData)
	return append(iv[:block.BlockSize()], cipherBlob...), nil
}

func (a *CBC) EncryptString(ctx context.Context, plainText string) (string, error) {
	a.log.Debug(ctx, "Plain Text", plainText)
	blobRes, err := a.Encrypt(ctx, []byte(plainText))
	if err != nil {
		return "", fmt.Errorf("AESCBC.EncryptString: %w", err)
	}
	res := base64.StdEncoding.EncodeToString(blobRes)
	a.log.Debug(ctx, "EncryptedString", res)
	return res, nil
}

func (a *CBC) Decrypt(ctx context.Context, encryptedData []byte) (plainData []byte, err error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, fmt.Errorf("AESCBC.Decrypt: %w", err)
	}
	blockSize := block.BlockSize()
	iv := encryptedData[:blockSize]
	encryptedData = encryptedData[blockSize:]
	if len(encryptedData)%blockSize != 0 {
		return nil, fmt.Errorf("AESCBC.Decrypt.Block: %w", ErrBlockError)
	}
	blockModel := cipher.NewCBCDecrypter(block, iv)
	plainData = make([]byte, len(encryptedData))
	blockModel.CryptBlocks(plainData, encryptedData)
	plainData = a.padder.UnPad(plainData)
	return plainData, nil
}

func (a *CBC) DecryptString(ctx context.Context, encryptedText string) (string, error) {
	a.log.Debug(ctx, "EncryptedString", encryptedText)
	decoded, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("AESCBC.DecryptString.B64Decode: %w", err)
	}
	blobRes, err := a.Decrypt(ctx, []byte(decoded))
	if err != nil {
		return "", fmt.Errorf("AESCBC.DecryptString: %w", err)
	}
	res := string(blobRes)
	a.log.Debug(ctx, "DecryptedString", res)
	return res, nil
}

func (a *CBC) getIv() []byte {
	iv := a.iv
	if a.iv == nil {
		iv = []byte(strings.Replace(uuid.New().String(), "-", "", -1))
	}
	return iv
}
