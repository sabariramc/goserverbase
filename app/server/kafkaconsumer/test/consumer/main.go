package main

import (
	"context"

	server "github.com/sabariramc/goserverbase/v5/app/server/kafkaconsumer/test"
)

func main() {
	s := server.NewServer(nil)
	s.StartConsumer(context.Background())
}
