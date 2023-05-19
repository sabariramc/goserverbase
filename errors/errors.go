package errors

import (
	"encoding/json"
	"net/http"
)

type CustomError struct {
	ErrorData        interface{} `json:"errorData"`
	ErrorMessage     string      `json:"errorMessage"`
	ErrorDescription interface{} `json:"errorDescription"`
	ErrorCode        string      `json:"errorCode"`
	Notify           bool        `json:"-"`
}

func (e *CustomError) Error() string {
	blob, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		e.ErrorData = ErrParse
		e.ErrorDescription = ErrParse
		blob, _ = json.MarshalIndent(e, "", "    ")
	}
	return string(blob)
}

func (e *CustomError) GetErrorResponse() ([]byte, error) {
	data := map[string]any{
		"errorMessage":     e.ErrorMessage,
		"errorCode":        e.ErrorCode,
		"errorDescription": e.ErrorDescription,
	}
	blob, err := json.Marshal(data)
	if err != nil {
		e.ErrorDescription = "{\"error\":Internal error occurred, if persist contact technical team}"
		blob, _ = json.Marshal(e)
	}
	return blob, err
}

func NewCustomError(errorCode, errorMessage string, errorData interface{}, errorDescription interface{}, notify bool) *CustomError {
	return &CustomError{ErrorCode: errorCode, ErrorMessage: errorMessage, ErrorData: errorData, Notify: notify, ErrorDescription: errorDescription}
}

type HTTPError struct {
	CustomError
	ErrorStatusCode int `json:""`
}

func NewHTTPError(statusCode int, errorCode, errorMessage string, errorData interface{}, errorDescription interface{}, notify bool) *HTTPError {
	if errorCode == "" {
		errorCode = http.StatusText(statusCode)
	}
	err := NewCustomError(errorCode, errorMessage, errorData, errorDescription, notify)
	return &HTTPError{CustomError: *err, ErrorStatusCode: statusCode}
}

func NewHTTPClientError(statusCode int, errorCode, errorMessage string, errorData interface{}, errorDescription interface{}) *HTTPError {
	return NewHTTPError(statusCode, errorCode, errorMessage, errorData, errorDescription, false)
}

func NewHTTPServerError(statusCode int, errorCode, errorMessage string, errorData interface{}, errorDescription interface{}) *HTTPError {
	return NewHTTPError(statusCode, errorCode, errorMessage, errorData, errorDescription, true)
}
