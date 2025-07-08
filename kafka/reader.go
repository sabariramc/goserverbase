package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/segmentio/kafka-go"
)

// ConsumerTracer defines an interface for tracing consumer operations.
type ConsumerTracer interface {
	span.SpanOp
}

type OffsetMap map[string]map[int]int64 // Map to store offsets for each topic and partition

// Reader extends kafka.Reader with batch commit, tracing, and a StatusCheck hook.
type Reader struct {
	*kafka.Reader
	log        log.Log
	commitLock sync.Mutex
	offsetMap  OffsetMap // Buffer to store messages for committing
	tr         ConsumerTracer
}

// ErrReaderBufferFull is returned when the buffer is full and a message cannot be stored.
var ErrReaderBufferFull = fmt.Errorf("Reader.StoreMessage: Buffer full")

// NewReader creates a new Reader with the provided context, logger, Kafka reader, buffer size, and tracer.
func NewReader(ctx context.Context, log log.Log, r *kafka.Reader, tr ConsumerTracer) *Reader {
	return &Reader{
		Reader:    r,
		log:       log.NewResourceLogger("KafkaReader"),
		offsetMap: OffsetMap{},
		tr:        tr,
	}
}

// Commit commits the messages stored in the buffer.
func (k *Reader) Commit(ctx context.Context) (OffsetMap, error) {
	k.commitLock.Lock()
	defer k.commitLock.Unlock()
	if len(k.offsetMap) == 0 {
		k.log.Debug(ctx, "no messages to commit", nil)
		return nil, nil
	}
	if k.tr != nil {
		var crSpan span.Span
		ctx, crSpan = k.tr.NewSpanFromContext(ctx, "kafka.consumer.commit", span.SpanKindConsumer, "")
		defer crSpan.Finish()
	}
	k.log.Notice(ctx, "committing messages", k.offsetMap)
	msgList := make([]kafka.Message, len(k.offsetMap))
	for topic, partitionMap := range k.offsetMap {
		for partition, msg := range partitionMap {
			msgList = append(msgList, kafka.Message{
				Topic:     topic,
				Partition: partition,
				Offset:    msg,
			})
		}
	}
	err := k.CommitMessages(ctx, msgList...)
	if err != nil {
		k.log.Error(ctx, "error in commit", err)
		return nil, fmt.Errorf("kafka.Reader.Commit: error committing message: %w", err)
	}
	res := k.offsetMap
	k.offsetMap = make(OffsetMap) // Reset the offset map after committing
	k.log.Notice(ctx, "messages committed", nil)
	return res, nil
}

// StoreOffset stores a offset for commit
func (k *Reader) StoreOffset(ctx context.Context, msg *kafka.Message) {
	k.commitLock.Lock()
	defer k.commitLock.Unlock()
	if k.offsetMap[msg.Topic] == nil {
		k.offsetMap[msg.Topic] = map[int]int64{}
	}
	k.offsetMap[msg.Topic][msg.Partition] = msg.Offset
}

// Close closes the Kafka reader.
func (k *Reader) Close(ctx context.Context) error {
	err := k.Reader.Close()
	if err != nil {
		k.log.Error(ctx, "error in closing reader", err)
		return fmt.Errorf("kafka.Reader.Close: error in closing reader: %w", err)
	}
	return nil
}

// StatusCheck returns the current stats of the Kafka reader.
func (k *Reader) StatusCheck(ctx context.Context) (any, error) {
	return k.Stats(), nil
}
