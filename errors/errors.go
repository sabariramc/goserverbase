package errors

import (
	"encoding/json"
)

type CustomError struct {
	ErrorData    interface{} `json:"errorData"`
	ErrorMessage string      `json:"errorMessage"`
	ErrorCode    string      `json:"errorCode"`
}

func (e *CustomError) Error() string {
	blob, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		e.ErrorData = ParseErrorMsg
		blob, _ = json.MarshalIndent(e, "", "    ")
	}
	return string(blob)
}

func NewCustomError(errorCode, errorMessage string, errorData interface{}) *CustomError {
	if v, ok := errorData.(error); ok {
		errorData = v.Error()
	}
	return &CustomError{ErrorCode: errorCode, ErrorMessage: errorMessage, ErrorData: errorData}
}
