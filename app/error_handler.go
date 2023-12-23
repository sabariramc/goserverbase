package baseapp

import (
	"context"
	"encoding/json"
	e "errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/sabariramc/goserverbase/v4/errors"
)

func (b *BaseApp) PanicRecovery(ctx context.Context, rec any) (int, []byte) {
	stackTrace := string(debug.Stack())
	b.log.Error(ctx, "Recovered - Panic", rec)
	b.log.Error(ctx, "Recovered - StackTrace", stackTrace)
	err, ok := rec.(error)
	if !ok {
		blob, _ := json.Marshal(rec)
		err = fmt.Errorf("non error panic: %v", string(blob))
	}
	return b.ProcessError(ctx, stackTrace, err)
}

func (b *BaseApp) ProcessError(ctx context.Context, stackTrace string, err error) (int, []byte) {
	var statusCode int
	var body []byte
	var errorData interface{}
	var errorCode string
	statusCode = http.StatusInternalServerError
	notify := true
	var parseErr error
	var customError *errors.CustomError
	var httpErr *errors.HTTPError
	b.log.Error(ctx, "Error", err)
	if e.As(err, &httpErr) {
		statusCode = httpErr.ErrorStatusCode
		notify = httpErr.Notify
		body, parseErr = httpErr.GetErrorResponse()
		errorCode = httpErr.ErrorCode
		errorData = httpErr.ErrorData
	} else if e.As(err, &customError) {
		statusCode = http.StatusInternalServerError
		notify = customError.Notify
		body, parseErr = customError.GetErrorResponse()
		errorData = customError.ErrorData
	} else {
		statusCode = http.StatusInternalServerError
		customError = errors.NewCustomError("UNKNOWN", "Unknown error", nil, map[string]string{"error": "Internal error occurred, if persist contact technical team"}, true, err)
		body, parseErr = customError.GetErrorResponse()
		err = customError
	}
	if parseErr != nil {
		b.log.Error(ctx, "Error occurred during marshal of errors", parseErr)
	}
	b.log.Error(ctx, "Wrapped Error", err)
	if notify && b.errorNotifier != nil {
		if statusCode >= 500 {
			b.errorNotifier.Send5XX(ctx, errorCode, err, stackTrace, errorData)
		} else {
			b.errorNotifier.Send4XX(ctx, errorCode, err, stackTrace, errorData)
		}
	}
	return statusCode, body
}
