package utils

import (
	"encoding/gob"
	"fmt"
	"io"
)

func Encode(src interface{}, dest io.Writer) error {
	enc := gob.NewEncoder(dest)
	err := enc.Encode(src)
	if err != nil {
		return fmt.Errorf("utils.Encode: %w", err)
	}
	return nil
}

func Decode(src io.Reader, dest interface{}) error {
	enc := gob.NewDecoder(src)
	err := enc.Decode(dest)
	if err != nil {
		return fmt.Errorf("utils.Decode: %w", err)
	}
	return nil
}
