package utils

import (
	jsoniter "github.com/json-iterator/go"
)

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
