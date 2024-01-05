package api

import (
	"context"
	"fmt"
	"sync"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/segmentio/kafka-go"
)

type Writer struct {
	*kafka.Writer
	messageList []kafka.Message
	produceLock sync.Mutex
	bufferLen   int
	log         log.Logger
	msgCh       chan kafka.Message
	wg          sync.WaitGroup
	idx         int
}

var ErrWriterBufferFull = fmt.Errorf("Reader.Send: Buffer full")

func NewWriter(ctx context.Context, w *kafka.Writer, bufferLen int, log log.Logger) *Writer {
	if w.Async {
		log.Notice(ctx, "Kafak writer is set to async mode", nil)
	}
	return &Writer{
		Writer:      w,
		messageList: make([]kafka.Message, bufferLen),
		bufferLen:   bufferLen,
		log:         *log.NewResourceLogger("KafkaWriter"),
		idx:         0,
	}
}

func (w *Writer) Send(ctx context.Context, msg *kafka.Message) error {
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

func (w *Writer) Flush(ctx context.Context) error {
	w.produceLock.Lock()
	defer w.produceLock.Unlock()
	if w.idx == 0 {
		return nil
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
