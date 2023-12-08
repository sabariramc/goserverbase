package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/sabariramc/goserverbase/v4/errors"
	"github.com/sabariramc/goserverbase/v4/errors/message"
	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/utils"
)

type Producer interface {
	ProduceMessage(ctx context.Context, key string, message *utils.Message, headers map[string]string) error
}

type ErrorNotifierKafka struct {
	producer    Producer
	log         *log.Logger
	serviceName string
}

func New(ctx context.Context, log *log.Logger, serviceName string, producer Producer) *ErrorNotifierKafka {
	return &ErrorNotifierKafka{producer: producer, log: log.NewResourceLogger("ErrorNotifierKafka"), serviceName: serviceName}
}

func (e *ErrorNotifierKafka) Send5XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error {
	return e.send(ctx, errorCode, err, stackTrace, errorData, errors.ERROR_5xx)
}
func (e *ErrorNotifierKafka) Send4XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error {
	return e.send(ctx, errorCode, err, stackTrace, errorData, errors.ERROR_4xx)
}
func (e *ErrorNotifierKafka) GetProcessor() any {
	return e.producer
}

func (e *ErrorNotifierKafka) send(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}, alertType string) error {
	msg := message.CreateErrorMessage(ctx, e.serviceName, errorCode, err, stackTrace, errorData, alertType)
	err = e.producer.ProduceMessage(ctx, "", msg, map[string]string{"x-error-timestamp": strconv.FormatInt(time.Now().UnixMilli(), 10)})
	if err != nil {
		e.log.Error(ctx, "Error in error-notifier", err)
		err = fmt.Errorf("ErrorNotifierKafka.send : %w", err)
	}
	return err
}
