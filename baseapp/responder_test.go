package baseapp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sabariramc/goserverbase/baseapp"
	"github.com/sabariramc/goserverbase/utils"
	"gotest.tools/assert"
)

func HTTPTestFunction(r *http.Request) (statusCode int, response interface{}, err error) {
	err = fmt.Errorf("Error")
	err = fmt.Errorf("Level 1 : %w", err)
	err = fmt.Errorf("Level 2 : %w", err)
	return http.StatusInternalServerError, nil, err
}

func TestJsonResponder(t *testing.T) {
	srv := baseapp.NewBaseApp(baseapp.ServerConfig{
		LoggerConfig: ServerTestConfig.Logger,
		ServerConfig: ServerTestConfig.App,
	}, ServerTestLMux, nil, nil)
	ip := make(map[string]string)
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(map[string]string{"FASDFs": "fasdf"})
	req := httptest.NewRequest(http.MethodPost, "/tenant", &buf)
	req.Header.Set("x-api-key", utils.GetEnv("TEST_API_KEY", ""))
	w := httptest.NewRecorder()
	srv.JSONResponder(&ip, HTTPTestFunction)(w, req)
	blob, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, w.Result().StatusCode, http.StatusInternalServerError)
	assert.Equal(t, string(blob), "{\"errorData\":\"Level 2 : Level 1 : Error\",\"errorMessage\":\"Unknown error\",\"errorCode\":\"UNKNOWN\"}")
}

func HTTPTestFunctionPanic(r *http.Request) (statusCode int, response interface{}, err error) {
	panic("fadafsfs")
}

func TestJsonResponderPanic(t *testing.T) {
	srv := baseapp.NewBaseApp(baseapp.ServerConfig{
		LoggerConfig: ServerTestConfig.Logger,
		ServerConfig: ServerTestConfig.App,
	}, ServerTestLMux, nil, nil)
	ip := make(map[string]string)
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(map[string]string{"FASDFs": "fasdf"})
	req := httptest.NewRequest(http.MethodPost, "/tenant", &buf)
	req.Header.Set("x-api-key", utils.GetEnv("TEST_API_KEY", ""))
	w := httptest.NewRecorder()
	srv.JSONResponder(&ip, HTTPTestFunctionPanic)(w, req)
	blob, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, w.Result().StatusCode, http.StatusInternalServerError)
	assert.Equal(t, string(blob), "{\"error\":\"Internal error occcured, if persist contact technical team\"}")
}
