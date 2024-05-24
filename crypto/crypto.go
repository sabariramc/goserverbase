// Package crypto defines common interfaces for encryption and decryption ciphers, including string-based methods.
package crypto

import "context"

// Cipher provides methods for encrypting and decrypting byte slices.
type Cipher interface {
	// Encrypt encrypts the given plainBlob and returns the encrypted data.
	Encrypt(ctx context.Context, plainBlob []byte) ([]byte, error)
	// Decrypt decrypts the given encryptedBlob and returns the original data.
	Decrypt(ctx context.Context, encryptedBlob []byte) ([]byte, error)
}

// CipherText extends the Cipher interface with methods for encrypting and decrypting strings.
type CipherText interface {
	Cipher
	// EncryptString encrypts the given plaintext string and returns the encrypted string.
	EncryptString(ctx context.Context, plaintext string) (string, error)
	// DecryptString decrypts the given encryptedString and returns the original string.
	DecryptString(ctx context.Context, encryptedString string) (string, error)
}
