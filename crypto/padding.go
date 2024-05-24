// Package crypto defines common interfaces for encryption, decryption, and data padding.
package crypto

// Padder provides methods for adding and removing padding from byte slices.
type Padder interface {
	// Pad applies padding to the given plainData and returns the padded data.
	Pad(plainData []byte) []byte
	// UnPad removes padding from the given encryptedData and returns the unpadded data.
	UnPad(encryptedData []byte) []byte
}
