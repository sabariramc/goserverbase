package httpserver_test

import (
	"github.com/sabariramc/goserverbase/v6/app/server/httpserver"
)

func Example() {
	sev := httpserver.New()
	sev.StartServer()
}
