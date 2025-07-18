package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/utils"
	"github.com/segmentio/kafka-go"
)

// Poller is a high-level API that extends Reader with time and count-based auto commit
// and implements a shutdown hook.
/*
If Reader is not set in [ConsumerConfig] then creates a new [kafka.Reader] with the options passed to the function, and adds addition params

		readerConfig := kafka.ReaderConfig{
			Brokers:           config.Brokers,
			GroupID:           config.GroupID,
			GroupTopics:       config.Topics,
			HeartbeatInterval: time.Second,
			QueueCapacity:     int(config.MaxBuffer),
			MaxBytes:          10e6, // 10MB,
			Dialer: &kafka.Dialer{
				Timeout:       10 * time.Second,
				DualStack:     true,
				SASLMechanism: config.SASLMechanism,
				TLS:           config.TLSConfig,
			},
			Logger: &kafkaLogger{
				Log:     logger.NewResourceLogger(config.ModuleName + ":InfoLog"),
				ctx:     ctx,
				isError: false,
			},
			ErrorLogger: &kafkaLogger{
				Log:     logger.NewResourceLogger(config.ModuleName + ":ErrorLog"),
				ctx:     ctx,
				isError: true,
			},
		}
		config.Reader = kafka.NewReader(readerConfig)
*/
type Poller struct {
	*Reader
	consumerCount    uint // Number of consumers in the group
	config           *ConsumerConfig
	log              log.Log
	topics           []string
	autoCommitCancel context.CancelFunc
	wg               sync.WaitGroup
}

// NewPoller creates a new Poller with the provided consumer options.
func NewPoller(options ...ConsumerOption) (*Poller, error) {
	config := GetDefaultConsumerConfig()
	// Apply options
	for _, opt := range options {
		opt(config)
	}
	logger := config.Log
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), &correlation.CorrelationParam{CorrelationID: config.ModuleName})
	if config.Reader == nil {
		readerConfig := kafka.ReaderConfig{
			Brokers:           config.Brokers,
			GroupID:           config.GroupID,
			GroupTopics:       config.Topics,
			HeartbeatInterval: time.Second,
			QueueCapacity:     int(config.MaxBuffer),
			MaxBytes:          10e6, // 10MB,
			Dialer: &kafka.Dialer{
				Timeout:       10 * time.Second,
				DualStack:     true,
				SASLMechanism: config.SASLMechanism,
				TLS:           config.TLSConfig,
				ClientID:      fmt.Sprintf("%v-%v", config.ClientId, uuid.NewString()),
			},
			Logger: &kafkaLogger{
				Log:     logger.NewResourceLogger(config.ModuleName + ":InfoLog"),
				ctx:     ctx,
				isError: false,
			},
			ErrorLogger: &kafkaLogger{
				Log:     logger.NewResourceLogger(config.ModuleName + ":ErrorLog"),
				ctx:     ctx,
				isError: true,
			},
		}
		config.Reader = kafka.NewReader(readerConfig)
	}
	k := &Poller{
		log:           config.Log,
		consumerCount: 0,
		config:        config,
		Reader:        NewReader(ctx, logger, config.Reader, config.Trace),
		topics:        config.Topics,
	}
	if k.config.AutoCommit {
		commitCtx, cancel := context.WithCancel(ctx)
		k.autoCommitCancel = cancel
		k.wg.Add(1)
		go k.autoCommit(commitCtx)
	}
	return k, nil
}

// Poll fetches messages from the broker and passes them to the provided channel.
// This function is meant to be run as a goroutine.
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
			k.consumerCount++
			k.StoreOffset(ctx, &msg)
			k.commit(ctx)
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

// commit commits the current state if auto-commit is enabled.
func (k *Poller) commit(ctx context.Context) error {
	if k.config.AutoCommit && k.consumerCount >= k.config.MaxBuffer {
		_, err := k.Commit(ctx)
		if err != nil {
			return fmt.Errorf("Poller.commit: error committing message: %w", err)
		}
		k.consumerCount = 0
	}
	return nil
}

// autoCommit handles time-based background commit to broker in case of auto commit poller.
func (k *Poller) autoCommit(ctx context.Context) {
	timeout, _ := context.WithTimeout(context.Background(), time.Duration(k.config.AutoCommitInterval*uint64(time.Millisecond)))
	defer k.wg.Done()
	defer k.log.Warning(ctx, "auto commit stopped", nil)
	nCtx := context.WithoutCancel(ctx)
	for {
		select {
		case <-timeout.Done():
			err := k.commit(ctx)
			if err != nil {
				k.log.Emergency(ctx, "Error while writing kafka message", fmt.Errorf("Consumer.autoCommit: %w", err))
			}
			timeout, _ = context.WithTimeout(context.Background(), time.Duration(k.config.AutoCommitInterval*uint64(time.Millisecond)))
		case <-ctx.Done():
			err := k.commit(nCtx)
			if err != nil {
				k.log.Error(nCtx, "error in auto commit", err)
			}
			return
		}
	}
}

// Close closes the Poller and waits for any ongoing operations to complete.
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

// LoadMessage loads a message from the given Kafka message.
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
