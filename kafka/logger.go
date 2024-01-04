package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/segmentio/kafka-go"
)

type kafkaLogger struct {
	*log.Logger
	ctx     context.Context
	isError bool
}

var debugLogPrefix = []string{"no messages received from kafka within the allocated time for partition"}

func (k *kafkaLogger) Printf(shortMessage string, logMessage ...interface{}) {
	message := fmt.Sprintf(shortMessage, logMessage...)
	if k.isError {
		k.Error(k.ctx, message, nil)
	} else {
		for _, v := range debugLogPrefix {
			if strings.HasPrefix(shortMessage, v) {
				k.Debug(k.ctx, message, nil)
				return
			}
		}
		k.Notice(k.ctx, message, nil)
	}
}

type kafkaDeliveryReportLogger struct {
	*log.Logger
	ctx context.Context
}

func (k *kafkaDeliveryReportLogger) DeliveryReport(messages []kafka.Message, err error) {
	if err != nil {
		k.Error(k.ctx, "Error Writing to topic", err)
		k.Error(k.ctx, "Affected Message", messages)
		return
	}
	k.Debug(k.ctx, "Delivery report", messages)
}
