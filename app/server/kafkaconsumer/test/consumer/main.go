package main

import (
	"context"
	"time"

	server "github.com/sabariramc/goserverbase/v4/app/server/kafkaconsumer/test"
)

func main() {
	s := server.NewServer()
	ctx, _ := context.WithTimeout(context.Background(), time.Minute*30)
	s.StartConsumer(ctx)
}
