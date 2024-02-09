package main

import (
	server "github.com/sabariramc/goserverbase/v5/app/server/httpserver/test"
)

func main() {
	s := server.NewServer(nil)
	s.StartTLSServer()
}
