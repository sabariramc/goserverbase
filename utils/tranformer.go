package utils

import (
	"bytes"
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// JSONTransformer copies fields from src(map / struct object) to dest object based on key(incase of map) or json struct tag
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

/*StrictJSONTransformer copies fields from src(map/ struct object) to dest object throws error if there are keys in src that doesn't have slot in dest object*/
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
