package baseapp

import (
	"context"
	e "errors"
	"net/http"

	"github.com/sabariramc/goserverbase/v2/errors"
)

func (b *BaseApp) ProcessError(ctx context.Context, stackTrace string, err error, requestData any) (int, []byte) {
	var statusCode int
	var body []byte
	var errorData interface{}
	var errorCode string
	statusCode = http.StatusInternalServerError
	notify := true
	var parseErr error
	var customError *errors.CustomError
	var httpErr *errors.HTTPError
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
		customError = errors.NewCustomError("UNKNOWN", "Unknown error", err, map[string]string{"error": "Internal error occurred, if persist contact technical team"}, true)
		body, parseErr = customError.GetErrorResponse()
		err = customError
	}
	if parseErr != nil {
		b.log.Alert(ctx, "Error occurred during marshal of errors", parseErr)
	}
	b.log.Error(ctx, "Error request data", requestData)
	if errorData == nil {
		errorData = requestData
	} else {
		b.log.Error(ctx, "Error data", errorData)
	}
	if notify && b.errorNotifier != nil {
		if statusCode >= 500 {
			b.errorNotifier.Send5XX(ctx, errorCode, err, stackTrace, errorData)
		} else {
			b.errorNotifier.Send4XX(ctx, errorCode, err, stackTrace, errorData)
		}
	}

	return statusCode, body
}