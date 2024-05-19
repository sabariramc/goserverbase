// Package httputil implements utility for httpClient with request retry and optional tracing
package httputil

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"reflect"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"github.com/sabariramc/goserverbase/v6/log"
	"golang.org/x/net/http2"
)

var ErrResponseUnmarshal = fmt.Errorf("error marshalling response body")
var ErrResponseFromUpstream = fmt.Errorf("HttpClient.Call: non 2xx status")

type CheckRetry func(ctx context.Context, resp *http.Response, err error) (bool, error)

type Backoff func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration

/*
HTTPClient extends http.Client with following
 1. Default exponential backoff and retry mechanism
 2. Request body can be any object
 3. Optionally decode success response to the passed object
 4. Additional debug logging
 5. Custom tracing interface
*/
type HTTPClient struct {
	*http.Client
	log          log.Log
	retryMax     uint
	minRetryWait time.Duration
	maxRetryWait time.Duration
	checkRetry   CheckRetry
	backoff      Backoff
	tr           Tracer
}

type Tracer interface {
	HTTPWrapTransport(http.RoundTripper) http.RoundTripper
	HTTPRequestTrace(context.Context) *httptrace.ClientTrace
	span.SpanOp
}

// NewDefaultHTTPClient creates a new HTTPClient with default options
func NewDefaultHTTPClient(log log.Log, t Tracer) *HTTPClient {
	return NewHTTPClient(log, t, 4, time.Second*1, time.Second*5)
}

// NewHTTPClient creates a new HTTPClient
func NewHTTPClient(log log.Log, tr Tracer, retryMax int, retryWaitMin, retryWaitMax time.Duration) *HTTPClient {
	var rt http.RoundTripper
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	rt = t
	if tr != nil {
		rt = tr.HTTPWrapTransport(t)
	}
	return New(log, tr, &http.Client{Transport: rt}, retryMax, retryWaitMin, retryWaitMax)
}

// NewH2CClient creates a new HTTPClient with configuration to support h2c servers
func NewH2CClient(log log.Log, tr Tracer, retryMax int, retryWaitMin, retryWaitMax time.Duration) *HTTPClient {
	var rt http.RoundTripper
	t := &http2.Transport{
		// So http2.Transport doesn't complain the URL scheme isn't 'https'
		AllowHTTP: false,
		// Pretend we are dialing a TLS endpoint.
		// Note, we ignore the passed tls.Config
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}
	rt = t
	if tr != nil {
		rt = tr.HTTPWrapTransport(t)
	}
	return New(log, tr, &http.Client{Transport: rt}, retryMax, retryWaitMin, retryWaitMax)
}

// New creates a new HTTPClient wrapper for passed *http.Client
func New(log log.Log, tr Tracer, c *http.Client, retryMax int, retryWaitMin, retryWaitMax time.Duration) *HTTPClient {
	return &HTTPClient{Client: c, log: log.NewResourceLogger("HttpClient"), retryMax: retryMax, minRetryWait: retryWaitMin, maxRetryWait: retryWaitMax, checkRetry: retryablehttp.DefaultRetryPolicy, backoff: retryablehttp.DefaultBackoff, tr: tr}
}

func (h *HTTPClient) validateResponseBody(resBody interface{}) error {
	if resBody != nil {
		v := reflect.ValueOf(resBody)
		if v.Kind() != reflect.Ptr {
			return fmt.Errorf("HttpClient.Validator: `resBody` is not ptr type is %T", resBody)
		}
	}
	return nil
}

// Encode marshals the passed object and creates the io.Reader object
func (h *HTTPClient) Encode(ctx context.Context, data interface{}) (io.Reader, error) {
	var body io.Reader
	if data != nil {
		switch v := data.(type) {
		case string:
			body = bytes.NewReader([]byte(v))
		case []byte:
			body = bytes.NewReader(v)
		case io.ReadCloser:
			body = v
		case io.Reader:
			body = v
		default:
			buff := &bytes.Buffer{}
			err := json.NewEncoder(buff).Encode(&data)
			if err != nil {
				return nil, fmt.Errorf("HttpClient.Encode: error encoding payload: %w", err)
			}
			body = buff
		}
	}
	return body, nil
}

// Decode un-marshals the passed bytes to an object and returns a copy in io.ReadCloser
func (h *HTTPClient) Decode(ctx context.Context, body []byte, data interface{}) (io.ReadCloser, error) {
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
				return newBuff, fmt.Errorf("HttpClient.Decode: %w: %w", ErrResponseUnmarshal, err)
			}
		}
		return nil, nil
	}
	newBuff := io.NopCloser(bytes.NewReader(body))
	return newBuff, nil
}

// Get HTTP GET method invocation
func (h *HTTPClient) Get(ctx context.Context, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	return h.Call(ctx, http.MethodGet, url, reqBody, resBody, headers)
}

// Post HTTP POST method invocation
func (h *HTTPClient) Post(ctx context.Context, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	return h.Call(ctx, http.MethodPost, url, reqBody, resBody, headers)
}

// Put HTTP PUT method invocation
func (h *HTTPClient) Put(ctx context.Context, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	return h.Call(ctx, http.MethodPut, url, reqBody, resBody, headers)
}

// Patch HTTP PATCH method invocation
func (h *HTTPClient) Patch(ctx context.Context, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	return h.Call(ctx, http.MethodPatch, url, reqBody, resBody, headers)
}

// Delete HTTP DELETE method invocation
func (h *HTTPClient) Delete(ctx context.Context, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	return h.Call(ctx, http.MethodDelete, url, reqBody, resBody, headers)
}

// Call pre-processes the request and response
func (h *HTTPClient) Call(ctx context.Context, method, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	err := h.validateResponseBody(resBody)
	if err != nil {
		return nil, err
	}
	var req *http.Request
	body, err := h.Encode(ctx, reqBody)
	if err != nil {
		return nil, err
	}
	reqCtx := ctx
	if h.tr != nil {
		reqCtx = httptrace.WithClientTrace(ctx, h.tr.HTTPRequestTrace(ctx))
	}
	req, err = http.NewRequestWithContext(reqCtx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("HttpClient.Call: error creating request: %w", err)
	}
	log.SetCorrelationHeader(ctx, req)
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	h.log.Info(ctx, "Request", map[string]interface{}{
		"method":  method,
		"url":     url,
		"headers": req.Header,
	})
	r, err := h.Do(req)
	if err != nil {
		return r, fmt.Errorf("HttpClient.Call: network call error: %w", err)
	} else if r == nil {
		return r, err
	}
	logRes := map[string]interface{}{
		"statusCode": r.StatusCode,
		"headers":    r.Header,
	}
	if r.StatusCode > 299 {
		h.log.Error(ctx, "Response", logRes)
		return r, ErrResponseFromUpstream
	} else {
		h.log.Info(ctx, "Response", logRes)
	}
	resBlob, _ := io.ReadAll(r.Body)
	r.Body, err = h.Decode(ctx, resBlob, resBody)
	return r, err
}

// Do adds additional retry and backoff mechanism
func (h *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	/*this is a modified version of go-retryablehttp*/
	var resp *http.Response
	var attempt int
	var shouldRetry bool
	var doErr, respErr error
	var reqBody []byte
	if req.ContentLength > 0 {
		reqBody, _ = io.ReadAll(req.Body)
	}
	for i := 0; ; i++ {
		doErr = nil
		attempt++
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		resp, doErr = h.Client.Do(req)
		shouldRetry, respErr = h.backOffAndRetry(i, req, resp, doErr)
		if !shouldRetry {
			break
		}
	}

	// this is the closest we have to success criteria
	if doErr == nil && respErr == nil && !shouldRetry {
		return resp, nil
	}

	var err error
	if respErr != nil {
		err = respErr
	} else {
		err = doErr
	}

	return resp, err

}

func (h *HTTPClient) backOffAndRetry(i int, req *http.Request, resp *http.Response, doErr error) (bool, error) {
	shouldRetry, respErr := h.checkRetry(req.Context(), resp, doErr)
	if !shouldRetry || respErr != nil {
		return shouldRetry, respErr
	}
	remain := h.retryMax - i
	if remain <= 0 {
		return false, respErr
	}
	wait := h.backoff(h.minRetryWait, h.maxRetryWait, i, resp)
	if resp != nil && resp.ContentLength > 0 {
		defer resp.Body.Close()
		resBlob, _ := io.ReadAll(resp.Body)
		h.log.Notice(req.Context(), fmt.Sprintf("request failed with status code %v retry %v of %v in %vms", resp.StatusCode, i+1, h.retryMax, wait.Milliseconds()), string(resBlob))
	} else if doErr != nil {
		h.log.Notice(req.Context(), fmt.Sprintf("request failed with error - retry %v of %v in %vms", i+1, h.retryMax, wait.Milliseconds()), doErr)
	} else {
		h.log.Notice(req.Context(), fmt.Sprintf("request failed - retry %v of %v in %vms", i+1, h.retryMax, wait.Milliseconds()), nil)
	}
	if h.tr != nil {
		_, span := h.tr.NewSpanFromContext(req.Context(), "http.Backoff", span.SpanKindInternal, "")
		if span != nil {
			span.SetAttribute("http.retryCount", i+1)
			span.SetAttribute("http.maxRetryCount", h.retryMax)
			span.SetAttribute("http.retryBackoffDurationMS", wait.Milliseconds())
			defer span.Finish()
		}
	}
	timer := time.NewTimer(wait)
	select {
	case <-req.Context().Done():
		timer.Stop()
		h.Client.CloseIdleConnections()
		return false, req.Context().Err()
	case <-timer.C:
	}
	return true, nil
}
