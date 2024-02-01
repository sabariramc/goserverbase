package api

import (
	"context"
	"fmt"
	"sync"

	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/segmentio/kafka-go"
)

type Reader struct {
	*kafka.Reader
	log              log.Log
	commitLock       sync.Mutex
	consumedMessages []kafka.Message
	bufferSize       int
	idx              int
}

var ErrReaderBufferFull = fmt.Errorf("Reader.StoreMessage: Buffer full")

func NewReader(ctx context.Context, log log.Log, r *kafka.Reader, bufferSize int) *Reader {
	return &Reader{
		Reader:           r,
		log:              log.NewResourceLogger("KafkaReader"),
		consumedMessages: make([]kafka.Message, bufferSize),
		bufferSize:       bufferSize,
		idx:              0,
	}
}

func (k *Reader) Commit(ctx context.Context) error {
	k.commitLock.Lock()
	defer k.commitLock.Unlock()
	if k.idx == 0 {
		return nil
	}
	k.log.Notice(ctx, "committing messages", k.idx)
	err := k.CommitMessages(ctx, k.consumedMessages[:k.idx]...)
	k.idx = 0
	if err != nil {
		k.log.Error(ctx, "error in commit", err)
		return fmt.Errorf("kafka.Reader.Commit: error committing message: %w", err)
	}
	k.log.Notice(ctx, "messages committed", nil)
	return nil
}

func (k *Reader) StoreMessage(ctx context.Context, msg *kafka.Message) error {
	if k.idx >= k.bufferSize {
		return ErrReaderBufferFull
	}
	k.commitLock.Lock()
	defer k.commitLock.Unlock()
	k.consumedMessages[k.idx] = *msg
	k.idx++
	return nil
}

func (k *Reader) Close(ctx context.Context) error {
	err := k.Reader.Close()
	if err != nil {
		k.log.Error(ctx, "error in closing reader", err)
		return fmt.Errorf("kafka.Reader.Close: error in closing reader: %w", err)
	}
	return nil
}
