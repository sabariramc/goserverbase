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

	"github.com/sabariramc/goserverbase/config"
)

type KafkaConsumer struct {
	*kafka.Consumer
	config *config.KafkaConsumerConfig
	log    *log.Logger
	topic  string
	ready  bool
}

func NewKafkaConsumer(ctx context.Context, log *log.Logger, config *config.KafkaConsumerConfig, topic string) (*KafkaConsumer, error) {
	parsedConfig := &kafka.ConfigMap{}
	utils.StrictJsonTransformer(config, parsedConfig)
	c, err := kafka.NewConsumer(parsedConfig)

	if err != nil {
		log.Error(ctx, "Failed to create kafka consumer", err)
		return nil, fmt.Errorf("kafka.NewKafkaConsumer.CreateConsumer: %w", err)
	}
	k := &KafkaConsumer{
		config:   config,
		log:      log,
		Consumer: c,
		topic:    topic,
	}
	err = k.SubscribeTopics([]string{topic}, k.logRebalance)
	if err != nil {
		k.log.Error(ctx, "Failed to create kafka consumer subscription", err)
		return nil, fmt.Errorf("kafka.NewKafkaConsumer.SubscribeTopics: %w", err)
	}
	return k, nil
}

func (k *KafkaConsumer) logRebalance(consumer *kafka.Consumer, e kafka.Event) error {
	k.log.Notice(context.Background(), fmt.Sprintf("Rebalance Event for topic %v", k.topic), e.String())
	return nil
}

func (k *KafkaConsumer) Poll(ctx context.Context, timeout int, outChannel chan *kafka.Message) error {
	defer close(outChannel)
	var err error
	k.log.Info(ctx, "Polling started for topic : "+k.topic, nil)
outer:
	for {
		select {
		case <-ctx.Done():
			k.log.Notice(ctx, "Polling Timeout/cancelled", nil)
			break outer
		default:
			ev := k.Consumer.Poll(timeout)
			switch e := ev.(type) {
			case *kafka.Message:
				outChannel <- e
				k.log.Debug(ctx, "Polling result", e)
			case kafka.PartitionEOF:
				k.log.Info(ctx, "Reached EOF, Ending poll", e)
				break outer
			case kafka.Error:
				k.log.Error(ctx, "Poll error", e)
				err = fmt.Errorf("KafkaConsumer.Poll: %w", err)
				break outer
			default:
				k.log.Debug(ctx, "Polling next message from topic: "+k.topic, e)
			}
		}
	}
	k.log.Info(ctx, "Polling ended for topic : "+k.topic, nil)
	return err
}

func (k *KafkaConsumer) ReadMessage(ctx context.Context, timeout time.Duration) (*kafka.Message, *utils.Message, error) {
	ev, err := k.Consumer.ReadMessage(timeout)
	if err != nil {
		k.log.Error(ctx, "Error reading message from topic: "+k.topic, err)
		return nil, nil, fmt.Errorf("KafkaConsumer.ReadMessage: %w", err)
	}
	k.log.Debug(ctx, "Message read from topic: "+k.topic, ev)
	msg, err := LoadMessage(ev)
	if err != nil {
		k.log.Error(ctx, "Error decoding message from topic: "+k.topic, err)
		k.log.Error(ctx, "Message Content from topic: "+k.topic, string(ev.Value))
		return nil, nil, fmt.Errorf("KafkaConsumer.ReadMessage.Decode: %w", err)
	}
	return ev, msg, nil
}

func (k *KafkaConsumer) Close(ctx context.Context) error {
	err := k.Consumer.Close()
	if err != nil {
		k.log.Error(ctx, "Error closing message from topic: "+k.topic, err)
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
