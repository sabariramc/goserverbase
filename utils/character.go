package utils

import (
	"unicode"
	"unicode/utf8"
)

// IsASCII checks if the provided string contains only ASCII characters.
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// IsUTF8 checks if the provided string is a valid UTF-8 encoded string.
func IsUTF8(s string) bool {
	return utf8.ValidString(s)
}
