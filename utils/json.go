package utils

import (
	jsoniter "github.com/json-iterator/go"
)

// CustomJSONTagHandler is a wrapper around jsoniter.API to handle JSON encoding/decoding
// with a custom struct tag.
type CustomJSONTagHandler struct {
	jsoniter.API
}

// NewCustomJSONTagHandler creates a new CustomJSONTagHandler with a custom struct tag.
// It sets up the JSON configuration to escape HTML, sort map keys, validate JSON raw messages,
// and use the specified tag key.
func NewCustomJSONTagHandler(tag string) *CustomJSONTagHandler {
	return &CustomJSONTagHandler{
		API: jsoniter.Config{
			EscapeHTML:             true,
			SortMapKeys:            true,
			ValidateJsonRawMessage: true,
			TagKey:                 tag,
		}.Froze(),
	}
}

// HeaderJSON is a CustomJSONTagHandler instance with the custom tag "header".
var HeaderJSON = NewCustomJSONTagHandler("header")

// BodyJSON is a CustomJSONTagHandler instance with the custom tag "body".
var BodyJSON = NewCustomJSONTagHandler("body")
