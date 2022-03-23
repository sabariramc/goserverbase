package tests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/assert"
	"sabariram.com/goserverbase/baseapp"
	"sabariram.com/goserverbase/utils"
)

func HTTPTestFunction(r *http.Request) (statusCode int, response interface{}, header http.Header, err error) {
	err = fmt.Errorf("Error")
	err = fmt.Errorf("Level 1 : %w", err)
	err = fmt.Errorf("Level 2 : %w", err)
	panic(err)
}

func TestJsonResponder(t *testing.T) {
	srv := &baseapp.BaseApp{}
	srv.SetConfig(baseapp.ServerConfig{
		LoggerConfig: ServerTestConfig.Logger,
		AppConfig:    ServerTestConfig.App,
	})
	srv.SetLogger(ServerTestLogger)
	ip := make(map[string]string)
	req := httptest.NewRequest(http.MethodGet, "/tenant", nil)
	req.Header.Set("x-api-key", utils.GetEnv("TEST_API_KEY", ""))
	w := httptest.NewRecorder()
	srv.JSONResponderWithHeader(ip, HTTPTestFunction)(w, req)
	blob, _ := ioutil.ReadAll(w.Body)
	fmt.Println(string(blob))
	assert.Equal(t, w.Result().StatusCode, http.StatusInternalServerError)
}
