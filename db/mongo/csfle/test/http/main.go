package main

import (
	server "github.com/sabariramc/goserverbase/v5/db/mongo/csfle/test"
)

func main() {
	s := server.NewServer(nil)
	s.StartServer()
}
