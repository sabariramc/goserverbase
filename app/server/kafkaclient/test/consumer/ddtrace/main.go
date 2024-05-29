package main

import (
	"log"

	server "github.com/sabariramc/goserverbase/v6/app/server/kafkaclient/test"
	"github.com/sabariramc/goserverbase/v6/instrumentation/contrib/ddtrace"
)

func main() {
	tr, err := ddtrace.Init()
	if err != nil {
		log.Fatal("tracer failed", err)
	}
	defer ddtrace.ShutDown()
	s := server.New(tr)
	s.StartClient()
}
