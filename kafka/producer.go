package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sabariramc/goserverbase/v3/errors"
	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/sabariramc/goserverbase/v3/utils"
)

type Producer struct {
	*kafka.Producer
	config      *KafkaProducerConfig
	log         *log.Logger
	topic       string
	deliveryCh  chan kafka.Event
	serviceName string
	wg          sync.WaitGroup
	notifier    errors.ErrorNotifier
}

func createProducer(ctx context.Context, log *log.Logger, config *KafkaProducerConfig, serviceName, topic string) (*Producer, error) {
	parsedConfig := &kafka.ConfigMap{}
	utils.StrictJsonTransformer(config, parsedConfig)
	p, err := kafka.NewProducer(parsedConfig)
	if err != nil {
		log.Error(ctx, "Failed to create kafka producer", err)
		return nil, fmt.Errorf("kafka.createProducer: %w", err)
	}
	return &Producer{
		config:      config,
		log:         log,
		Producer:    p,
		topic:       topic,
		serviceName: serviceName,
	}, nil
}

func NewProducer(ctx context.Context, log *log.Logger, config *KafkaProducerConfig, serviceName, topic string, notifier errors.ErrorNotifier) (*Producer, error) {
	k, err := createProducer(ctx, log, config, serviceName, topic)
	k.deliveryCh = make(chan kafka.Event, config.MaxBuffer+100)
	if notifier != nil {
		_, ok := notifier.GetProcessor().(*Producer)
		if ok {
			return nil, fmt.Errorf("kafka.NewProducer: notifier cannot be of same type")
		}
	}
	k.notifier = notifier
	if err != nil {
		return nil, fmt.Errorf("kafka.NewProducer: %w", err)
	}
	k.wg.Add(1)
	go func() {
		k.deliveryReport()
		k.wg.Done()
	}()
	return k, nil
}

func (k *Producer) ProduceMessage(ctx context.Context, key string, message *utils.Message, headers map[string]string) (err error) {
	blob, err := json.Marshal(message)
	if err != nil {
		k.log.Error(ctx, "Failed to encode message", err)
		k.log.Error(ctx, "Message", message)
		return fmt.Errorf("kafka.Producer.ProduceMessage.EncodeMessage: %w", err)
	}
	return k.Produce(ctx, key, blob, headers)
}

func (k *Producer) handleEvent(defaultCtx context.Context, ev kafka.Event) (context.Context, error) {
	switch e := ev.(type) {
	case *kafka.Message:
		logMsg := &Message{
			Message: e,
		}
		headers := logMsg.GetHeaders()
		corr := &log.CorrelationParam{}
		data, _ := json.Marshal(headers)
		err := utils.HeaderJson.Unmarshal(data, corr)
		var ctx context.Context
		if err != nil || corr.CorrelationId == "" {
			k.log.Error(defaultCtx, "Error extracting header", headers)
			ctx = defaultCtx
		} else {
			ctx = log.GetContextWithCorrelation(context.Background(), corr)
		}
		err = e.TopicPartition.Error
		if err != nil {
			k.log.Error(ctx, "Error in publishing message", err)
			k.log.Error(ctx, "Error Meta", logMsg.GetMeta())
			k.log.Debug(ctx, "Error Body", logMsg.GetBody)
			return ctx, err
		}
		k.log.Info(ctx, "Send success for topic: "+k.topic, logMsg.GetMeta())
		k.log.Debug(ctx, "Send success for topic - body: "+k.topic, logMsg.GetBody)
	case kafka.Error:
		k.log.Error(defaultCtx, "Produce Error", e)
		return defaultCtx, e
	default:
		k.log.Notice(defaultCtx, "KafkaProducer: Event", e.String())
	}
	return nil, nil
}

func (k *Producer) deliveryReport() {
	defaultCtx := log.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam(k.serviceName))
	for ev := range k.deliveryCh {
		ctx, err := k.handleEvent(defaultCtx, ev)
		if err != nil && k.notifier != nil {
			k.notifier.Send5XX(ctx, fmt.Sprintf("com.%v.kafka.Producer.error", k.serviceName), err, "", ev.String())
		}
	}
}

func (k *Producer) Produce(ctx context.Context, key string, message []byte, headers map[string]string) (err error) {
	if k.Producer.Len() >= k.config.MaxBuffer {
		k.Producer.Flush(1000)
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
	k.log.Debug(ctx, "Message - meta", map[string]any{"key": key, "headers": headers})
	k.log.Debug(ctx, "Message - body", func() string { return string(message) })
	err = k.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          message,
		Headers:        messageHeader,
		Timestamp:      time.Now(),
	}, k.deliveryCh)
	if err != nil {
		k.log.Error(ctx, "Failed to enqueue message: "+k.topic, err)
		return fmt.Errorf("kafka.Producer.Produce: %w", err)
	}
	return nil
}

func (k *Producer) Close() {
	k.Producer.Flush(10000)
	close(k.deliveryCh)
	k.wg.Wait()
	k.Producer.Close()
}
