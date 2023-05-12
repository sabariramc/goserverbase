package main

import (
	"github.com/sabariramc/goserverbase/v2/app/server/kafkaclient/test/server"
)

func main() {
	s := server.NewServer()
	s.StartKafkaConsumer()
}
