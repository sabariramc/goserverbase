package kafkaclient_test

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/app/server/kafkaclient"
	"github.com/sabariramc/goserverbase/v6/errors"
	"github.com/sabariramc/goserverbase/v6/kafka"
)

func Example() {
	srv := kafkaclient.New()
	srv.AddHandler(context.Background(), "gobase.test.topic1", func(ctx context.Context, m *kafka.Message) error {
		return nil
	})
	srv.AddHandler(context.Background(), "gobase.test.topic2", func(ctx context.Context, m *kafka.Message) error {
		return &errors.CustomError{ErrorCode: "gobase.test.error", ErrorMessage: "error sample"}
	})
	srv.StartClient()
}
