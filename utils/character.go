package utils

import (
	"unicode"
	"unicode/utf8"
)

func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func IsUTF8(s string) bool {
	return utf8.ValidString(s)
}
