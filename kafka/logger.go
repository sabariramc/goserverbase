package kafka

import (
	"context"

	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/segmentio/kafka-go"
)

type kafkaLogger struct {
	*log.Logger
	ctx     context.Context
	isError bool
}

func (k *kafkaLogger) Printf(shortMessage string, logMessage ...interface{}) {
	if k.isError {
		k.Error(k.ctx, shortMessage, logMessage)
	} else {
		k.Debug(k.ctx, shortMessage, logMessage)
	}
}

func (k *kafkaLogger) DeliveryReport(messages []kafka.Message, err error) {
	if err != nil {
		k.Error(k.ctx, "Error Writing to topic", err)
		k.Error(k.ctx, "Affected Message", messages)
		return
	}
	k.Debug(k.ctx, "Delivery report", messages)
}
