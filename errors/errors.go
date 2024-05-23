package errors

import (
	"encoding/json"
)

// CustomError represents a custom error with additional attributes.
type CustomError struct {
	ErrorCode        string      `json:"errorCode"`
	ErrorMessage     string      `json:"errorMessage"`
	ErrorData        interface{} `json:"errorData"`
	ErrorDescription interface{} `json:"errorDescription"`
	Notify           bool        `json:"-"`
}

// Error returns the JSON representation of the custom error.
func (e *CustomError) Error() string {
	blob, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		e.ErrorData = ErrParse
		e.ErrorDescription = ErrParse
		blob, _ = json.MarshalIndent(e, "", "    ")
	}
	return string(blob)
}

// GetErrorResponse returns the JSON representation of the error response.
func (e *CustomError) GetErrorResponse() ([]byte, error) {
	data := map[string]interface{}{
		"errorCode":        e.ErrorCode,
		"errorMessage":     e.ErrorMessage,
		"errorDescription": e.ErrorDescription,
	}
	blob, err := json.Marshal(data)
	if err != nil {
		e.ErrorDescription = "{\"error\":Internal error occurred, if persist contact technical team}"
		blob, _ = json.Marshal(e)
	}
	return blob, err
}

// HTTPError represents an HTTP error with a custom error.
type HTTPError struct {
	*CustomError `json:",inline"`
	StatusCode   int `json:"statusCode"`
}
