package api

import (
	"context"
	"fmt"
	"sync"

	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/segmentio/kafka-go"
)

type Reader struct {
	*kafka.Reader
	log              log.Logger
	commitLock       sync.Mutex
	consumedMessages []kafka.Message
	msgCh            chan *kafka.Message
	bufferSize       uint64
}

func NewReader(ctx context.Context, log log.Logger, r *kafka.Reader, bufferSize uint64, msgCh chan *kafka.Message) *Reader {
	return &Reader{
		Reader:           r,
		log:              *log.NewResourceLogger("KafkaReader"),
		consumedMessages: make([]kafka.Message, 0, bufferSize),
		msgCh:            msgCh,
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
		return fmt.Errorf("kafka.Consumer.Commit: error during commit : %w", err)
	}
	k.consumedMessages = make([]kafka.Message, 0, k.bufferSize)
	return nil
}

func (k *Reader) Poll(ctx context.Context) error {
	k.log.Debug(ctx, "Starting fetch message", nil)
	for {
		select {
		case <-ctx.Done():
			k.log.Notice(ctx, "Fetch message ended", nil)
			return nil
		default:
			m, err := k.FetchMessage(ctx)
			if err != nil {
				k.log.Error(ctx, "Poll error", err)
				return fmt.Errorf("kafka.Consumer.poll: Error: %w", err)
			}
			k.msgCh <- &m
		}
	}
}

func (k *Reader) StoreMessage(ctx context.Context, msg *kafka.Message) {
	k.commitLock.Lock()
	defer k.commitLock.Unlock()
	k.consumedMessages = append(k.consumedMessages, *msg)
}
