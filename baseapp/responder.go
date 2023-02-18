package baseapp

import (
	"encoding/json"
	e "errors"
	"net/http"
	"runtime/debug"

	"github.com/sabariramc/goserverbase/constant"
	"github.com/sabariramc/goserverbase/errors"
)

type HandlerFunction func(*http.Request) (statusCode int, response interface{}, err error)

type HandlerFunctionWithHeader func(*http.Request) (statusCode int, response interface{}, header http.Header, err error)

func (b *BaseApp) JSONResponder(body interface{}, f HandlerFunction) http.HandlerFunc {
	return b.JSONResponderWithHeader(body, func(r *http.Request) (statusCode int, response interface{}, header http.Header, err error) {
		statusCode, response, err = f(r)
		header = make(http.Header)
		header.Add(constant.HeaderContentType, constant.ContentTypeJSON)
		return
	})
}

func (b *BaseApp) JSONResponderWithHeader(inputBody interface{}, f HandlerFunctionWithHeader) http.HandlerFunc {
	var bodyByte []byte
	return func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		var body interface{}
		var err error
		var headers http.Header
		ctx := r.Context()
		defer func() {
			if rec := recover(); rec != nil {
				body := map[string]string{"error": "Internal error occcured, if persist contact technical team"}
				w.Header().Set(constant.HeaderContentType, constant.ContentTypeJSON)
				w.WriteHeader(http.StatusInternalServerError)
				res, _ := json.Marshal(body)
				w.Write([]byte(res))
				b.PrintBody(ctx, bodyByte)
				statusCode = http.StatusInternalServerError
				if b.errorNotifier != nil {
					b.errorNotifier.Send(ctx, r, rec.(error))
				}
				b.log.Error(ctx, "Recovered in Responder - Error", rec)
				b.log.Error(ctx, "Recovered in Responder - StackTrace", string(debug.Stack()))
				b.log.Error(ctx, "Response-Body", body)
			}
			b.log.Info(ctx, "Response-StatusCode", statusCode)
			b.log.Info(ctx, "Response-Headers", w.Header())
		}()
		b.PrintRequest(ctx, r)
		if inputBody != nil && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) {
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			err = decoder.Decode(inputBody)
			if err != nil {
				err = errors.NewHTTPClientError(http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "invalid payload")
			}
			bodyByte, _ = json.Marshal(inputBody)
			b.log.Debug(ctx, "Request-Body", inputBody)
		}
		if err == nil {
			statusCode, body, headers, err = f(r)
			for key, valueList := range headers {
				for _, value := range valueList {
					w.Header().Set(key, value)
				}
			}
		}
		if err != nil {
			notify := false
			var custErr *errors.CustomError
			var httpErr *errors.HTTPError
			if e.As(err, &httpErr) {
				statusCode = httpErr.ErrorStatusCode
				notify = httpErr.Notify
				body = httpErr.GetErrorResponse()

			} else if e.As(err, &custErr) {
				statusCode = http.StatusInternalServerError
				notify = custErr.Notify
				body = custErr.GetErrorResponse()
			} else {
				statusCode = http.StatusInternalServerError
				custErr = errors.NewCustomError("UNKNOWN", "Unknown error", err, true)
				body = custErr.GetErrorResponse()
				err = custErr
			}
			if notify && b.errorNotifier != nil {
				b.errorNotifier.Send(ctx, r, err)
			}
			b.PrintBody(ctx, bodyByte)
			b.log.Error(ctx, "Response-Body", body)
		}
		w.Header().Set(constant.HeaderContentType, constant.ContentTypeJSON)
		w.WriteHeader(statusCode)
		if body != nil {
			b.log.Debug(ctx, "Response-Body", body)
			bodyData, ok := body.(string)
			if ok {
				_, err = w.Write([]byte(bodyData))
				if err != nil {
					b.log.Emergency(ctx, "BaseApp.JSONResponderWithHeader", err, err)
				}
			} else {
				err = json.NewEncoder(w).Encode(body)
				if err != nil {
					b.log.Emergency(ctx, "BaseApp.JSONResponderWithHeader", err, err)
				}
			}

		}
	}
}
