package api

import (
	"context"
	"fmt"
	"sync"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/segmentio/kafka-go"
)

type Reader struct {
	*kafka.Reader
	log              log.Logger
	commitLock       sync.Mutex
	consumedMessages []kafka.Message
	msgCh            chan *kafka.Message
	bufferSize       uint64
	cancelPoll       context.CancelFunc
}

func NewReader(ctx context.Context, log log.Logger, r *kafka.Reader, bufferSize uint64) *Reader {
	return &Reader{
		Reader:           r,
		log:              *log.NewResourceLogger("KafkaReader"),
		consumedMessages: make([]kafka.Message, 0, bufferSize),
		msgCh:            make(chan *kafka.Message, bufferSize),
		bufferSize:       bufferSize,
	}
}

func (k *Reader) Commit(ctx context.Context) error {
	k.commitLock.Lock()
	defer k.commitLock.Unlock()
	if len(k.consumedMessages) == 0 {
		return nil
	}
	k.log.Debug(ctx, "committing messages", k.consumedMessages)
	err := k.CommitMessages(ctx, k.consumedMessages...)
	if err != nil {
		k.log.Error(ctx, "error in commit", err)
		return fmt.Errorf("kafka.Reader.Commit: error committing message: %w", err)
	}
	k.consumedMessages = make([]kafka.Message, 0, k.bufferSize)
	return nil
}

func (k *Reader) Poll(ctx context.Context) error {
	pollCtx, pollCancel := context.WithCancel(ctx)
	k.cancelPoll = pollCancel
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-pollCtx.Done():
			return nil
		default:
			m, err := k.FetchMessage(ctx)
			if err != nil {
				k.log.Error(ctx, "error fetching message", err)
				return fmt.Errorf("kafka.Reader.Poll: error fetching message: %w", err)
			}
			k.msgCh <- &m
		}
	}
}

func (k *Reader) GetEventChannel() <-chan *kafka.Message {
	return k.msgCh
}

func (k *Reader) StoreMessage(ctx context.Context, msg *kafka.Message) {
	k.commitLock.Lock()
	k.consumedMessages = append(k.consumedMessages, *msg)
	k.commitLock.Unlock()
}

func (k *Reader) Close(ctx context.Context) error {
	if k.cancelPoll != nil {
		k.cancelPoll()
	}
	close(k.msgCh)
	err := k.Reader.Close()
	if err != nil {
		k.log.Error(ctx, "error in closing reader", err)
		return fmt.Errorf("kafka.Reader.Close: error in closing reader: %w", err)
	}
	return nil
}
