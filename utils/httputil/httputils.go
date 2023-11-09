package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sabariramc/goserverbase/v3/log"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
)

var ErrResponseUnmarshal = fmt.Errorf("http.do.responseBodyMarshall")

type CheckRetry func(ctx context.Context, resp *http.Response, err error) (bool, error)

type Backoff func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration

type LogConfig struct {
	MaxContentLength int64
}

var DefaultLogConfig = &LogConfig{
	MaxContentLength: 10240,
}

type HttpClient struct {
	*http.Client
	log          *log.Logger
	LogConfig    LogConfig
	RetryMax     int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
	CheckRetry   CheckRetry
	Backoff      Backoff
}

func NewDefaultHttpClient(log *log.Logger) *HttpClient {
	return NewHttpClient(log, *DefaultLogConfig, 4, time.Second*1, time.Second*5)
}

func NewHttpClient(log *log.Logger, logConfig LogConfig, retryMax int, retryWaitMin, retryWaitMax time.Duration) *HttpClient {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	c := &HttpClient{Client: ddtrace.WrapClient(&http.Client{Transport: t}), log: log.NewResourceLogger("HttpClient"), LogConfig: logConfig, RetryMax: retryMax, RetryWaitMin: retryWaitMin, RetryWaitMax: retryWaitMax, CheckRetry: retryablehttp.DefaultRetryPolicy, Backoff: retryablehttp.DefaultBackoff}
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

func (h *HttpClient) Encode(ctx context.Context, data interface{}) (interface{}, io.Reader, error) {
	var print interface{}
	var body io.Reader
	if data != nil {
		switch v := data.(type) {
		case string:
			if int64(len(v)) <= h.LogConfig.MaxContentLength {
				print = v
			}
			body = bytes.NewReader([]byte(v))
		case []byte:
			if int64(len(v)) <= h.LogConfig.MaxContentLength {
				print = string(v)
			}
			body = bytes.NewReader(v)
		case io.ReadCloser:
			body = v
		case io.Reader:
			body = v
		default:
			rv := reflect.ValueOf(v)
			if int64(rv.Type().Size()) <= h.LogConfig.MaxContentLength {
				print = data
			}
			buff := &bytes.Buffer{}
			err := json.NewEncoder(buff).Encode(&data)
			if err != nil {
				return nil, nil, fmt.Errorf("HttpClient.Encode: error in encoding payload: %w", err)
			}
			body = buff
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
				return newBuff, fmt.Errorf("HttpClient:Decode: %w : %w", ErrResponseUnmarshal, err)
			}
		}
		return nil, nil
	}
	newBuff := io.NopCloser(bytes.NewReader(body))
	return newBuff, nil
}

func (h *HttpClient) Get(ctx context.Context, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	return h.Call(ctx, http.MethodGet, url, reqBody, resBody, headers)
}

func (h *HttpClient) Post(ctx context.Context, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	return h.Call(ctx, http.MethodPost, url, reqBody, resBody, headers)
}

func (h *HttpClient) Put(ctx context.Context, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	return h.Call(ctx, http.MethodPut, url, reqBody, resBody, headers)
}

func (h *HttpClient) Patch(ctx context.Context, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	return h.Call(ctx, http.MethodPatch, url, reqBody, resBody, headers)
}

func (h *HttpClient) Delete(ctx context.Context, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	return h.Call(ctx, http.MethodDelete, url, reqBody, resBody, headers)
}

func (h *HttpClient) Call(ctx context.Context, method, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	err := h.Validator(resBody)
	if err != nil {
		return nil, err
	}
	var req *http.Request
	reqPrint, body, err := h.Encode(ctx, reqBody)
	if err != nil {
		return nil, err
	}
	req, err = http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("HttpClient.Call: error in request creation: %w", err)
	}
	log.SetCorrelationHeader(ctx, req)
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	if reqBody != nil && reqPrint == nil {
		h.log.Debug(ctx, "Request payload is not printed : either it is a interface or violates MaxContentLength", err)
	}
	h.log.Debug(ctx, "Request", map[string]interface{}{
		"method":  method,
		"url":     url,
		"headers": req.Header,
		"body":    reqPrint,
	})
	r, err := h.Do(req)
	if err != nil {
		return r, fmt.Errorf("HttpClient.Call: error in network call: %w", err)
	} else if r == nil {
		return r, err
	}
	logRes := map[string]interface{}{
		"statusCode": r.StatusCode,
		"headers":    r.Header,
	}
	defer r.Body.Close()
	resBlob, _ := io.ReadAll(r.Body)
	if r.StatusCode == http.StatusNoContent || len(resBlob) == 0 {
		h.log.Debug(ctx, "Response", logRes)
		return r, nil
	}
	if r.StatusCode > 299 {
		return r, fmt.Errorf("HttpClient.Call: Non 2xx status: %v", r.StatusCode)
	}
	r.Body, err = h.Decode(ctx, resBlob, resBody)
	if r.ContentLength > h.LogConfig.MaxContentLength {
		h.log.Debug(ctx, "Response payload is not printed - violets MaxContentLength config", nil)
	} else {
		if err != nil {
			logRes["body"] = string(resBlob)
		} else {
			logRes["body"] = resBody
		}
	}
	h.log.Debug(ctx, "Response", logRes)
	return r, err
}

func (h *HttpClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var attempt int
	var shouldRetry bool
	var doErr, respErr, checkErr error
	var reqBody, resBlob []byte
	if req.ContentLength > 0 {
		reqBody, _ = io.ReadAll(req.Body)
	}
	for i := 0; ; i++ {
		doErr, respErr = nil, nil
		attempt++
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		resp, doErr = h.Client.Do(req)
		shouldRetry, checkErr = h.CheckRetry(req.Context(), resp, doErr)
		if !shouldRetry || checkErr != nil {
			break
		}
		remain := h.RetryMax - i
		if remain <= 0 {
			break
		}
		wait := h.Backoff(h.RetryWaitMin, h.RetryWaitMax, i, resp)
		if resp != nil && resp.ContentLength > 0 {
			defer resp.Body.Close()
			resBlob, _ = io.ReadAll(resp.Body)
			h.log.Notice(req.Context(), fmt.Sprintf("request failed with status code %v retry %v of %v in %vms", resp.StatusCode, i+1, h.RetryMax, wait.Milliseconds()), string(resBlob))
		} else if doErr != nil {
			h.log.Notice(req.Context(), fmt.Sprintf("request failed with error - retry %v of %v in %vms", i+1, h.RetryMax, wait.Milliseconds()), doErr)
		} else {
			h.log.Notice(req.Context(), fmt.Sprintf("request failed - retry %v of %v in %vms", i+1, h.RetryMax, wait.Milliseconds()), nil)
		}
		timer := time.NewTimer(wait)
		select {
		case <-req.Context().Done():
			timer.Stop()
			h.Client.CloseIdleConnections()
			return nil, req.Context().Err()
		case <-timer.C:
		}

	}

	// this is the closest we have to success criteria
	if doErr == nil && respErr == nil && checkErr == nil && !shouldRetry {
		return resp, nil
	}

	defer h.Client.CloseIdleConnections()

	var err error
	if checkErr != nil {
		err = checkErr
	} else if respErr != nil {
		err = respErr
	} else {
		err = doErr
	}

	return resp, err

}
