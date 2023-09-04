package kafka

import (
	"bytes"
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

type Consumer struct {
	*kafka.Consumer
	config       *KafkaConsumerConfig
	log          *log.Logger
	notifier     errors.ErrorNotifier
	topic        []string
	serviceName  string
	resourceName string
	logCh        chan kafka.LogEvent
	msgCh        chan *kafka.Message
	wg           sync.WaitGroup
}

func NewConsumer(ctx context.Context, serviceName string, log *log.Logger, config *KafkaConsumerConfig, notifier errors.ErrorNotifier, topic ...string) (*Consumer, error) {
	return NewConsumerResource(ctx, serviceName, "KafkaConsumer", log, config, notifier, topic...)
}

func NewConsumerResource(ctx context.Context, serviceName, resourceName string, log *log.Logger, config *KafkaConsumerConfig, notifier errors.ErrorNotifier, topic ...string) (*Consumer, error) {
	parsedConfig := &kafka.ConfigMap{}
	ch := make(chan kafka.LogEvent, 10000)
	(*parsedConfig)["go.logs.channel.enable"] = true
	(*parsedConfig)["go.logs.channel"] = ch
	(*parsedConfig)["enable.auto.commit"] = false
	utils.StrictJsonTransformer(config, parsedConfig)
	c, err := kafka.NewConsumer(parsedConfig)
	if config.MaxBuffer <= 0 {
		config.MaxBuffer = 1000
	}
	if config.AutoCommitIntervalInMs <= 0 {
		config.AutoCommitIntervalInMs = 1000 * 10
	}
	if config.ConsumerLagToleranceInMs <= 0 {
		config.ConsumerLagToleranceInMs = 1000 * 3
	}
	if err != nil {
		log.Error(ctx, "Failed to create kafka consumer", err)
		return nil, fmt.Errorf("kafka.NewKafkaConsumer.CreateConsumer: %w", err)
	}
	k := &Consumer{
		log:          log.NewResourceLogger(resourceName),
		resourceName: resourceName,
		config:       config,
		Consumer:     c,
		topic:        topic,
		logCh:        ch,
		notifier:     notifier,
		serviceName:  serviceName,
	}
	err = k.SubscribeTopics(topic, k.logReBalance)
	if err != nil {
		k.log.Error(ctx, "Failed to create kafka consumer subscription", err)
		return nil, fmt.Errorf("kafka.NewKafkaConsumer.SubscribeTopics: %w", err)
	}
	k.wg.Add(1)
	go func() {
		k.printKafkaLog()
		k.wg.Done()
	}()
	return k, nil
}

func (k *Consumer) commit(ctx context.Context) error {
	offset, err := k.Consumer.Commit()
	if err != nil {
		if err.Error() == "Local: No offset stored" {
			k.log.Debug(ctx, "No offset to commit", err)
			err = nil
		} else {
			k.log.Error(ctx, "Error on commit", err)
		}
	}
	if len(offset) > 0 {
		k.log.Notice(ctx, "Committed offsets", offset)
	}
	return err
}

func (k *Consumer) logReBalance(consumer *kafka.Consumer, e kafka.Event) error {
	k.log.Notice(context.Background(), fmt.Sprintf("Re-balance Event for topic %v", k.topic), e.String())
	return nil
}

func (k *Consumer) printKafkaLog() {
	defaultCtx := log.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam(k.serviceName+"--"+k.resourceName))
	for kLog := range k.logCh {
		k.log.Log(defaultCtx, kLog.Level, kLog.Message, kLog, fmt.Errorf("%v", kLog.Message))
	}
}

func (k *Consumer) poll(ctx context.Context, timeout int) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			ev := k.Consumer.Poll(timeout)
			if ev != nil {
				switch e := ev.(type) {
				case *kafka.Message:
					if e.TopicPartition.Error != nil {
						return e.TopicPartition.Error
					}
					if e.TopicPartition.Offset < 0 {
						return fmt.Errorf("KafkaConsumer.Poll: offset is less than zero: topic - %v, partition- %v", e.TopicPartition.Topic, e.TopicPartition.Partition)
					}
					k.msgCh <- e
				case kafka.PartitionEOF:
					k.log.Error(ctx, "Reached EOF, Ending poll", e)
					return fmt.Errorf("KafkaConsumer.Poll: EOF: %v", e)
				case kafka.Error:
					k.log.Error(ctx, "Poll error", e)
					return fmt.Errorf("KafkaConsumer.Poll: Error: %w", e)
				case kafka.RevokedPartitions:
					k.log.Notice(ctx, "Partition revoked", e.Partitions)
					if k.notifier != nil {
						k.notifier.Send4XX(ctx, fmt.Sprintf("com.error.%v.kafka.partition.revoked", k.serviceName), nil, "", nil)
					}
				case kafka.AssignedPartitions:
					k.log.Notice(ctx, "Partition assigned", e.Partitions)
					if k.notifier != nil {
						k.notifier.Send4XX(ctx, fmt.Sprintf("com.notice.%v.kafka.partition.assigned", k.serviceName), nil, "", nil)
					}
				case kafka.OffsetsCommitted:
					k.log.Notice(ctx, "KafkaConsumer.Poll: Offset Committed", e.Offsets)
					if e.Error != nil {
						k.log.Error(ctx, "Poll Offset Committed error", e.Error)
						return fmt.Errorf("KafkaConsumer.Poll: Offset Committed Error: %w", e.Error)
					}
				default:
					k.log.Notice(ctx, "KafkaConsumer.Poll: Event", e.String())
				}
			}
		}
	}
}

func (k *Consumer) Poll(ctx context.Context, timeout int, ch chan *kafka.Message) error {
	k.msgCh = make(chan *kafka.Message, k.config.MaxBuffer)
	pollCtx, cancelPoll := context.WithCancel(ctx)
	var pollErr, commitErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer close(k.msgCh)
		defer wg.Done()
		pollErr = k.poll(pollCtx, timeout)
	}()
	defer close(ch)
	defer wg.Wait()
	k.log.Info(ctx, fmt.Sprintf("Polling started for topic : %v", k.topic), nil)
	commitTimeout, _ := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(k.config.AutoCommitIntervalInMs))
	infoConsumerLag := time.Millisecond * time.Duration(k.config.ConsumerLagToleranceInMs)
	noticeConsumerLag := 2 * infoConsumerLag
	warningConsumerLag := 2 * noticeConsumerLag
outer:
	for {
		for i := uint64(0); i < k.config.MaxBuffer; i++ {
			select {
			case <-ctx.Done():
				cancelPoll()
				commitErr = k.commit(ctx)
				k.log.Notice(ctx, "Polling Timeout/cancelled", nil)
				break outer
			case <-commitTimeout.Done():
				commitErr = k.commit(ctx)
				if commitErr != nil {
					cancelPoll()
					break outer
				}
				commitTimeout, _ = context.WithTimeout(context.Background(), time.Millisecond*time.Duration(k.config.AutoCommitIntervalInMs))
			case msg, ok := <-k.msgCh:
				if msg != nil {
					ch <- msg
					consumerLag := time.Since(msg.Timestamp)
					if consumerLag > infoConsumerLag {
						k.log.Info(ctx, "consumer lag in ms", consumerLag.Milliseconds())
					} else if consumerLag > noticeConsumerLag {
						k.log.Notice(ctx, "consumer lag in ms", consumerLag.Milliseconds())
					} else if consumerLag > warningConsumerLag {
						k.log.Warning(ctx, "consumer lag in ms", consumerLag.Milliseconds())
					}
					k.Consumer.StoreMessage(msg)
				}
				if !ok {
					cancelPoll()
					commitErr = k.commit(ctx)
					break outer
				}
			}
		}
		commitErr = k.commit(ctx)
		if commitErr != nil {
			cancelPoll()
			break
		}
	}
	if commitErr != nil {
		if pollErr == nil {
			pollErr = commitErr
		} else {
			pollErr = fmt.Errorf("kafka.Poll.Error: %w, %w", pollErr, commitErr)
		}
	}
	k.log.Notice(ctx, fmt.Sprintf("Polling ended for topic : %v", k.topic), nil)
	return pollErr
}

func (k *Consumer) ReadMessage(ctx context.Context, timeout time.Duration) (*kafka.Message, error) {
	ev, err := k.Consumer.ReadMessage(timeout)
	if err != nil {
		k.log.Error(ctx, fmt.Sprintf("Read message error for topic : %v", k.topic), err)
		return nil, fmt.Errorf("KafkaConsumer.ReadMessage: %w", err)
	}
	offset, err := k.Consumer.StoreMessage(ev)
	if err != nil {
		k.log.Error(ctx, "error storing offset", err)
	}
	k.log.Info(ctx, "stored offset", offset)
	err = k.commit(ctx)
	return ev, err
}

func (k *Consumer) Close(ctx context.Context) error {
	k.log.Notice(ctx, "Consumer closer initiated for topic", k.topic)
	commitErr := k.commit(ctx)
	close(k.logCh)
	if k.msgCh != nil {
		_, ok := <-k.msgCh
		if ok {
			close(k.msgCh)
		}
	}
	closeErr := k.Consumer.Close()
	k.wg.Wait()
	k.log.Notice(ctx, "Consumer closed for topic", k.topic)
	if commitErr != nil || closeErr != nil {
		k.log.Error(ctx, fmt.Sprintf("Consumer closed with error for topic : %v", k.topic), closeErr)
		return fmt.Errorf("KafkaConsumer.Close: %w, %w", commitErr, closeErr)
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
