package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v5/kafka/api"
	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/utils"
	"github.com/segmentio/kafka-go"
)

type Poller struct {
	*api.Reader
	config           KafkaConsumerConfig
	log              log.Log
	topics           []string
	autoCommitCancel context.CancelFunc
	wg               sync.WaitGroup
}

func NewPoller(ctx context.Context, logger log.Log, config KafkaConsumerConfig, tr api.ConsumerTracer, topics ...string) (*Poller, error) {
	if config.MaxBuffer <= 0 {
		config.MaxBuffer = 100
	}
	if config.AutoCommitIntervalInMs <= 0 {
		config.AutoCommitIntervalInMs = 1000
	}
	logger = logger.NewResourceLogger("KafkaConsumer")
	defaultCorrelationParam := &log.CorrelationParam{CorrelationID: "KafkaConsumer"}
	readerConfig := kafka.ReaderConfig{
		Brokers:           config.Brokers,
		GroupID:           config.GroupID,
		GroupTopics:       topics,
		HeartbeatInterval: time.Second,
		QueueCapacity:     config.MaxBuffer,
		MaxBytes:          10e6, // 10MB,
		Dialer: &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: config.SASLMechanism,
			TLS:           config.TLSConfig,
		},
	}
	if config.EnableLog {
		readerConfig.Logger = &kafkaLogger{
			Log:     logger.NewResourceLogger("KafkaConsumerInfoLog"),
			ctx:     log.GetContextWithCorrelationParam(context.Background(), defaultCorrelationParam),
			isError: false,
		}
		readerConfig.ErrorLogger = &kafkaLogger{
			Log:     logger.NewResourceLogger("KafkaConsumerErrorLog"),
			ctx:     log.GetContextWithCorrelationParam(context.Background(), defaultCorrelationParam),
			isError: true,
		}
	}
	r := kafka.NewReader(readerConfig)
	k := &Poller{
		log:    logger.NewResourceLogger("KafkaConsumer"),
		config: config,
		Reader: api.NewReader(ctx, logger, r, config.MaxBuffer, tr),
		topics: topics,
	}
	if k.config.AutoCommit {
		commitCtx, cancel := context.WithCancel(log.GetContextWithCorrelationParam(context.Background(), defaultCorrelationParam))
		k.autoCommitCancel = cancel
		k.wg.Add(1)
		go k.autoCommit(commitCtx)
	}
	return k, nil
}

func (k *Poller) Poll(ctx context.Context, ch chan<- *kafka.Message) error {
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
			commitErr = k.storeMessage(ctx, &msg)
			if commitErr != nil {
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

func (k *Poller) storeMessage(ctx context.Context, msg *kafka.Message) error {
	if k.config.AutoCommit {
		err := k.StoreMessage(ctx, msg)
		if err == api.ErrReaderBufferFull {
			err = k.Commit(ctx)
			if err != nil {
				return err
			}
			err = k.StoreMessage(ctx, msg)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (k *Poller) commit(ctx context.Context) error {
	if k.config.AutoCommit {
		return k.Commit(ctx)
	}
	return nil
}

func (k *Poller) autoCommit(ctx context.Context) {
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

func (k *Poller) Close(ctx context.Context) error {
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
