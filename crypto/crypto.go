package crypto

import "context"

type Cipher interface {
	Encrypt(ctx context.Context, plainBlob []byte) ([]byte, error)
	Decrypt(ctx context.Context, encryptedBlob []byte) ([]byte, error)
}

type CipherText interface {
	Cipher
	EncryptString(ctx context.Context, plaintext string) (string, error)
	DecryptString(ctx context.Context, encryptedSting string) (string, error)
}
