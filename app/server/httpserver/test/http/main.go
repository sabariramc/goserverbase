package main

import (
	server "github.com/sabariramc/goserverbase/v4/app/server/httpserver/test"
)

func main() {
	s := server.NewServer()
	s.StartServer()
}
