package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/utils"
)

type KafkaProducer struct {
	*kafka.Producer
	config *KafkaProducerConfig
	log    *log.Logger
	topic  string
}

func NewKafkaProducer(ctx context.Context, log *log.Logger, config *KafkaProducerConfig, topic string) (*KafkaProducer, error) {
	parsedConfig := &kafka.ConfigMap{}
	utils.StrictJsonTransformer(config, parsedConfig)
	p, err := kafka.NewProducer(parsedConfig)

	if err != nil {
		log.Error(ctx, "Failed to create kafka producer", err)
		return nil, fmt.Errorf("kafka.NewKafkaProducer.CreateProducer: %w", err)
	}
	k := &KafkaProducer{
		config:   config,
		log:      log,
		Producer: p,
		topic:    topic,
	}
	return k, nil
}

func (k *KafkaProducer) Produce(ctx context.Context, key string, message *utils.Message) (m *kafka.Message, err error) {
	var buf bytes.Buffer
	deliveryChannel := make(chan kafka.Event)
	defer close(deliveryChannel)
	err = json.NewEncoder(&buf).Encode(message)
	if err != nil {
		k.log.Error(ctx, "Failed to encode message", err)
		k.log.Error(ctx, "Message", message)
		return nil, fmt.Errorf("KafkaProducer.Send.EncodeMessage: %w", err)
	}
	correlationParam := log.GetCorrelationParam(ctx)
	headers := make(map[string]string, 0)
	utils.StrictJsonTransformer(correlationParam, &headers)
	messageHeader := make([]kafka.Header, 0)
	for i, v := range headers {
		messageHeader = append(messageHeader, kafka.Header{
			Key:   i,
			Value: []byte(v),
		})
	}
	k.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          buf.Bytes(),
		Headers:        messageHeader,
		Timestamp:      time.Now(),
	}, deliveryChannel)
	e := <-deliveryChannel
	m = e.(*kafka.Message)
	err = m.TopicPartition.Error
	if err != nil {
		k.log.Error(ctx, "Send failed for topic: "+k.topic, err)
		return nil, fmt.Errorf("KafkaProducer.Send.ProduceMessage: %w", err)
	}
	k.log.Info(ctx, "Send success for topic: "+k.topic, m)
	return m, nil
}
