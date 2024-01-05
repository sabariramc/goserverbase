package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	autoCommitCancel context.CancelFunc
	wg               sync.WaitGroup
}

func NewConsumer(ctx context.Context, logger *log.Logger, config KafkaConsumerConfig, topics ...string) (*Consumer, error) {
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
		QueueCapacity:     config.MaxBuffer,
		MaxBytes:          10e6, // 10MB,
		Logger: &kafkaLogger{
			Logger:  logger.NewResourceLogger("KafkaConsumerInfoLog"),
			ctx:     log.GetContextWithCorrelation(context.Background(), defaultCorrelationParam),
			isError: false,
		},
		ErrorLogger: &kafkaLogger{
			Logger:  logger.NewResourceLogger("KafkaConsumerErrorLog"),
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
		config:      config,
		Reader:      api.NewReader(ctx, *logger, r, config.MaxBuffer),
		topics:      topics,
		serviceName: config.ServiceName,
	}
	if k.config.AutoCommit {
		commitCtx, cancel := context.WithCancel(log.GetContextWithCorrelation(context.Background(), defaultCorrelationParam))
		k.autoCommitCancel = cancel
		k.wg.Add(1)
		go k.autoCommit(commitCtx)
	}
	return k, nil
}

func (k *Consumer) Poll(ctx context.Context, ch chan<- *kafka.Message) error {
	var pollErr, commitErr error
	defer close(ch)
	k.log.Info(ctx, fmt.Sprintf("Polling started for topics : %v", k.topics), nil)
	nCtx := context.WithoutCancel(ctx)
outer:
	for {
		select {
		case <-ctx.Done():
			commitErr = k.commit(nCtx)
			k.log.Notice(ctx, "Polling Timeout/cancelled", nil)
			break outer
		default:
			msg, err := k.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					commitErr = k.commit(nCtx)
					break outer
				}
				k.log.Error(ctx, "error fetching message", err)
				pollErr = fmt.Errorf("Consumer.Poll: error fetching message: %w", err)
				commitErr = k.commit(nCtx)
				break outer
			}
			ch <- &msg
			pollErr = k.storeMessage(ctx, &msg)
			if pollErr == api.ErrReaderBufferFull {
				commitErr = k.Commit(ctx)
				if commitErr != nil {
					break outer
				}
				err = k.storeMessage(ctx, &msg)
				if err != nil {
					pollErr = err
					break outer
				}
			} else if pollErr != nil {
				break outer
			}
		}
	}
	if commitErr != nil || pollErr != nil {
		if pollErr == nil {
			pollErr = commitErr
		} else if commitErr != nil {
			pollErr = fmt.Errorf("%w , commitError: %w", pollErr, commitErr)
		}
		k.log.Error(ctx, "error in consumer poll", pollErr)
	}
	k.log.Notice(ctx, fmt.Sprintf("Polling ended for topic : %v", k.topics), nil)
	return pollErr
}

func (k *Consumer) storeMessage(ctx context.Context, msg *kafka.Message) error {
	if k.config.AutoCommit {
		return k.StoreMessage(ctx, msg)
	}
	return nil
}

func (k *Consumer) commit(ctx context.Context) error {
	if k.config.AutoCommit {
		return k.Commit(ctx)
	}
	return nil
}

func (k *Consumer) autoCommit(ctx context.Context) {
	timeout, _ := context.WithTimeout(context.Background(), time.Duration(k.config.AutoCommitIntervalInMs*uint64(time.Millisecond)))
	defer k.wg.Done()
	defer k.log.Warning(ctx, "auto commit stopped", nil)
	nCtx := context.WithoutCancel(ctx)
	for {
		select {
		case <-timeout.Done():
			err := k.Commit(ctx)
			if err != nil {
				k.log.Emergency(ctx, "Error while writing kafka message", fmt.Errorf("Consumer.autoCommit: %w", err), nil)
			}
			timeout, _ = context.WithTimeout(context.Background(), time.Duration(k.config.AutoCommitIntervalInMs*uint64(time.Millisecond)))
		case <-ctx.Done():
			err := k.Commit(nCtx)
			if err != nil {
				k.log.Error(nCtx, "error in auto commit", err)
			}
			return
		}
	}
}

func (k *Consumer) Close(ctx context.Context) error {
	k.log.Notice(ctx, "Consumer closer initiated for topic", k.topics)
	if k.config.AutoCommit {
		k.autoCommitCancel()
	}
	closeErr := k.Reader.Close(ctx)
	if closeErr != nil {
		k.log.Error(ctx, fmt.Sprintf("Consumer closed with error for topic : %v", k.topics), closeErr)
		return fmt.Errorf("Consumer.Close: %w", closeErr)
	}
	k.wg.Wait()
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
