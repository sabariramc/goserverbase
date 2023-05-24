package utils

import (
	"bytes"

	"encoding/json"
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

func Serialize(val interface{}) (*string, error) {
	blob, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("utils.GetString : %w", err)
	}
	str := string(blob)
	return &str, nil
}

func LenientJsonTransformer(src interface{}, dest interface{}) error {
	blob, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("LenientJsonTransformer encoding: %w", err)
	}
	err = json.Unmarshal(blob, dest)
	if err != nil {
		return fmt.Errorf("LenientJsonTransformer decoding: %w", err)
	}
	return nil
}

func StrictJsonTransformer(src interface{}, dest interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(src)
	if err != nil {
		return fmt.Errorf("StrictJsonTransformer encoding: %w", err)
	}
	decoder := json.NewDecoder(&buf)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(dest)
	if err != nil {
		return fmt.Errorf("StrictJsonTransformer decoding: %w", err)
	}
	return nil
}

type CustomJsonTagHandler struct {
	jsoniter.API
}

func NewCustomJsonTagHandler(tag string) *CustomJsonTagHandler {
	return &CustomJsonTagHandler{
		API: jsoniter.Config{
			EscapeHTML:             true,
			SortMapKeys:            true,
			ValidateJsonRawMessage: true,
			TagKey:                 tag,
		}.Froze(),
	}
}

var HeaderJson = NewCustomJsonTagHandler("header")
var BodyJson = NewCustomJsonTagHandler("body")
