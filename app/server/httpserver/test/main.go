package main

import (
	"github.com/sabariramc/goserverbase/v3/app/server/httpserver/test/server"
)

func main() {
	s := server.NewServer()
	s.StartServer()
}
