package utils

import (
	jsoniter "github.com/json-iterator/go"
)

type CustomJSONTagHandler struct {
	jsoniter.API
}

/*
NewCustomJSONTagHandler creates new JSON handler with custom struct tag
*/
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

var HeaderJSON = NewCustomJSONTagHandler("header")
var BodyJSON = NewCustomJSONTagHandler("body")
