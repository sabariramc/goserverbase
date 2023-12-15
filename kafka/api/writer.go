package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v4/kafka/api/trace"
	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/segmentio/kafka-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type Writer struct {
	*kafka.Writer
	messageList     []kafka.Message
	produceLock     sync.Mutex
	bufferLen       int
	log             log.Logger
	msgCh           chan kafka.Message
	isChannelWriter bool
	wg              sync.WaitGroup
	idx             int
}

func NewWriter(ctx context.Context, w *kafka.Writer, bufferLen int, log log.Logger) *Writer {
	return &Writer{
		Writer:          w,
		messageList:     make([]kafka.Message, bufferLen),
		bufferLen:       bufferLen,
		log:             *log.NewResourceLogger("KafkaWriter"),
		isChannelWriter: false,
		idx:             0,
	}
}

func NewChanneledWriter(ctx context.Context, w *kafka.Writer, bufferLen int, log log.Logger) *Writer {
	writer := &Writer{
		Writer:          w,
		bufferLen:       bufferLen,
		log:             *log.NewResourceLogger("KafkaChanneledWriter"),
		isChannelWriter: true,
		msgCh:           make(chan kafka.Message, bufferLen),
	}
	writer.wg.Add(1)
	writer.log.Warning(ctx, "Channeled writer is an experimental implementation", nil)
	go writer.writeChannelMessage(context.Background())
	return writer
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
	msg := kafka.Message{
		Key:     []byte(key),
		Value:   message,
		Headers: messageHeader,
		Time:    time.Now(),
	}
	traceMsg := trace.NewMessageCarrier(&msg)
	tracer.Inject(span.Context(), traceMsg)
	if w.isChannelWriter {
		w.msgCh <- msg
		return nil
	}
	w.produceLock.Lock()
	w.messageList[w.idx] = msg
	w.idx++
	w.produceLock.Unlock()
	if w.idx >= w.bufferLen {
		return w.Flush(ctx)
	}
	return nil
}

func (w *Writer) Flush(ctx context.Context) error {
	if w.isChannelWriter {
		w.log.Notice(ctx, "Flush is not operational for channeled writer", nil)
		return nil
	}
	w.produceLock.Lock()
	defer w.produceLock.Unlock()
	if w.idx == 0 {
		return nil
	}
	opts := []tracer.StartSpanOption{
		tracer.Tag("messaging.kafka.topic", w.Topic),
		tracer.Tag(ext.SpanKind, ext.SpanKindInternal),
		tracer.Measured(),
	}
	span, ctx := tracer.StartSpanFromContext(ctx, "kafka.produce.flush", opts...)
	defer span.Finish()
	corr := log.GetCorrelationParam(ctx)
	span.SetTag("correlationId", corr.CorrelationId)
	w.log.Notice(ctx, "Flushing messages", w.idx)
	err := w.WriteMessages(context.Background(), w.messageList[:w.idx]...)
	w.idx = 0
	if err != nil {
		span.SetTag(ext.Error, err)
		w.log.Error(ctx, "Failed to flush message", err)
		return fmt.Errorf("Writer.Flush: error in flushing message: %w", err)
	}
	return nil
}

func (w *Writer) writeChannelMessage(ctx context.Context) {
	defer w.wg.Done()
	for msg := range w.msgCh {
		err := w.WriteMessages(ctx, msg)
		if err != nil {
			w.log.Emergency(ctx, "Failed to writing message", fmt.Errorf("kafka.Writer.writeChannelMessage: error in flushing message: %w", err), nil)
		}
	}
}

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
