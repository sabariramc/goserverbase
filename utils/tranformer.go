package utils

import (
	"bytes"
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// JSONTransformer copies fields from src (a map or struct object) to dest (another object).
// The field mapping is based on keys (if src is a map) or JSON struct tags (if src is a struct).
// This function first marshals src into a JSON byte slice and then unmarshals it into dest.
// Returns an error if marshalling or unmarshalling fails.
func JSONTransformer(src interface{}, dest interface{}) error {
	blob, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("JSONTransformer: error encoding content: %w", err)
	}
	err = json.Unmarshal(blob, dest)
	if err != nil {
		return fmt.Errorf("JSONTransformer: error decoding content: %w", err)
	}
	return nil
}

/*
StrictJSONTransformer copies fields from src (a map or struct object) to dest (another object).
It enforces strict field matching, throwing an error if there are keys in src that do not have a corresponding field in dest.
This function first encodes src into a JSON byte slice and then decodes it into dest using a decoder that disallows unknown fields.
Returns an error if encoding or decoding fails.
*/
func StrictJSONTransformer(src interface{}, dest interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(src)
	if err != nil {
		return fmt.Errorf("StrictJSONTransformer: error encoding source: %w", err)
	}
	decoder := json.NewDecoder(&buf)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(dest)
	if err != nil {
		return fmt.Errorf("StrictJSONTransformer: error decoding content: %w", err)
	}
	return nil
}
