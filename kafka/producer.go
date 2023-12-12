package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sabariramc/goserverbase/v4/kafka/api"
	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/utils"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	*api.Writer
	config          KafkaProducerConfig
	log             *log.Logger
	topic           string
	serviceName     string
	autoFlushCancel context.CancelFunc
}

func NewProducer(ctx context.Context, logger *log.Logger, config *KafkaProducerConfig, topic string) (*Producer, error) {
	if config.MaxBuffer == 0 {
		config.MaxBuffer = 100
	}
	if config.AutoFlushIntervalInMs == 0 {
		config.AutoFlushIntervalInMs = 1000
	}
	logger = logger.NewResourceLogger("KafkaProducer")
	defaultCorrelationParam := &log.CorrelationParam{CorrelationId: config.ServiceName + ":KafkaProducer"}
	kLog := &kafkaLogger{
		Logger:  logger,
		ctx:     log.GetContextWithCorrelation(context.Background(), defaultCorrelationParam),
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
			ctx:     log.GetContextWithCorrelation(context.Background(), defaultCorrelationParam),
		},
		Completion:   kLog.DeliveryReport,
		BatchSize:    config.MaxBuffer,
		RequiredAcks: kafka.RequiredAcks(config.Acknowledge),
	}
	var writer *api.Writer
	if config.Channeled {
		writer = api.NewChanneledWriter(ctx, p, config.MaxBuffer, *logger)
	} else {
		writer = api.NewWriter(ctx, p, config.MaxBuffer, *logger)
	}
	k := &Producer{
		serviceName: config.ServiceName,
		log:         logger,
		config:      *config,
		Writer:      writer,
		topic:       topic,
	}
	autoFlushContext, cancel := context.WithCancel(log.GetContextWithCorrelation(context.Background(), defaultCorrelationParam))
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
	k.log.Info(ctx, "MessageMeta", map[string]any{"key": key, "headers": headers, "topic": k.topic})
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

func (k *Producer) Close(ctx context.Context) error {
	k.log.Notice(ctx, "Producer closer initiated for topic", k.topic)
	k.autoFlushCancel()
	k.Flush(ctx)
	err := k.Writer.Close(ctx)
	if err == nil {
		k.log.Notice(ctx, "Producer closed for topic", k.topic)
	}
	return err
}
