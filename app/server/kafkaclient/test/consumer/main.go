package main

import (
	server "github.com/sabariramc/goserverbase/v6/app/server/kafkaclient/test"
)

func main() {
	s := server.New(nil)
	s.StartClient()
}
