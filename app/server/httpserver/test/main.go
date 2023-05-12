package main

import (
	"github.com/sabariramc/goserverbase/v2/baseapp/server/httpserver/test/server"
)

func main() {
	s := server.NewServer()
	s.StartHttpServer()
}
