package main

import (
	"log"

	server "github.com/sabariramc/goserverbase/v5/app/server/httpserver/test"
	"github.com/sabariramc/goserverbase/v5/instrumentation/contrib/otel"
)

func main() {
	tr, err := otel.Init()
	if err != nil {
		log.Fatal("tracer failed", err)
	}
	defer otel.ShutDown()
	s := server.NewServer(tr)
	s.StartServer()
}
