package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v4/kafka/api"
	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/utils"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	*api.Reader
	config           KafkaConsumerConfig
	log              *log.Logger
	topics           []string
	serviceName      string
	wg               sync.WaitGroup
	autoCommitCancel context.CancelFunc
	count            uint64
	countLock        sync.Mutex
}

func NewConsumer(ctx context.Context, logger *log.Logger, config *KafkaConsumerConfig, topics ...string) (*Consumer, error) {
	if config.MaxBuffer <= 0 {
		config.MaxBuffer = 100
	}
	if config.AutoCommitIntervalInMs <= 0 {
		config.AutoCommitIntervalInMs = 1000
	}

	logger = logger.NewResourceLogger("KafkaConsumer")
	defaultCorrelationParam := &log.CorrelationParam{CorrelationId: config.ServiceName + ":KafkaConsumer"}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           config.Brokers,
		GroupID:           config.GroupID,
		GroupTopics:       topics,
		HeartbeatInterval: time.Second,
		MaxBytes:          10e6, // 10MB,
		Logger: &kafkaLogger{
			Logger:  logger,
			ctx:     log.GetContextWithCorrelation(context.Background(), defaultCorrelationParam),
			isError: false,
		},
		ErrorLogger: &kafkaLogger{
			Logger:  logger,
			ctx:     log.GetContextWithCorrelation(context.Background(), defaultCorrelationParam),
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
		log:         logger.NewResourceLogger("KafkaConsumer"),
		config:      *config,
		Reader:      api.NewReader(ctx, *logger, r, config.MaxBuffer),
		topics:      topics,
		serviceName: config.ServiceName,
		count:       0,
	}
	if k.config.AutoCommit {
		commitCtx, cancel := context.WithCancel(log.GetContextWithCorrelation(context.Background(), defaultCorrelationParam))
		k.autoCommitCancel = cancel
		go k.autoCommit(commitCtx)
	}
	return k, nil
}

func (k *Consumer) Poll(ctx context.Context, ch chan<- *kafka.Message) error {
	pollCtx, cancelPoll := context.WithCancel(ctx)
	var pollErr, commitErr error
	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		pollErr = k.Reader.Poll(pollCtx)
	}()
	defer close(ch)
	defer k.wg.Wait()
	k.log.Info(ctx, fmt.Sprintf("Polling started for topics : %v", k.topics), nil)
outer:
	for {
		select {
		case <-ctx.Done():
			cancelPoll()
			commitErr = k.Commit(ctx)
			k.log.Notice(ctx, "Polling Timeout/cancelled", nil)
			break outer
		case msg, ok := <-k.GetEventChannel():
			if !ok {
				cancelPoll()
				commitErr = k.Commit(ctx)
				break outer
			}
			k.incrementCount(1)
			ch <- msg
			if k.config.AutoCommit {
				k.StoreMessage(ctx, msg)
			}
			if k.count >= k.config.MaxBuffer {
				k.countLock.Lock()
				commitErr = k.Commit(ctx)
				if commitErr != nil {
					cancelPoll()
					break outer
				}
				k.resetCount()
			}
		}
	}
	if commitErr != nil {
		if pollErr == nil {
			pollErr = commitErr
		} else {
			pollErr = fmt.Errorf("Consumer.Poll: %w , %w", pollErr, commitErr)
		}
	}
	k.log.Notice(ctx, fmt.Sprintf("Polling ended for topic : %v", k.topics), nil)
	return pollErr
}

func (k *Consumer) incrementCount(i uint64) {
	k.countLock.Lock()
	k.count += i
	k.countLock.Unlock()
}

func (k *Consumer) resetCount() {
	k.count = 0
	k.countLock.Unlock()
}

func (k *Consumer) autoCommit(ctx context.Context) {
	timeout, _ := context.WithTimeout(context.Background(), time.Duration(k.config.AutoCommitIntervalInMs*uint64(time.Millisecond)))
	defer k.log.Warning(ctx, "auto commit stopped", nil)
	for {
		select {
		case <-timeout.Done():
			k.countLock.Lock()
			err := k.Commit(ctx)
			if err != nil {
				k.log.Emergency(ctx, "Error while writing kafka message", fmt.Errorf("Producer.autoFlush: %w", err), nil)
			}
			k.resetCount()
			timeout, _ = context.WithTimeout(context.Background(), time.Duration(k.config.AutoCommitIntervalInMs*uint64(time.Millisecond)))
		case <-ctx.Done():
			return
		}
	}
}

func (k *Consumer) Close(ctx context.Context) error {
	k.log.Notice(ctx, "Consumer closer initiated for topic", k.topics)
	closeErr := k.Reader.Close(ctx)
	k.wg.Wait()
	k.autoCommitCancel()
	commitErr := k.Commit(ctx)
	if commitErr != nil || closeErr != nil {
		k.log.Error(ctx, fmt.Sprintf("Consumer closed with error for topic : %v", k.topics), closeErr)
		return fmt.Errorf("Consumer.Close: %w, %w", commitErr, closeErr)
	}
	k.log.Notice(ctx, "Consumer closed for topic", k.topics)
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
