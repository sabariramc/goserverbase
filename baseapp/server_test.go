package baseapp_test

import (
	"log"
	"net/http"

	"github.com/sabariramc/goserverbase/baseapp"
)

func main() {
	s := baseapp.New(*ServerTestConfig.App, *ServerTestConfig.Logger, ServerTestLMux, nil, nil)

	log.Fatal(http.ListenAndServe(s.GetPort(), s))
}
