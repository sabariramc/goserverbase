package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, fmt.Errorf("baseapp.GetBytes: %w", err)
	}
	return buf.Bytes(), nil
}
