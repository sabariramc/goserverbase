package utils

import (
	"bytes"

	"encoding/json"
	"fmt"
)

func Serialize(val interface{}) (*string, error) {
	blob, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("utils.GetString : %w", err)
	}
	str := string(blob)
	return &str, nil
}

func StrictJsonTransformer(src interface{}, dest interface{}) error {
	return jsonTransformer(src, dest, true)
}

func LenientJsonTransformer(src interface{}, dest interface{}) error {
	return jsonTransformer(src, dest, false)
}

func jsonTransformer(src interface{}, dest interface{}, strict bool) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(src)
	if err != nil {
		return fmt.Errorf("JsonTransformer encoding: %w", err)
	}
	decoder := json.NewDecoder(&buf)
	if strict {
		decoder.DisallowUnknownFields()
	}
	err = decoder.Decode(dest)
	if err != nil {
		return fmt.Errorf("JsonTransformer decoding: %w", err)
	}
	return nil
}
