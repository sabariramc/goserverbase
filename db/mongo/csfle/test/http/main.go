package main

import (
	server "github.com/sabariramc/goserverbase/v6/db/mongo/csfle/test"
)

func main() {
	s := server.NewServer(nil)
	s.StartServer()
}
