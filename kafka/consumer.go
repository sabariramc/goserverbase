package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v3/kafka/api"
	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/sabariramc/goserverbase/v3/utils"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	*api.Reader
	config           KafkaConsumerConfig
	log              *log.Logger
	topics           []string
	serviceName      string
	resourceName     string
	wg               sync.WaitGroup
	msgCh            chan *kafka.Message
	commitLock       sync.Mutex
	consumedMessages []kafka.Message
}

func NewConsumer(ctx context.Context, logger *log.Logger, config *KafkaConsumerConfig, resourceName string, topics ...string) (*Consumer, error) {
	if config.MaxBuffer <= 0 {
		config.MaxBuffer = 100
	}
	if config.AutoCommitIntervalInMs <= 0 {
		config.AutoCommitIntervalInMs = 1000
	}
	if config.ConsumerLagToleranceInMs <= 0 {
		config.ConsumerLagToleranceInMs = 1000
	}
	logger = logger.NewResourceLogger(resourceName)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           config.Brokers,
		GroupID:           config.GroupID,
		GroupTopics:       topics,
		HeartbeatInterval: time.Second,
		MaxBytes:          10e6, // 10MB,
		Logger: &kafkaLogger{
			Logger:  logger,
			ctx:     log.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam(config.ServiceName+"--"+resourceName)),
			isError: false,
		},
		ErrorLogger: &kafkaLogger{
			Logger:  logger,
			ctx:     log.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam(config.ServiceName+"--"+resourceName)),
			isError: true,
		},
		Dialer: &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: config.SASLMechanism,
			TLS:           config.TLSConfig,
		},
	})
	k := &Consumer{
		log:          logger.NewResourceLogger(resourceName),
		resourceName: resourceName,
		config:       *config,
		Reader:       api.NewReader(ctx, *logger, r, config.MaxBuffer),
		topics:       topics,
		serviceName:  config.ServiceName,
	}
	return k, nil
}

func (k *Consumer) Poll(ctx context.Context, ch chan *kafka.Message) error {
	k.msgCh = make(chan *kafka.Message, k.config.MaxBuffer)
	pollCtx, cancelPoll := context.WithCancel(ctx)
	var pollErr, commitErr error
	k.wg.Add(1)
	go func() {
		defer close(k.msgCh)
		defer k.wg.Done()
		pollErr = k.Reader.Poll(pollCtx)
	}()
	defer close(ch)
	defer k.wg.Wait()
	k.log.Info(ctx, fmt.Sprintf("Polling started for topics : %v", k.topics), nil)
	commitTimeout, commitNow := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(k.config.AutoCommitIntervalInMs))
	infoConsumerLag := time.Millisecond * time.Duration(k.config.ConsumerLagToleranceInMs)
	noticeConsumerLag := 2 * infoConsumerLag
	warningConsumerLag := 2 * noticeConsumerLag
	var count uint64
	count = 0
outer:
	for {
		select {
		case <-ctx.Done():
			cancelPoll()
			commitErr = k.Commit(ctx)
			k.log.Notice(ctx, "Polling Timeout/cancelled", nil)
			break outer
		case <-commitTimeout.Done():
			if k.config.AutoCommit {
				commitErr = k.Commit(ctx)
				if commitErr != nil {
					cancelPoll()
					break outer
				}
			}
			count = 0
			commitTimeout, commitNow = context.WithTimeout(context.Background(), time.Millisecond*time.Duration(k.config.AutoCommitIntervalInMs))
		case msg, ok := <-k.msgCh:
			if !ok {
				cancelPoll()
				commitErr = k.Commit(ctx)
				break outer
			}
			count++
			ch <- msg
			consumerLag := time.Since(msg.Time)
			if consumerLag > infoConsumerLag {
				k.log.Info(ctx, "consumer lag in ms", consumerLag.Milliseconds())
			} else if consumerLag > noticeConsumerLag {
				k.log.Notice(ctx, "consumer lag in ms", consumerLag.Milliseconds())
			} else if consumerLag > warningConsumerLag {
				k.log.Warning(ctx, "consumer lag in ms", consumerLag.Milliseconds())
			}
			if k.config.AutoCommit {
				k.StoreMessage(ctx, msg)
			}
			if count >= k.config.MaxBuffer {
				commitNow()
				count = 0
			}
		}
	}
	if commitErr != nil {
		if pollErr == nil {
			pollErr = commitErr
		} else {
			pollErr = fmt.Errorf("kafka.Consumer.Poll.Error: %w, %w", pollErr, commitErr)
		}
	}
	k.log.Notice(ctx, fmt.Sprintf("Polling ended for topic : %v", k.topics), nil)
	return pollErr
}

func (k *Consumer) Close(ctx context.Context) error {
	k.log.Notice(ctx, "Consumer closer initiated for topic", k.topics)
	commitErr := k.Commit(ctx)
	if k.msgCh != nil {
		_, ok := <-k.msgCh
		if ok {
			close(k.msgCh)
		}
	}
	closeErr := k.Reader.Close()
	k.wg.Wait()
	k.log.Notice(ctx, "Consumer closed for topic", k.topics)
	if commitErr != nil || closeErr != nil {
		k.log.Error(ctx, fmt.Sprintf("Consumer closed with error for topic : %v", k.topics), closeErr)
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
