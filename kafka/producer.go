package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v5/kafka/api"
	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/utils"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	*api.Writer
	config                  KafkaProducerConfig
	log                     log.Log
	topic                   string
	serviceName             string
	autoFlushCancel         context.CancelFunc
	isTopicSpecificProducer bool
	wg                      sync.WaitGroup
	isBatch                 bool
}

func NewProducer(ctx context.Context, logger log.Log, config *KafkaProducerConfig) (*Producer, error) {
	if config.Batch && config.Async {
		return nil, fmt.Errorf("NewProducer: `Batch` and `Async` are mutually exclusive")
	}
	if !config.Batch && !config.Async {
		return nil, fmt.Errorf("NewProducer: set either `Batch` or `Async`")
	}
	if config.Batch {
		if config.BatchMaxBuffer <= 0 {
			config.BatchMaxBuffer = 100
		}
		if config.BatchFlushIntervalInMs <= 0 {
			config.BatchFlushIntervalInMs = 1000
		}
	}
	logger = logger.NewResourceLogger("KafkaProducer")
	defaultCorrelationParam := &log.CorrelationParam{CorrelationId: config.ServiceName + ":KafkaProducer"}
	kLog := &kafkaDeliveryReportLogger{
		Log: logger.NewResourceLogger("KafkaProducerDeliveryLog"),
		ctx: log.GetContextWithCorrelation(context.Background(), defaultCorrelationParam),
	}
	if config.Acknowledge == 0 {
		logger.Warning(ctx, "Kafka replica acknowledgement is set to None", nil)
	}
	p := &kafka.Writer{
		Addr:     kafka.TCP(config.Brokers...),
		Topic:    config.Topic,
		Balancer: &kafka.Hash{},
		Transport: &kafka.Transport{
			SASL: config.SASLMechanism,
			TLS:  config.TLSConfig,
		},
		Logger: &kafkaLogger{
			Log:     logger.NewResourceLogger("KafkaProducerInfoLog"),
			ctx:     log.GetContextWithCorrelation(context.Background(), defaultCorrelationParam),
			isError: false,
		},
		ErrorLogger: &kafkaLogger{
			isError: true,
			Log:     logger.NewResourceLogger("KafkaProducerErrorLog"),
			ctx:     log.GetContextWithCorrelation(context.Background(), defaultCorrelationParam),
		},
		Completion:   kLog.DeliveryReport,
		RequiredAcks: kafka.RequiredAcks(config.Acknowledge),
		Async:        config.Async,
	}
	writer := api.NewWriter(ctx, p, config.BatchMaxBuffer, logger)
	isTopicSpecificProducer := false
	if config.Topic != "" {
		isTopicSpecificProducer = true
	}
	k := &Producer{
		serviceName:             config.ServiceName,
		log:                     logger,
		config:                  *config,
		Writer:                  writer,
		topic:                   config.Topic,
		isTopicSpecificProducer: isTopicSpecificProducer,
		isBatch:                 config.Batch,
	}
	if config.Batch {
		autoFlushContext, cancel := context.WithCancel(log.GetContextWithCorrelation(context.Background(), defaultCorrelationParam))
		k.autoFlushCancel = cancel
		k.wg.Add(1)
		go k.autoFlush(autoFlushContext)
		logger.Notice(ctx, "Kafak producer is set to batch mode", nil)
	}
	return k, nil
}

func (k *Producer) ProduceMessage(ctx context.Context, key string, message *utils.Message, headers map[string]string) (err error) {
	blob, err := json.Marshal(message)
	if err != nil {
		k.log.Error(ctx, "Failed to encode message", err)
		k.log.Error(ctx, "Message", message)
		return fmt.Errorf("Producer.ProduceMessage: error marshalling message: %w", err)
	}
	return k.ProduceToTopic(ctx, k.topic, key, blob, headers)
}

func (k *Producer) ProduceMessageWithTopic(ctx context.Context, topic, key string, message *utils.Message, headers map[string]string) (err error) {
	blob, err := json.Marshal(message)
	if err != nil {
		k.log.Error(ctx, "Failed to encode message", err)
		k.log.Error(ctx, "Message", message)
		return fmt.Errorf("Producer.ProduceMessageWithTopic: error marshalling message: %w", err)
	}
	return k.ProduceToTopic(ctx, topic, key, blob, headers)
}

func (k *Producer) ProduceToTopic(ctx context.Context, topic, key string, message []byte, headers map[string]string) (err error) {
	if k.isTopicSpecificProducer && topic != k.topic {
		err := fmt.Errorf("Producer.ProduceToTopic: topic is set for producer use `Produce` method")
		k.log.Error(ctx, "topic is set for producer use `Produce` method", err)
		return err
	}
	if headers == nil {
		headers = make(map[string]string, 0)
	}
	corr := log.GetCorrelationHeader(ctx)
	messageHeader := make([]kafka.Header, 0)
	for i, v := range corr {
		headers[i] = v
	}
	for i, v := range headers {
		messageHeader = append(messageHeader, kafka.Header{
			Key:   i,
			Value: []byte(v),
		})
	}
	k.log.Info(ctx, "MessageMeta", map[string]any{"key": key, "headers": headers, "topic": topic})
	msg := &kafka.Message{
		Key:     []byte(key),
		Value:   message,
		Headers: messageHeader,
		Time:    time.Now(),
	}
	if !k.isTopicSpecificProducer {
		msg.Topic = topic
	}
	err = k.Send(ctx, msg)
	if err == api.ErrWriterBufferFull {
		err = k.Flush(ctx)
		if err != nil {
			return err
		}
		return k.Send(ctx, msg)
	}
	return nil
}

func (k *Producer) autoFlush(ctx context.Context) {
	defer k.wg.Done()
	nCtx := context.WithoutCancel(ctx)
	timeout, _ := context.WithTimeout(context.Background(), time.Duration(k.config.BatchFlushIntervalInMs*uint64(time.Millisecond)))
	defer k.log.Notice(ctx, "auto flush stopped", nil)
	for {
		select {
		case <-timeout.Done():
			err := k.Flush(ctx)
			if err != nil {
				k.log.Emergency(ctx, "Error while writing kafka message", fmt.Errorf("Producer.autoFlush: %w", err), nil)
			}
			timeout, _ = context.WithTimeout(context.Background(), time.Duration(k.config.BatchFlushIntervalInMs*uint64(time.Millisecond)))
		case <-ctx.Done():
			err := k.Flush(nCtx)
			if err != nil {
				k.log.Error(nCtx, "error in auto flush", err)
			}
			return
		}
	}
}

func (k *Producer) Close(ctx context.Context) error {
	k.log.Notice(ctx, "Producer closer initiated for topic", k.topic)
	if k.isBatch {
		k.autoFlushCancel()
	}
	k.wg.Wait()
	err := k.Writer.Close(ctx)
	if err == nil {
		k.log.Notice(ctx, "Producer closed for topic", k.topic)
	}
	return err
}
