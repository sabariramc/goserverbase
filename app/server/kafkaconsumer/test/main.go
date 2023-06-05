package main

import (
	"github.com/sabariramc/goserverbase/v3/app/server/kafkaconsumer/test/server"
)

func main() {
	s := server.NewServer()
	s.StartConsumer()
}
