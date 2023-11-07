package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sabariramc/goserverbase/v3/kafka/api"
	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/sabariramc/goserverbase/v3/utils"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	*api.Writer
	config          KafkaProducerConfig
	log             *log.Logger
	topic           string
	serviceName     string
	resourceName    string
	autoFlushCancel context.CancelFunc
}

func NewProducer(ctx context.Context, logger *log.Logger, config *KafkaProducerConfig, resourceName, topic string) (*Producer, error) {
	if config.MaxBuffer == 0 {
		config.MaxBuffer = 100
	}
	if config.AutoFlushIntervalInMs == 0 {
		config.AutoFlushIntervalInMs = 1000
	}
	logger = logger.NewResourceLogger(resourceName)
	kLog := &kafkaLogger{
		Logger:  logger,
		ctx:     log.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam(config.ServiceName+"--"+resourceName)),
		isError: false,
	}
	p := &kafka.Writer{
		Addr:     kafka.TCP(config.Brokers...),
		Topic:    topic,
		Balancer: &kafka.Hash{},
		Transport: &kafka.Transport{
			SASL: config.SASLMechanism,
			TLS:  config.TLSConfig,
		},
		Logger: kLog,
		ErrorLogger: &kafkaLogger{
			isError: true,
			Logger:  logger,
			ctx:     log.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam(config.ServiceName+"--"+resourceName)),
		},
		Completion:   kLog.DeliveryReport,
		BatchSize:    config.MaxBuffer,
		RequiredAcks: kafka.RequiredAcks(config.Acknowledge),
	}
	k := &Producer{
		serviceName:  config.ServiceName,
		resourceName: resourceName,
		log:          logger,
		config:       *config,
		Writer:       api.NewWriter(ctx, p, config.MaxBuffer, *logger),
		topic:        topic,
	}
	autoFlushContext, cancel := context.WithCancel(log.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam(config.ServiceName+"--"+resourceName)))
	k.autoFlushCancel = cancel
	go k.autoFlush(autoFlushContext)
	return k, nil
}

func (k *Producer) ProduceMessage(ctx context.Context, key string, message *utils.Message, headers map[string]string) (err error) {
	blob, err := json.Marshal(message)
	if err != nil {
		k.log.Error(ctx, "Failed to encode message", err)
		k.log.Error(ctx, "Message", message)
		return fmt.Errorf("kafka.Producer.ProduceMessage: %w", err)
	}
	return k.Produce(ctx, key, blob, headers)
}

func (k *Producer) Produce(ctx context.Context, key string, message []byte, headers map[string]string) (err error) {
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
	k.log.Info(ctx, "Message", map[string]any{"key": key, "headers": headers, "topic": k.topic})
	k.log.Debug(ctx, "Message Body", func() string { return string(message) })
	return k.Send(ctx, key, message, messageHeader)
}

func (k *Producer) autoFlush(ctx context.Context) {
	timeout, _ := context.WithTimeout(context.Background(), time.Duration(k.config.AutoFlushIntervalInMs*uint64(time.Millisecond)))
	select {
	case <-timeout.Done():
		err := k.Flush(ctx)
		if err != nil {
			k.log.Error(ctx, "Error while writing kafka message", err)
		}
		timeout, _ = context.WithTimeout(context.Background(), time.Duration(k.config.AutoFlushIntervalInMs*uint64(time.Millisecond)))
	case <-ctx.Done():
		return
	}
}

func (k *Producer) Close(ctx context.Context) {
	k.log.Notice(ctx, "Producer closer initiated for topic", k.topic)
	k.autoFlushCancel()
	k.Flush(ctx)
	k.Writer.Close()
	k.log.Notice(ctx, "Producer closed for topic", k.topic)
}
