package utils

import (
	"encoding/gob"
	"fmt"
	"io"
)

// Encode encodes the source interface and writes it to the destination writer using gob encoding.
func Encode(src interface{}, dest io.Writer) error {
	enc := gob.NewEncoder(dest)
	if err := enc.Encode(src); err != nil {
		return fmt.Errorf("utils.Encode: %w", err)
	}
	return nil
}

// Decode decodes data from the source reader using gob decoding and stores it in the destination interface.
func Decode(src io.Reader, dest interface{}) error {
	dec := gob.NewDecoder(src)
	if err := dec.Decode(dest); err != nil {
		return fmt.Errorf("utils.Decode: %w", err)
	}
	return nil
}
