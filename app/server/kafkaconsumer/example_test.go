package kafkaconsumer_test

import (
	"context"
	"testing"

	"github.com/sabariramc/goserverbase/v6/app/server/kafkaconsumer"
	"github.com/sabariramc/goserverbase/v6/errors"
	"github.com/sabariramc/goserverbase/v6/kafka"
)

func Example() {
	srv := kafkaconsumer.New()
	srv.AddHandler(context.Background(), "gobase.test.topic1", func(ctx context.Context, m *kafka.Message) error {
		return nil
	})
	srv.AddHandler(context.Background(), "gobase.test.topic2", func(ctx context.Context, m *kafka.Message) error {
		return &errors.CustomError{ErrorCode: "gobase.test.error", ErrorMessage: "error sample"}
	})
	srv.StartConsumer(context.Background())
}

func TestDefaultKafkaServer(t *testing.T) {
	srv := kafkaconsumer.New()
	srv.AddHandler(context.Background(), "gobase.test.topic1", func(ctx context.Context, m *kafka.Message) error {
		return nil
	})
	srv.AddHandler(context.Background(), "gobase.test.topic2", func(ctx context.Context, m *kafka.Message) error {
		return &errors.CustomError{ErrorCode: "gobase.test.error", ErrorMessage: "error sample"}
	})
	srv.StartConsumer(context.Background())
}
