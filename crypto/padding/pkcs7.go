// Package padding implements PKCS7 padding.
package padding

import "bytes"

// PKCS7 implements PKCS7 padding for block ciphers.
type PKCS7 struct {
	blockSize int
}

// NewPKCS7 creates a new PKCS7 instance with the specified block size.
func NewPKCS7(blockSize int) *PKCS7 {
	return &PKCS7{blockSize: blockSize}
}

// UnPad removes PKCS7 padding from encryptedData and returns the unpadded data.
func (p *PKCS7) UnPad(encryptedData []byte) []byte {
	length := len(encryptedData)
	unPadding := int(encryptedData[length-1])
	return encryptedData[:(length - unPadding)]
}

// Pad adds PKCS7 padding to plainData and returns the padded data.
func (p *PKCS7) Pad(plainData []byte) []byte {
	padding := p.blockSize - len(plainData)%p.blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plainData, padText...)
}
