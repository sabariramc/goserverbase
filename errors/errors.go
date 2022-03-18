package errors

import (
	"encoding/json"
	"net/http"
)

type ErrorData map[string]string

type Error struct {
	StatusCode   int               `json:"-"`
	ErrorData    map[string]string `json:"errorData"`
	ErrorMessage string            `json:"errorMessage"`
	ErrorCode    string            `json:"errorCode"`
}

func (e *Error) Error() string {
	blob, _ := json.Marshal(e)
	return string(blob)
}

func NewHTTPError(statusCode int, errorMessage string, errorData ErrorData) *Error {
	return &Error{StatusCode: statusCode, ErrorMessage: errorMessage, ErrorData: errorData, ErrorCode: http.StatusText(statusCode)}
}
