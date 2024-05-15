package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/segmentio/kafka-go"
)

type ConsumerTracer interface {
	span.SpanOp
}

// Reader extends kafka.Reader(github.com/segmentio/kafka-go) with batch commit, tracing and StatusCheck hook
type Reader struct {
	*kafka.Reader
	log              log.Log
	commitLock       sync.Mutex
	consumedMessages []kafka.Message
	bufferSize       int
	idx              int
	tr               ConsumerTracer
}

var ErrReaderBufferFull = fmt.Errorf("Reader.StoreMessage: Buffer full")

func NewReader(ctx context.Context, log log.Log, r *kafka.Reader, bufferSize int, tr ConsumerTracer) *Reader {
	return &Reader{
		Reader:           r,
		log:              log.NewResourceLogger("KafkaReader"),
		consumedMessages: make([]kafka.Message, bufferSize),
		bufferSize:       bufferSize,
		idx:              0,
		tr:               tr,
	}
}

func (k *Reader) Commit(ctx context.Context) error {
	k.commitLock.Lock()
	defer k.commitLock.Unlock()
	if k.idx == 0 {
		return nil
	}
	if k.tr != nil {
		var crSpan span.Span
		ctx, crSpan = k.tr.NewSpanFromContext(ctx, "kafka.consumer.commit", span.SpanKindConsumer, "")
		defer crSpan.Finish()
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

func (k *Reader) StatusCheck(ctx context.Context) (any, error) {
	return k.Stats(), nil
}
