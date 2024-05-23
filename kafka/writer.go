// Package kafka wraps [github.com/segmentio/kafka-go] API and adds additional functionalities such as tracing and buffering.
package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/segmentio/kafka-go"
)

// ProduceTracer is an interface for injecting tracing information into Kafka messages.
type ProduceTracer interface {
	KafkaInject(ctx context.Context, msg *kafka.Message)
	span.SpanOp
}

// Writer extends [kafka.Writer] with batch writing, tracing, and StatusCheck hook.
type Writer struct {
	*kafka.Writer
	messageList []kafka.Message // Storage for batch messages
	produceLock sync.Mutex
	bufferLen   int // Max size of the batch
	log         log.Log
	msgCh       chan kafka.Message
	wg          sync.WaitGroup
	idx         int
	tr          ProduceTracer
}

// ErrWriterBufferFull is an error indicating that the writer buffer is full.
var ErrWriterBufferFull = fmt.Errorf("Reader.Send: Buffer full")

// NewWriter creates a new Writer.
func NewWriter(ctx context.Context, w *kafka.Writer, bufferLen int, log log.Log, tr ProduceTracer) *Writer {
	if w.Async {
		log.Notice(ctx, "Kafka writer is set to async mode", nil)
	}
	return &Writer{
		Writer:      w,
		messageList: make([]kafka.Message, bufferLen),
		bufferLen:   bufferLen,
		log:         log.NewResourceLogger("KafkaWriter"),
		idx:         0,
		tr:          tr,
	}
}

// Send writes the message to the broker in async mode or batch mode.
func (w *Writer) Send(ctx context.Context, msg *kafka.Message) error {
	if w.tr != nil {
		var crSpan span.Span
		ctx, crSpan = w.tr.NewSpanFromContext(ctx, "kafka.produce", span.SpanKindProducer, msg.Topic)
		crSpan.SetAttribute("messaging.kafka.topic", msg.Topic)
		crSpan.SetAttribute("messaging.kafka.key", string(msg.Key))
		crSpan.SetAttribute("messaging.kafka.timestamp", msg.Time)
		defer crSpan.Finish()
		w.tr.KafkaInject(ctx, msg)
	}
	if w.Async {
		return w.WriteMessages(ctx, *msg)
	}
	w.produceLock.Lock()
	defer w.produceLock.Unlock()
	if w.idx >= w.bufferLen {
		return ErrWriterBufferFull
	}
	w.messageList[w.idx] = *msg
	w.idx++
	return nil
}

// Flush writes the message batch to the broker.
func (w *Writer) Flush(ctx context.Context) error {
	w.produceLock.Lock()
	defer w.produceLock.Unlock()
	if w.idx == 0 {
		return nil
	}
	if w.tr != nil {
		var crSpan span.Span
		ctx, crSpan = w.tr.NewSpanFromContext(ctx, "kafka.producer.flush", span.SpanKindProducer, "")
		defer crSpan.Finish()
	}
	w.log.Notice(ctx, "Flushing messages", w.idx)
	err := w.WriteMessages(context.Background(), w.messageList[:w.idx]...)
	w.idx = 0
	if err != nil {
		w.log.Error(ctx, "Failed to flush message", err)
		return fmt.Errorf("Writer.Flush: error in flushing message: %w", err)
	}
	return nil
}

// Close closes the Writer, ensuring all messages are flushed.
func (w *Writer) Close(ctx context.Context) error {
	if w.msgCh != nil {
		close(w.msgCh)
	}
	w.wg.Wait()
	err := w.Writer.Close()
	if err != nil {
		w.log.Error(ctx, "Error in closing writer", err)
		return fmt.Errorf("kafka.Writer.Close: error in closing writer: %w", err)
	}
	return nil
}

// StatusCheck returns the current status of the Writer.
func (w *Writer) StatusCheck(ctx context.Context) (any, error) {
	return w.Stats(), nil
}
