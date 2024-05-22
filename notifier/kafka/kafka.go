package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/errors/message"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/notifier"
)

// NotifierKafka implements the Notifier interface with Kafka as the messaging system.
type NotifierKafka struct {
	producer    *kafka.Producer
	log         log.Log
	serviceName string
	topic       string
}

// New creates a new instance of NotifierKafka.
func New(options ...Option) *NotifierKafka {
	config := GetDefaultConfig()
	for _, fn := range options {
		fn(&config)
	}
	return &NotifierKafka{
		producer:    config.Producer,
		log:         config.Log,
		serviceName: config.ServiceName,
		topic:       config.Topic,
	}
}

// Notify5XX sends a notification for 5XX errors.
func (nk *NotifierKafka) Notify5XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error {
	return nk.notify(ctx, errorCode, err, stackTrace, errorData, notifier.NotificationCode5XX)
}

// Notify4XX sends a notification for 4XX errors.
func (nk *NotifierKafka) Notify4XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error {
	return nk.notify(ctx, errorCode, err, stackTrace, errorData, notifier.NotificationCode4XX)
}

func (nk *NotifierKafka) notify(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}, alertType string) error {
	msg := message.CreateErrorMessage(ctx, nk.serviceName, errorCode, err, stackTrace, errorData, alertType)
	if nk.topic == "" || nk.producer == nil {
		return nil
	}
	err = nk.producer.ProduceMessageWithTopic(ctx, nk.topic, uuid.NewString(), msg, map[string]string{"x-error-timestamp": strconv.FormatInt(time.Now().UnixMilli(), 10)})
	if err != nil {
		nk.log.Error(ctx, "Error in error-notifier", err)
		err = fmt.Errorf("ErrorNotifierKafka.send: %w", err)
	}
	return err
}
