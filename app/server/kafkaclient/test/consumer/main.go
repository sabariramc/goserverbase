package main

import (
	"context"

	server "github.com/sabariramc/goserverbase/v6/app/server/kafkaclient/test"
)

func main() {
	s := server.NewServer(nil)
	s.StartConsumer(context.Background())
}
