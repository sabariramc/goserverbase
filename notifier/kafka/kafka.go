package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v5/errors/message"
	"github.com/sabariramc/goserverbase/v5/kafka"
	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/notifier"
)

type ErrorNotifierKafka struct {
	producer    *kafka.Producer
	log         log.Log
	serviceName string
	topic       string
}

func New(ctx context.Context, log log.Log, serviceName string, topic string, producer *kafka.Producer) *ErrorNotifierKafka {
	return &ErrorNotifierKafka{producer: producer, log: log.NewResourceLogger("ErrorNotifierKafka"), serviceName: serviceName, topic: topic}
}

func (e *ErrorNotifierKafka) Send5XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error {
	return e.send(ctx, errorCode, err, stackTrace, errorData, notifier.NotificationCode5XX)
}
func (e *ErrorNotifierKafka) Send4XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error {
	return e.send(ctx, errorCode, err, stackTrace, errorData, notifier.NotificationCode4XX)
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
