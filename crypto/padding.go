package crypto

type Padder interface {
	Pad(plainData []byte) []byte
	UnPad(encryptedData []byte) []byte
}
