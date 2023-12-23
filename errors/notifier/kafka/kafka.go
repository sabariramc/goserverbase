package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v4/errors"
	"github.com/sabariramc/goserverbase/v4/errors/message"
	"github.com/sabariramc/goserverbase/v4/kafka"
	"github.com/sabariramc/goserverbase/v4/log"
)

type ErrorNotifierKafka struct {
	producer    *kafka.Producer
	log         *log.Logger
	serviceName string
	topic       string
}

func New(ctx context.Context, log *log.Logger, serviceName string, topic string, producer *kafka.Producer) *ErrorNotifierKafka {
	return &ErrorNotifierKafka{producer: producer, log: log.NewResourceLogger("ErrorNotifierKafka"), serviceName: serviceName, topic: topic}
}

func (e *ErrorNotifierKafka) Send5XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error {
	return e.send(ctx, errorCode, err, stackTrace, errorData, errors.ErrorCode5XX)
}
func (e *ErrorNotifierKafka) Send4XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error {
	return e.send(ctx, errorCode, err, stackTrace, errorData, errors.ErrorCode4XX)
}

func (e *ErrorNotifierKafka) send(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}, alertType string) error {
	msg := message.CreateErrorMessage(ctx, e.serviceName, errorCode, err, stackTrace, errorData, alertType)
	err = e.producer.ProduceMessageWithTopic(ctx, e.topic, uuid.NewString(), msg, map[string]string{"x-error-timestamp": strconv.FormatInt(time.Now().UnixMilli(), 10)})
	if err != nil {
		e.log.Error(ctx, "Error in error-notifier", err)
		err = fmt.Errorf("ErrorNotifierKafka.send: %w", err)
	}
	return err
}
