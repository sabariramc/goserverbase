package httputil_test

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"golang.org/x/net/http2"
	"gotest.tools/assert"
)

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	c := &Client{
		client: &http.Client{
			Transport: &http2.Transport{
				// So http2.Transport doesn't complain the URL scheme isn't 'https'
				AllowHTTP: false,
				// Pretend we are dialing a TLS endpoint.
				// Note, we ignore the passed tls.Config
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		},
	}
	return c
}

func (c *Client) Post(host, path string, data []byte) ([]byte, error) {
	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "https",
			Host:   host,
			Path:   path,
		},
		Header: http.Header{},
		Body:   io.NopCloser(bytes.NewReader(data)),
	}

	// Sends the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 500 {
		return nil, err
	}

	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func TestH2CServer(t *testing.T) {
	client := NewClient()
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				data, err := client.Post("localhost:8080", "/service/v1/test/req", []byte("fasdfasfasdfsadf"))
				assert.NilError(t, err)
				fmt.Println(string(data))
			}
		}()
	}
	wg.Wait()
}
