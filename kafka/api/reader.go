package api

import (
	"context"
	"fmt"
	"sync"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/segmentio/kafka-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type Reader struct {
	*kafka.Reader
	log              log.Logger
	commitLock       sync.Mutex
	consumedMessages []kafka.Message
	msgCh            chan *kafka.Message
	bufferSize       uint64
	cancelPoll       context.CancelFunc
	idx              int
}

func NewReader(ctx context.Context, log log.Logger, r *kafka.Reader, bufferSize uint64) *Reader {
	return &Reader{
		Reader:           r,
		log:              *log.NewResourceLogger("KafkaReader"),
		consumedMessages: make([]kafka.Message, bufferSize),
		msgCh:            make(chan *kafka.Message, bufferSize),
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
	opts := []tracer.StartSpanOption{
		tracer.Tag(ext.SpanKind, ext.SpanKindInternal),
		tracer.Measured(),
	}
	span, ctx := tracer.StartSpanFromContext(ctx, "kafka.consume.commit", opts...)
	defer span.Finish()
	corr := log.GetCorrelationParam(ctx)
	span.SetTag("correlationId", corr.CorrelationId)
	k.log.Debug(ctx, "committing messages", k.consumedMessages)
	err := k.CommitMessages(ctx, k.consumedMessages[:k.idx]...)
	k.idx = 0
	if err != nil {
		k.log.Error(ctx, "error in commit", err)
		return fmt.Errorf("kafka.Reader.Commit: error committing message: %w", err)
	}
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

func (k *Reader) StoreMessage(ctx context.Context, msg *kafka.Message) error {
	k.commitLock.Lock()
	k.consumedMessages[k.idx] = *msg
	k.idx++
	k.commitLock.Unlock()
	if k.idx >= int(k.bufferSize) {
		return k.Commit(ctx)
	}
	return nil
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
