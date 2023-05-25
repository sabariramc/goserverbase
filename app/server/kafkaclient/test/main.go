package main

import (
	"github.com/sabariramc/goserverbase/v3/app/server/kafkaclient/test/server"
)

func main() {
	s := server.NewServer()
	s.StartConsumer()
}
