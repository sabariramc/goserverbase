package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/segmentio/kafka-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type Writer struct {
	*kafka.Writer
	messageList []kafka.Message
	produceLock sync.Mutex
	bufferLen   int
	log         log.Logger
}

func NewWriter(ctx context.Context, w *kafka.Writer, bufferLen int, log log.Logger) *Writer {
	return &Writer{
		Writer:      w,
		messageList: make([]kafka.Message, 0, bufferLen),
		bufferLen:   bufferLen,
		log:         log,
	}
}

func (w *Writer) Send(ctx context.Context, key string, message []byte, messageHeader []kafka.Header) error {
	opts := []tracer.StartSpanOption{
		tracer.Tag("messaging.kafka.topic", w.Topic),
		tracer.Tag("messaging.kafka.key", key),
		tracer.Tag("messaging.kafka.timestamp", time.Now().UnixMilli()),
		tracer.Tag(ext.SpanKind, ext.SpanKindProducer),
		tracer.Tag(ext.MessagingSystem, "kafka"),
		tracer.Measured(),
	}
	span, ctx := tracer.StartSpanFromContext(ctx, "kafka.produce", opts...)
	defer span.Finish()
	w.produceLock.Lock()
	w.messageList = append(w.messageList, kafka.Message{
		Key:     []byte(key),
		Value:   message,
		Headers: messageHeader,
		Time:    time.Now(),
	})
	w.produceLock.Unlock()
	if len(w.messageList) >= w.bufferLen {
		return w.Flush(ctx)
	}
	return nil
}

func (w *Writer) Flush(ctx context.Context) error {
	opts := []tracer.StartSpanOption{
		tracer.Tag("messaging.kafka.topic", w.Topic),
		tracer.Tag(ext.SpanKind, ext.SpanKindInternal),
		tracer.Measured(),
	}
	span, ctx := tracer.StartSpanFromContext(ctx, "kafka.flush", opts...)
	defer span.Finish()
	w.produceLock.Lock()
	defer w.produceLock.Unlock()
	if len(w.messageList) == 0 {
		return nil
	}
	w.log.Debug(ctx, "Flushing messages", nil)
	err := w.WriteMessages(context.Background(), w.messageList...)
	w.messageList = make([]kafka.Message, 0, w.bufferLen)
	if err != nil {
		span.SetTag(ext.Error, err)
		w.log.Error(ctx, "Failed to flush message", err)
		return fmt.Errorf("kafka.Producer.Flush: %w", err)
	}
	return nil
}
