package baseapp

import (
	"context"
	e "errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/sabariramc/goserverbase/v6/errors"
)

// PanicRecovery recovers from panics, logs the panic details, and returns the stack trace and error.
//
// This function should be used within a defer statement to capture any panics that occur during execution.
// It logs the panic details, captures the stack trace, and converts the panic to an error if it is not already one.
func (b *BaseApp) PanicRecovery(ctx context.Context, rec any) (string, error) {
	stackTrace := string(debug.Stack())
	err, ok := rec.(error)
	if !ok {
		b.log.Error(ctx, "Recovered - Panic", rec)
		err = fmt.Errorf("panic: %v", rec)
	} else {
		b.log.Error(ctx, "Recovered - Panic")
	}
	return stackTrace, err
}

// ProcessError processes an error, logs it, and returns the HTTP status code and error response body.
//
// This function determines the type of the provided error, logs the appropriate details, and constructs an HTTP response.
// It handles custom errors, HTTP errors, and general errors, providing the necessary error response and logging stack traces if available.
// Notifications are sent if configured, based on the error type and status code.
func (b *BaseApp) ProcessError(ctx context.Context, stackTrace string, err error) (int, []byte) {
	var statusCode int
	var body []byte
	var errorData interface{}
	var errorCode string
	statusCode = http.StatusInternalServerError
	notify := true
	var parseErr error
	var customErrorPtr *errors.CustomError
	var httpErrPtr *errors.HTTPError
	var customError errors.CustomError
	var httpErr errors.HTTPError
	if e.As(err, &httpErrPtr) || e.As(err, &httpErr) {
		if httpErrPtr == nil {
			httpErrPtr = &httpErr
		}
		statusCode = httpErrPtr.StatusCode
		notify = httpErrPtr.Notify
		body, parseErr = httpErrPtr.GetErrorResponse()
		errorCode = httpErrPtr.ErrorCode
		errorData = httpErrPtr.ErrorData
	} else if e.As(err, &customErrorPtr) || e.As(err, &customError) {
		if customErrorPtr == nil {
			customErrorPtr = &customError
		}
		statusCode = http.StatusInternalServerError
		notify = customErrorPtr.Notify
		body, parseErr = customErrorPtr.GetErrorResponse()
		errorData = customErrorPtr.ErrorData
		errorCode = customErrorPtr.ErrorCode
	} else {
		statusCode = http.StatusInternalServerError
		customErrorPtr = &errors.CustomError{ErrorCode: "com.base.internalServerError", ErrorMessage: "Unknown error", ErrorDescription: map[string]string{"error": "Internal error occurred, if persist contact technical team"}, Notify: true}
		body, parseErr = customErrorPtr.GetErrorResponse()
		err = customErrorPtr
	}
	if parseErr != nil {
		b.log.Error(ctx, "Error occurred during marshal of errors", parseErr)
	}
	b.log.Error(ctx, "Error", err)
	if stackTrace != "" {
		b.log.Error(ctx, "Stack trace", stackTrace)
	}
	if notify && b.notifier != nil {
		if statusCode >= 500 {
			b.notifier.Notify5XX(ctx, errorCode, err, stackTrace, errorData)
		} else {
			b.notifier.Notify4XX(ctx, errorCode, err, stackTrace, errorData)
		}
	}
	return statusCode, body
}
