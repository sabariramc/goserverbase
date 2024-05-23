package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/segmentio/kafka-go"
)

// kafkaLogger is a custom logger for Kafka operations.
type kafkaLogger struct {
	log.Log
	ctx     context.Context
	isError bool
}

var debugLogPrefix = []string{"no messages received from kafka within the allocated time for partition", "writing %d messages to"}

// Printf logs a formatted message. Depending on the severity, it either logs an error or a debug/notice message.
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

// kafkaDeliveryReportLogger is a custom logger for Kafka delivery reports.
type kafkaDeliveryReportLogger struct {
	log.Log
	ctx context.Context
}

// DeliveryReport logs the delivery report of Kafka messages. If there's an error, it logs the error and the affected messages.
func (k *kafkaDeliveryReportLogger) DeliveryReport(messages []kafka.Message, err error) {
	if err != nil {
		k.Error(k.ctx, "Error Writing to topic", err)
		k.Error(k.ctx, "Affected Message", messages)
		return
	}
	k.Debug(k.ctx, "Delivery report", messages)
}
