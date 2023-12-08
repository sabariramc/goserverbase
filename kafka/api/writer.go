package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/segmentio/kafka-go"
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
}

func NewWriter(ctx context.Context, w *kafka.Writer, bufferLen int, log log.Logger) *Writer {
	return &Writer{
		Writer:          w,
		messageList:     make([]kafka.Message, 0, bufferLen),
		bufferLen:       bufferLen,
		log:             *log.NewResourceLogger("KafkaWriter"),
		isChannelWriter: false,
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
	msg := kafka.Message{
		Key:     []byte(key),
		Value:   message,
		Headers: messageHeader,
		Time:    time.Now(),
	}
	if w.isChannelWriter {
		w.msgCh <- msg
		return nil
	}
	w.produceLock.Lock()
	w.messageList = append(w.messageList, msg)
	w.produceLock.Unlock()
	if len(w.messageList) >= w.bufferLen {
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
	if len(w.messageList) == 0 {
		return nil
	}
	w.log.Debug(ctx, "Flushing messages", nil)
	err := w.WriteMessages(context.Background(), w.messageList...)
	w.messageList = make([]kafka.Message, 0, w.bufferLen)
	if err != nil {
		w.log.Error(ctx, "Failed to flush message", err)
		return fmt.Errorf("kafka.Writer.Flush: error in flushing message: %w", err)
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
