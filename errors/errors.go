package errors

import (
	"encoding/json"
	"net/http"
)

type CustomError struct {
	ErrorData        interface{} `json:"-"`
	ErrorMessage     string      `json:"errorMessage"`
	ErrorDescription interface{} `json:"errorDescription"`
	ErrorCode        string      `json:"errorCode"`
	Notify           bool        `json:"-"`
}

func (e *CustomError) Error() string {
	blob, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		e.ErrorData = ParseErrorMsg
		blob, _ = json.MarshalIndent(e, "", "    ")
	}
	return string(blob)
}

func (e *CustomError) GetErrorResponse() string {
	blob, err := json.Marshal(e)
	if err != nil {
		e.ErrorData = ParseErrorMsg
		blob, _ = json.Marshal(e)
	}
	return string(blob)
}

func NewCustomError(errorCode, errorMessage string, errorData interface{}, errorDescription interface{}, notify bool) *CustomError {
	if v, ok := errorData.(error); ok {
		errorData = v.Error()
	}
	return &CustomError{ErrorCode: errorCode, ErrorMessage: errorMessage, ErrorData: errorData, Notify: notify, ErrorDescription: errorDescription}
}

type HTTPError struct {
	CustomError
	ErrorStatusCode int `json:""`
}

func NewHTTPError(statusCode int, errorMessage string, errorData interface{}, errorDescription interface{}, notify bool) *HTTPError {
	errorCode := http.StatusText(statusCode)
	err := NewCustomError(errorCode, errorMessage, errorData, errorDescription, notify)
	return &HTTPError{CustomError: *err, ErrorStatusCode: statusCode}
}

func NewHTTPClientError(statusCode int, errorMessage string, errorData interface{}, errorDescription interface{}) *HTTPError {
	return NewHTTPError(statusCode, errorMessage, errorData, errorDescription, false)
}

func NewHTTPServerError(statusCode int, errorMessage string, errorData interface{}, errorDescription interface{}) *HTTPError {
	return NewHTTPError(statusCode, errorMessage, errorData, errorDescription, true)
}
