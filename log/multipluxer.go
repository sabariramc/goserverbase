package log

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/log/logwriter"
	m "github.com/sabariramc/goserverbase/v6/log/message"
)

// Mux interface abstracts how the logger interacts with the the log handlers
type Mux interface {
	Print(context.Context, *m.LogMessage)
	AddLogWriter(context.Context, logwriter.LogWriter)
}

// DefaultLogMux is a implementation of LogMux and calls the associated log handlers sequentially over a for loop
type DefaultLogMux struct {
	writer []logwriter.LogWriter
}

func NewDefaultLogMux(logWriterList ...logwriter.LogWriter) *DefaultLogMux {
	ls := &DefaultLogMux{writer: logWriterList}
	return ls
}

func (ls *DefaultLogMux) Print(ctx context.Context, msg *m.LogMessage) {
	for _, w := range ls.writer {
		_ = w.WriteMessage(ctx, msg)
	}
}

func (ls *DefaultLogMux) AddLogWriter(ctx context.Context, writer logwriter.LogWriter) {
	ls.writer = append(ls.writer, writer)
}

type ChanneledLogMux struct {
	inChannel  chan m.MuxLogMessage
	outChannel []chan m.MuxLogMessage
}

func NewChanneledLogMux(bufferSize uint8, logWriterList ...logwriter.ChanneledLogWriter) *ChanneledLogMux {
	outChannelList := make([]chan m.MuxLogMessage, len(logWriterList))
	for i, logWriter := range logWriterList {
		lBufferSize := logWriter.GetBufferSize()
		if lBufferSize < 1 {
			lBufferSize = int(bufferSize)
		}
		outChannel := make(chan m.MuxLogMessage, lBufferSize)
		outChannelList[i] = outChannel
		go logWriter.Start(outChannel)
	}
	ls := &ChanneledLogMux{inChannel: make(chan m.MuxLogMessage, bufferSize), outChannel: outChannelList}
	go ls.start()
	return ls
}

func (ls *ChanneledLogMux) Print(ctx context.Context, msg *m.LogMessage) {
	ls.inChannel <- m.MuxLogMessage{
		Ctx:        ctx,
		LogMessage: *msg,
	}
}

func (ls *ChanneledLogMux) start() {
	for log := range ls.inChannel {
		for _, outChannel := range ls.outChannel {
			outChannel <- log
		}
	}
}
