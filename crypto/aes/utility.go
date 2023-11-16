package aes

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
