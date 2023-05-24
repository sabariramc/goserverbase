package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	cKafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sabariramc/goserverbase/v2/errors"
	"github.com/sabariramc/goserverbase/v2/errors/message"
	"github.com/sabariramc/goserverbase/v2/log"
	"github.com/sabariramc/goserverbase/v2/utils"
)

type Producer interface {
	Produce(ctx context.Context, key string, message *utils.Message, headers map[string]string) (*cKafka.Message, error)
}

type ErrorNotifierKafka struct {
	producer    Producer
	log         *log.Logger
	serviceName string
}

func New(ctx context.Context, log *log.Logger, baseURL, topicName, serviceName string, producer Producer) *ErrorNotifierKafka {
	return &ErrorNotifierKafka{producer: producer, log: log, serviceName: serviceName}
}

func (e *ErrorNotifierKafka) Send5XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error {
	return e.send(ctx, errorCode, err, stackTrace, errorData, errors.ERROR_5xx)
}
func (e *ErrorNotifierKafka) Send4XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error {
	return e.send(ctx, errorCode, err, stackTrace, errorData, errors.ERROR_4xx)
}

func (e *ErrorNotifierKafka) send(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}, alertType string) error {
	msg := message.CreateErrorMessage(ctx, e.serviceName, errorCode, err, stackTrace, errorData, alertType)
	_, err = e.producer.Produce(ctx, "", msg, map[string]string{"x-error-timestamp": strconv.FormatInt(time.Now().UnixMilli(), 10)})
	if err != nil {
		e.log.Error(ctx, "Error in error-notifier", err)
		err = fmt.Errorf("ErrorNotifierKafka.send : %w", err)
	}
	return err
}
