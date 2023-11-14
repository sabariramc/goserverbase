package kafka

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/segmentio/kafka-go"
)

type kafkaLogger struct {
	*log.Logger
	ctx     context.Context
	isError bool
}

func (k *kafkaLogger) Printf(shortMessage string, logMessage ...interface{}) {
	message := fmt.Sprintf(shortMessage, logMessage...)
	if k.isError {
		k.Error(k.ctx, message, nil)
	} else {
		k.Debug(k.ctx, message, nil)
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
