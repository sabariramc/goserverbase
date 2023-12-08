package main

import (
	"github.com/sabariramc/goserverbase/v4/app/server/httpserver/test/server"
)

func main() {
	s := server.NewServer()
	s.StartServer()
}
