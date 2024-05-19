package log

import (
	"context"

	m "github.com/sabariramc/goserverbase/v6/log/message"
)

type ChanneledLogWriter interface {
	Start(chan m.MuxLog)
	WriteMessage(context.Context, *m.Log) error
	GetBufferSize() int
}

type LogWriter interface {
	WriteMessage(context.Context, *m.Log) error
}

// Mux interface abstracts how the logger interacts with the the log handlers
type Mux interface {
	Print(context.Context, *m.Log)
	AddLogWriter(context.Context, LogWriter)
}

// DefaultLogMux is a implementation of LogMux and calls the associated log handlers sequentially over a for loop
type DefaultLogMux struct {
	writer []LogWriter
}

func NewDefaultLogMux(logWriterList ...LogWriter) *DefaultLogMux {
	ls := &DefaultLogMux{writer: logWriterList}
	return ls
}

func (ls *DefaultLogMux) Print(ctx context.Context, msg *m.Log) {
	for _, w := range ls.writer {
		_ = w.WriteMessage(ctx, msg)
	}
}

func (ls *DefaultLogMux) AddLogWriter(ctx context.Context, writer LogWriter) {
	ls.writer = append(ls.writer, writer)
}

type ChanneledLogMux struct {
	inChannel  chan m.MuxLog
	outChannel []chan m.MuxLog
}

func NewChanneledLogMux(bufferSize uint8, logWriterList ...ChanneledLogWriter) *ChanneledLogMux {
	outChannelList := make([]chan m.MuxLog, len(logWriterList))
	for i, logWriter := range logWriterList {
		lBufferSize := logWriter.GetBufferSize()
		if lBufferSize < 1 {
			lBufferSize = int(bufferSize)
		}
		outChannel := make(chan m.MuxLog, lBufferSize)
		outChannelList[i] = outChannel
		go logWriter.Start(outChannel)
	}
	ls := &ChanneledLogMux{inChannel: make(chan m.MuxLog, bufferSize), outChannel: outChannelList}
	go ls.start()
	return ls
}

func (ls *ChanneledLogMux) Print(ctx context.Context, msg *m.Log) {
	ls.inChannel <- m.MuxLog{
		Ctx: ctx,
		Log: *msg,
	}
}

func (ls *ChanneledLogMux) start() {
	for log := range ls.inChannel {
		for _, outChannel := range ls.outChannel {
			outChannel <- log
		}
	}
}
