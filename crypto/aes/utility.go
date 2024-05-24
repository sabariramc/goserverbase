package aes

// getKeyBytes converts the given key string into a byte slice and performs validation on the key length.
// It returns an error if the key length is not 16, 24, or 32 bytes.
func getKeyBytes(key string) ([]byte, error) {
	keyByte := []byte(key)
	keyLen := len(keyByte)
	switch keyLen {
	default:
		return nil, ErrInvalidKeyLength
	case 16, 24, 32:
		return keyByte, nil
	}
}
