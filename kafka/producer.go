package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/utils"
	"github.com/segmentio/kafka-go"
)

// Producer is a high-level API that extends Writer with time and count based auto flush
// and implements a Shutdown hook.
type Producer struct {
	*Writer
	config          ProducerConfig
	log             log.Log
	topic           string
	autoFlushCancel context.CancelFunc
	isTopicSpecific bool
	wg              sync.WaitGroup
	isBatch         bool
}

// NewProducer creates a new Producer instance with the provided configuration options.
/*
If writer is not set in [ProducerConfig] then creates a new [kafka.Writer] with the options passed to the function, and adds addition params

		kafka.Writer{
			Addr:     kafka.TCP(config.Brokers...),
			Topic:    config.Topic,
			Balancer: &kafka.Hash{},
			Transport: &kafka.Transport{
				SASL: config.SASLMechanism,
				TLS:  config.TLSConfig,
			},
			Completion:   kLog.DeliveryReport,
			RequiredAcks: kafka.RequiredAcks(config.RequiredAcks),
			Async:        config.Async,
			Logger: &kafkaLogger{
				Log:     logger.NewResourceLogger(config.ModuleName + ":InfoLog"),
				ctx:     ctx,
				isError: false,
			},
			ErrorLogger: &kafkaLogger{
				isError: true,
				Log:     logger.NewResourceLogger(config.ModuleName + ":ErrorLog"),
				ctx:     ctx,
			},
		}
*/
func NewProducer(options ...ProducerOption) (*Producer, error) {
	config := GetDefaultProducerConfig()
	// Apply options
	for _, opt := range options {
		opt(config)
	}
	err := ValidateProducerConfig(config)
	if err != nil {
		return nil, err
	}
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), &correlation.CorrelationParam{CorrelationID: config.ModuleName})
	logger := config.Log
	kLog := &kafkaDeliveryReportLogger{
		Log: logger.NewResourceLogger(config.ModuleName + ":DeliveryLog"),
		ctx: ctx,
	}
	if config.RequiredAcks == 0 {
		logger.Warning(ctx, "Kafka replica acknowledgement is set to None", nil)
	}
	if config.Writer == nil {
		config.Writer = &kafka.Writer{
			Addr:     kafka.TCP(config.Brokers...),
			Topic:    config.Topic,
			Balancer: &kafka.Hash{},
			Transport: &kafka.Transport{
				SASL: config.SASLMechanism,
				TLS:  config.TLSConfig,
			},
			Completion:   kLog.DeliveryReport,
			RequiredAcks: kafka.RequiredAcks(config.RequiredAcks),
			Async:        config.Async,
			Logger: &kafkaLogger{
				Log:     logger.NewResourceLogger(config.ModuleName + ":InfoLog"),
				ctx:     ctx,
				isError: false,
			},
			ErrorLogger: &kafkaLogger{
				isError: true,
				Log:     logger.NewResourceLogger(config.ModuleName + ":ErrorLog"),
				ctx:     ctx,
			},
		}
	}
	writer := NewWriter(ctx, config.Writer, config.MaxBuffer, logger, config.Trace)
	isTopicSpecificProducer := false
	if config.Topic != "" {
		isTopicSpecificProducer = true
	}
	k := &Producer{
		log:             logger,
		config:          *config,
		Writer:          writer,
		topic:           config.Topic,
		isTopicSpecific: isTopicSpecificProducer,
		isBatch:         config.Batch,
	}
	if config.Batch {
		autoFlushContext, cancel := context.WithCancel(ctx)
		k.autoFlushCancel = cancel
		k.wg.Add(1)
		go k.autoFlush(autoFlushContext)
		logger.Notice(ctx, config.ModuleName+" is set to batch mode", nil)
	}
	return k, nil
}

// ProduceMessage writes a message (utils.Message) to the topic with the given key and headers.
// Appends correlation and user identity header.
func (k *Producer) ProduceMessage(ctx context.Context, key string, message *utils.Message, headers map[string]string) (err error) {
	blob, err := json.Marshal(message)
	if err != nil {
		k.log.Error(ctx, "Failed to encode message", err)
		k.log.Error(ctx, "Message", message)
		return fmt.Errorf("Producer.ProduceMessage: error marshalling message: %w", err)
	}
	return k.Produce(ctx, k.topic, key, blob, headers)
}

// ProduceMessageWithTopic writes a message (utils.Message) to a specific topic with the given key and headers.
// Appends correlation and user identity header.
func (k *Producer) ProduceMessageWithTopic(ctx context.Context, topic, key string, message *utils.Message, headers map[string]string) (err error) {
	blob, err := json.Marshal(message)
	if err != nil {
		k.log.Error(ctx, "Failed to encode message", err)
		k.log.Error(ctx, "Message", message)
		return fmt.Errorf("Producer.ProduceMessageWithTopic: error marshalling message: %w", err)
	}
	return k.Produce(ctx, topic, key, blob, headers)
}

// Produce writes a message to a specific topic with the given key and headers.
// Appends correlation and user identity header.
func (k *Producer) Produce(ctx context.Context, topic, key string, message []byte, headers map[string]string) (err error) {
	if k.isTopicSpecific && topic != k.topic {
		err := fmt.Errorf("Producer.Produce: topic is set for producer use `Producer.ProduceMessage` method")
		k.log.Error(ctx, "topic is set for producer use `Producer.ProduceMessage` method", err)
		return err
	}
	if headers == nil {
		headers = make(map[string]string, 0)
	}
	corr := correlation.GetHeader(ctx)
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
	if !k.isTopicSpecific {
		msg.Topic = topic
	}
	err = k.Send(ctx, msg)
	if err == ErrWriterBufferFull {
		err = k.Flush(ctx)
		if err != nil {
			return err
		}
		return k.Send(ctx, msg)
	}
	return nil
}

// autoFlush handles time-based background writes to the broker in case of batch producer.
func (k *Producer) autoFlush(ctx context.Context) {
	defer k.wg.Done()
	nCtx := context.WithoutCancel(ctx)
	timeout, _ := context.WithTimeout(context.Background(), time.Duration(k.config.AutoFlushInterval*uint64(time.Millisecond)))
	defer k.log.Notice(ctx, "auto flush stopped", nil)
	for {
		select {
		case <-timeout.Done():
			err := k.Flush(ctx)
			if err != nil {
				k.log.Emergency(ctx, "Error while writing kafka message", fmt.Errorf("Producer.autoFlush: %w", err), nil)
			}
			timeout, _ = context.WithTimeout(context.Background(), time.Duration(k.config.AutoFlushInterval*uint64(time.Millisecond)))
		case <-ctx.Done():
			err := k.Flush(nCtx)
			if err != nil {
				k.log.Error(nCtx, "error in auto flush", err)
			}
			return
		}
	}
}

// Close gracefully closes the Producer, ensuring all messages are flushed.
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

// Name returns the module name of the Producer.
func (k *Producer) Name(ctx context.Context) string {
	return k.config.ModuleName
}

// Shutdown gracefully shuts down the Producer, ensuring all messages are flushed.
func (k *Producer) Shutdown(ctx context.Context) error {
	return k.Close(ctx)
}
