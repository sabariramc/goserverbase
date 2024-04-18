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
	OriginalError    string      `json:"error"`
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

func NewCustomError(errorCode, errorMessage string, errorData interface{}, errorDescription interface{}, notify bool, err error) *CustomError {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}
	return &CustomError{ErrorCode: errorCode, ErrorMessage: errorMessage, ErrorData: errorData, Notify: notify, ErrorDescription: errorDescription, OriginalError: errStr}
}

type HTTPError struct {
	CustomError
	StatusCode int `json:"statusCode"`
}

func NewHTTPError(statusCode int, errorCode, errorMessage string, errorData interface{}, errorDescription interface{}, notify bool, err error) *HTTPError {
	if errorCode == "" {
		errorCode = http.StatusText(statusCode)
	}
	custErr := NewCustomError(errorCode, errorMessage, errorData, errorDescription, notify, err)
	return &HTTPError{CustomError: *custErr, StatusCode: statusCode}
}

func NewHTTPClientError(statusCode int, errorCode, errorMessage string, errorData interface{}, errorDescription interface{}, err error) *HTTPError {
	return NewHTTPError(statusCode, errorCode, errorMessage, errorData, errorDescription, false, err)
}

func NewHTTPServerError(statusCode int, errorCode, errorMessage string, errorData interface{}, errorDescription interface{}, err error) *HTTPError {
	return NewHTTPError(statusCode, errorCode, errorMessage, errorData, errorDescription, true, err)
}
