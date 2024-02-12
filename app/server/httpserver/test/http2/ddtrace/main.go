package main

import (
	"log"

	server "github.com/sabariramc/goserverbase/v5/app/server/httpserver/test"
	"github.com/sabariramc/goserverbase/v5/instrumentation/ddtrace"
)

func main() {
	tr, err := ddtrace.Init()
	if err != nil {
		log.Fatal("tracer failed", err)
	}
	defer ddtrace.ShutDown()
	s := server.NewServer(tr)
	s.StartTLSServer()
}
