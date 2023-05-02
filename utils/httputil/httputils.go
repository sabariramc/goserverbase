package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sabariramc/goserverbase/v2/log"
)

var ErrResponseUnmarshal = fmt.Errorf("http.do.responseBodyMarshall")

type LogConfig struct {
	MaxContentLength int64
}

var DefaultLogConfig = &LogConfig{
	MaxContentLength: 1024,
}

type HttpClient struct {
	*retryablehttp.Client
	log       *log.Logger
	LogConfig LogConfig
}

func NewDefaultHttpClient(log *log.Logger) *HttpClient {
	return NewHttpClient(time.Minute, log, *DefaultLogConfig)
}

func NewHttpClient(timeout time.Duration, log *log.Logger, logConfig LogConfig) *HttpClient {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	c := &HttpClient{Client: retryablehttp.NewClient(), log: log, LogConfig: logConfig}
	c.HTTPClient.Transport = t
	return c
}

func (h *HttpClient) Validator(resBody interface{}) error {
	if resBody != nil {
		v := reflect.ValueOf(resBody)
		if v.Kind() != reflect.Ptr {
			return fmt.Errorf("`resBody` is not ptr; is %T", resBody)
		}
	}
	return nil
}

func (h *HttpClient) Encode(ctx context.Context, data interface{}) (interface{}, io.ReadCloser, error) {
	var print interface{}
	var body io.ReadCloser
	if data != nil {
		switch v := data.(type) {
		case string:
			if int64(len(v)) <= h.LogConfig.MaxContentLength {
				print = v
			}
			body = io.NopCloser(bytes.NewReader([]byte(v)))
		case []byte:
			if int64(len(v)) <= h.LogConfig.MaxContentLength {
				print = string(v)
			}
			body = io.NopCloser(bytes.NewReader(v))
		case io.ReadCloser:
			body = v
		case io.Reader:
			body = io.NopCloser(v)
		default:
			rv := reflect.ValueOf(v)
			if rv.Kind() != reflect.Struct {
				return nil, nil, fmt.Errorf("payload is not a struct/string/[]byte/io.ReadCloser/io.Reader object")
			}
			if int64(rv.Type().Size()) <= h.LogConfig.MaxContentLength {
				print = data
			}
			buff := &bytes.Buffer{}
			err := json.NewEncoder(buff).Encode(&data)
			body = io.NopCloser(buff)
			if err != nil {
				h.log.Error(ctx, "Http.Do.PayloadEncoding", err)
				return nil, nil, fmt.Errorf("http.do.PayloadEncoding: %w", err)
			}
		}
	}
	return print, body, nil
}

func (h *HttpClient) Decode(ctx context.Context, body []byte, data interface{}) (io.ReadCloser, error) {
	if data != nil {
		switch v := data.(type) {
		case *string:
			*v = string(body)
		case *[]byte:
			*v = body
		default:
			err := json.Unmarshal(body, data)
			if err != nil {
				newBuff := io.NopCloser(bytes.NewReader(body))
				h.log.Error(ctx, "http.Do.responseBodyMarshall", err)
				return newBuff, fmt.Errorf("%w : %w", ErrResponseUnmarshal, err)
			}
		}
		return nil, nil
	}
	newBuff := io.NopCloser(bytes.NewReader(body))
	return newBuff, nil
}

func (h *HttpClient) Do(ctx context.Context, method, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	err := h.Validator(resBody)
	if err != nil {
		return nil, err
	}
	reqPrint, body, err := h.Encode(ctx, reqBody)
	if err != nil {
		return nil, err
	}
	req, err := retryablehttp.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		h.log.Error(ctx, "Http.Do.RequestCreation", err)
		return nil, fmt.Errorf("http.do.requestCreation: %w", err)
	}
	log.SetCorrelationHeader(ctx, req.Request)
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	if reqPrint == nil {
		h.log.Debug(ctx, "Request payload is not printed : either it is a interface or violates MaxContentLength", err)
	}
	h.log.Debug(ctx, "Request", map[string]interface{}{
		"method":  method,
		"url":     url,
		"headers": req.Header,
		"body":    reqPrint,
	})
	if body != nil {
		req.Body = body
	}
	r, err := h.Client.Do(req)
	if err != nil {
		h.log.Error(ctx, "http.do.networkCall", err)
		return nil, fmt.Errorf("http.do.networkCall: %w", err)
	}
	logRes := map[string]interface{}{
		"statusCode": r.StatusCode,
	}
	defer body.Close()
	blobBody, _ := ioutil.ReadAll(body)
	if r.StatusCode == http.StatusNoContent || r.ContentLength <= 0 {
		h.log.Debug(ctx, "Response", logRes)
		return r, nil
	}
	if r.ContentLength > h.LogConfig.MaxContentLength {
		h.log.Debug(ctx, "Response payload is not printed - violets MaxContentLength config", nil)
	} else {
		logRes["body"] = string(blobBody)
	}
	h.log.Debug(ctx, "Response", logRes)
	r.Body, err = h.Decode(ctx, blobBody, resBody)
	return r, err
}
