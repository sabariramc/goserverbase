package main

import (
	"context"
	"time"

	"github.com/sabariramc/goserverbase/v4/app/server/kafkaconsumer/test/server"
)

func main() {
	s := server.NewServer()
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	s.StartConsumer(ctx)
}
