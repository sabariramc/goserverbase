package baseapp

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"sabariram.com/goserverbase/errors"
)

const (
	ContentTypeJSON   = "application/json"
	HeaderContentType = "Content-Type"
)

type HandlerFunction func(*http.Request) (statusCode int, response interface{}, err error)

type HandlerFunctionWithHeader func(*http.Request) (statusCode int, response interface{}, header http.Header, err error)

func (b *BaseApp) JSONResponder(body interface{}, f HandlerFunction) http.HandlerFunc {
	return b.JSONResponderWithHeader(body, func(r *http.Request) (statusCode int, response interface{}, header http.Header, err error) {
		statusCode, response, err = f(r)
		header = make(http.Header)
		header.Add(HeaderContentType, ContentTypeJSON)
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
			if r := recover(); r != nil {
				body := map[string]interface{}{
					"error": "Internal error occcured, if persist contact technical team",
				}
				w.Header().Set(HeaderContentType, ContentTypeJSON)
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(body)
				b.PrintBody(ctx, bodyByte)
				b.log.Error(ctx, "Response-Body", body)
				b.log.Error(ctx, "Recovered in Responder - Error", r)
				b.log.Error(ctx, "Recovered in Responder - StackTrace", string(debug.Stack()))
			}
			b.log.Info(ctx, "Response-StatusCode", statusCode)
			b.log.Info(ctx, "Response-Headers", w.Header())
		}()
		b.PrintHeader(ctx, r.Header)
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			err = json.NewDecoder(r.Body).Decode(inputBody)
			if err != nil {
				statusCode = http.StatusBadRequest
				err = errors.NewCustomError("INVALID_REQUEST_PAYLOAD", "invalid payload", err)
			}
			bodyByte, _ = json.Marshal(body)
			b.log.Debug(ctx, "Request-Body", body)
		}
		if err == nil {
			statusCode, body, headers, err = f(r)
			for key, valueList := range headers {
				for _, value := range valueList {
					w.Header().Set(key, value)
				}
			}
		}
		w.Header().Set(HeaderContentType, ContentTypeJSON)
		w.WriteHeader(statusCode)
		if err != nil {
			body = err.Error()
			b.PrintBody(ctx, bodyByte)
			b.log.Error(ctx, "Response-Body", body)
		}
		if body != nil {
			b.log.Debug(ctx, "Response-Body", body)
			b, ok := body.(string)
			if ok {
				_, err = w.Write([]byte(b))
				if err != nil {
					panic(err)
				}
			} else {
				err = json.NewEncoder(w).Encode(body)
				if err != nil {
					panic(err)
				}
			}

		}
	}
}
