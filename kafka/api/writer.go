package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/segmentio/kafka-go"
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
		return fmt.Errorf("kafka.Producer.Flush: %w", err)
	}
	return nil
}
