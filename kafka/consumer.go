package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sabariramc/goserverbase/v2/log"
	"github.com/sabariramc/goserverbase/v2/utils"
)

type Consumer struct {
	*kafka.Consumer
	config *KafkaConsumerConfig
	log    *log.Logger
	topic  []string
}

func NewConsumer(ctx context.Context, log *log.Logger, config *KafkaConsumerConfig, topic ...string) (*Consumer, error) {
	parsedConfig := &kafka.ConfigMap{}
	utils.StrictJsonTransformer(config, parsedConfig)
	c, err := kafka.NewConsumer(parsedConfig)

	if err != nil {
		log.Error(ctx, "Failed to create kafka consumer", err)
		return nil, fmt.Errorf("kafka.NewKafkaConsumer.CreateConsumer: %w", err)
	}
	k := &Consumer{
		config:   config,
		log:      log,
		Consumer: c,
		topic:    topic,
	}
	err = k.SubscribeTopics(topic, k.logReBalance)
	if err != nil {
		k.log.Error(ctx, "Failed to create kafka consumer subscription", err)
		return nil, fmt.Errorf("kafka.NewKafkaConsumer.SubscribeTopics: %w", err)
	}
	return k, nil
}

func (k *Consumer) logReBalance(consumer *kafka.Consumer, e kafka.Event) error {
	k.log.Notice(context.Background(), fmt.Sprintf("Re-balance Event for topic %v", k.topic), e.String())
	return nil
}

func (k *Consumer) Poll(ctx context.Context, timeout int, outChannel chan *kafka.Message) error {
	var err error
	k.log.Info(ctx, fmt.Sprintf("Polling started for topic : %v", k.topic), nil)
outer:
	for {
		select {
		case <-ctx.Done():
			k.log.Notice(ctx, "Polling Timeout/cancelled", nil)
			break outer
		default:
			ev := k.Consumer.Poll(timeout)
			if ev != nil {
				switch e := ev.(type) {
				case *kafka.Message:
					outChannel <- e
				case kafka.PartitionEOF:
					k.log.Error(ctx, "Reached EOF, Ending poll", e)
					err = fmt.Errorf("KafkaConsumer.Poll: EOF: %v", e)
					break outer
				case kafka.Error:
					k.log.Error(ctx, "Poll error", e)
					err = fmt.Errorf("KafkaConsumer.Poll: Error: %w", e)
					break outer
				default:
					k.log.Error(ctx, "Poll error", e)
					err = fmt.Errorf("KafkaConsumer.Poll: Unknown: %v", e.String())
					break outer
				}
			}
		}
	}
	k.log.Warning(ctx, fmt.Sprintf("Polling ended for topic : %v", k.topic), nil)
	return err
}

func (k *Consumer) ReadMessage(ctx context.Context, timeout time.Duration) (*kafka.Message, error) {
	ev, err := k.Consumer.ReadMessage(timeout)
	if err != nil {
		k.log.Error(ctx, fmt.Sprintf("Polling started for topic : %v", k.topic), err)
		return nil, fmt.Errorf("KafkaConsumer.ReadMessage: %w", err)
	}
	return ev, err
}

func (k *Consumer) Close(ctx context.Context) error {
	err := k.Consumer.Close()
	if err != nil {
		k.log.Error(ctx, fmt.Sprintf("Polling started for topic : %v", k.topic), err)
		return fmt.Errorf("KafkaConsumer.Close: %w", err)
	}
	return nil
}

func LoadMessage(src *kafka.Message) (*utils.Message, error) {
	msg := &utils.Message{}
	r := bytes.NewReader(src.Value)
	de := json.NewDecoder(r)
	de.DisallowUnknownFields()
	err := de.Decode(msg)
	if err != nil {
		err = fmt.Errorf("kafka.LoadMessage: %w", err)
	}
	return msg, err
}
