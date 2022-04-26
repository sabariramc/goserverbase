package crypto

type Chiper interface {
	Encrypt(plainBlob []byte) ([]byte, error)
	Decrypt(encryptedBlob []byte) ([]byte, error)
	EncryptString(plaintext string) (string, error)
	DecryptString(encryptedSting string) (string, error)
}
