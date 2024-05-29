package main

import (
	server "github.com/sabariramc/goserverbase/v6/app/server/httpserver/test"
)

func main() {
	s := server.New(nil)
	s.StartH2CServer()
}
