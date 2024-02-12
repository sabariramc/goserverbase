package server_test

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/google/uuid"
)

func callURL(url string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	body, _ := json.Marshal(map[string]string{"fasdfsda": "fasdfas", "fasdfas": "fasdfas"})
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-correlation-id", "sabariram-load-"+uuid.NewString())
		http.DefaultClient.Do(req)
	}
}

func TestRoutes(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			callURL("https://localhost:60006/service/v1/test/all")
		}()
	}
	wg.Wait()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			callURL("http://localhost:60005/service/v1/test/all")
		}()
	}
	wg.Wait()
}
