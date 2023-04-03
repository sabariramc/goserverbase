package notifier

import (
	"context"
	"fmt"
	"time"

	"github.com/sabariramc/goserverbase/errors"
	"github.com/sabariramc/goserverbase/kafka"
	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/utils"
)

type ErrorNotifierKafka struct {
	producer    *kafka.HTTPProducer
	log         *log.Logger
	serviceName string
}

func New(ctx context.Context, log *log.Logger, baseURL, topicName, serviceName string) *ErrorNotifierKafka {
	return &ErrorNotifierKafka{producer: kafka.NewHTTPProducer(ctx, log, baseURL, topicName, time.Minute), log: log, serviceName: serviceName}
}

func (e ErrorNotifierKafka) Send5XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error {
	return e.send(ctx, errorCode, err, stackTrace, errorData, errors.ERROR_5xx)
}
func (e ErrorNotifierKafka) Send4XX(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}) error {
	return e.send(ctx, errorCode, err, stackTrace, errorData, errors.ERROR_4xx)
}

func (e ErrorNotifierKafka) send(ctx context.Context, errorCode string, err error, stackTrace string, errorData interface{}, alertType string) error {
	correlation_params := log.GetCorrelationParam(ctx)
	correlation := make(map[string]any, 0)
	utils.StrictJsonTransformer(correlation_params, &correlation)
	correlation["timestamp"] = time.Now().UnixMilli()
	correlation["identity"] = log.GetCustomerIdentifier(ctx)
	msg := utils.NewMessage("error", errorCode)
	msg.AddPayload("category", &utils.Payload{"entity": map[string]interface{}{"category": alertType}})
	msg.AddPayload("correlation", &utils.Payload{"entity": correlation})
	msg.AddPayload("source", &utils.Payload{"entity": map[string]interface{}{"source": e.serviceName}})
	msg.AddPayload("stackTrace", &utils.Payload{"entity": map[string]interface{}{"stackTrace": stackTrace, "error": err}})
	msg.AddPayload("version", &utils.Payload{"entity": map[string]interface{}{"version": "v1"}})
	msg.AddPayload("errorData", &utils.Payload{"entity": map[string]interface{}{"errorData": errorData}})
	_, err = e.producer.Produce(ctx, "", msg)
	if err != nil {
		e.log.Error(ctx, "Error in error-notifier", err)
		err = fmt.Errorf("ErrorNotifierKafka.send : %w", err)
	}
	return err
}
