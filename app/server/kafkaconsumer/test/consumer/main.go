package main

import (
	"context"

	server "github.com/sabariramc/goserverbase/v4/app/server/kafkaconsumer/test"
)

func main() {
	s := server.NewServer()
	s.StartConsumer(context.Background())
}
