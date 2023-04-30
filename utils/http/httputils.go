package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sabariramc/goserverbase/v2/log"
)

type HttpClient struct {
	*retryablehttp.Client
	log *log.Logger
}

func NewHttpClient(timeout time.Duration, log *log.Logger) *HttpClient {
	t := cleanhttp.DefaultPooledTransport()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	c := &HttpClient{Client: retryablehttp.NewClient(), log: log}
	c.HTTPClient.Transport = t
	return c
}

func (h *HttpClient) Do(ctx context.Context, method, url string, reqBody, resBody interface{}, headers map[string]string) (*http.Response, error) {
	var body io.ReadCloser
	var reqPrint interface{}
	req, err := retryablehttp.NewRequestWithContext(ctx, method, url, nil)
	if reqBody != nil {
		if v, ok := reqBody.([]byte); ok {
			reqPrint = string(v)
			body = io.NopCloser(bytes.NewReader(v))
		} else if v, ok := reqBody.(io.Reader); ok {
			body = io.NopCloser(v)
		} else if v, ok := reqBody.(io.ReadCloser); ok {
			body = v
		} else {
			reqPrint = reqBody
			buff := &bytes.Buffer{}
			err := json.NewEncoder(buff).Encode(&reqBody)
			body = io.NopCloser(buff)
			if err != nil {
				h.log.Error(ctx, "Http.Do.PayloadEncoding", err)
				return nil, fmt.Errorf("http.do.PayloadEncoding: %w", err)
			}
		}
	}
	if err != nil {
		h.log.Error(ctx, "Http.Do.RequestCreation", err)
		return nil, fmt.Errorf("http.do.requestCreation: %w", err)
	}
	log.SetCorrelationHeader(ctx, req.Request)
	for key, val := range headers {
		req.Header.Add(key, val)
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
	if r.StatusCode == http.StatusNoContent || r.ContentLength <= 0 {
		h.log.Debug(ctx, "Response", map[string]interface{}{
			"statusCode": r.StatusCode,
		})
		return r, nil
	}
	defer r.Body.Close()
	blobBody, _ := ioutil.ReadAll(r.Body)
	h.log.Debug(ctx, "Response", map[string]interface{}{
		"statusCode": r.StatusCode,
		"body":       string(blobBody),
	})
	err = json.Unmarshal(blobBody, resBody)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(blobBody))
		h.log.Error(ctx, "http.Do.responseBodyMarshall", err)
		return r, fmt.Errorf("http.do.responseBodyMarshall: %w", err)
	}
	return r, nil
}
