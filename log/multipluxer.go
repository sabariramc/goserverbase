package log

import "context"

type ChanneledLogWriter interface {
	Start(chan MuxLogMessage)
	WriteMessage(context.Context, *LogMessage) error
	GetBufferSize() int
}

type LogWriter interface {
	WriteMessage(context.Context, *LogMessage) error
}

type LogMux interface {
	Print(context.Context, *LogMessage)
}

type MuxLogMessage struct {
	Ctx        context.Context
	LogMessage LogMessage
}

type ChanneledLogMux struct {
	inChannel  chan MuxLogMessage
	outChannel []chan MuxLogMessage
}

func NewChanneledLogMux(bufferSize uint8, logWriterList ...ChanneledLogWriter) *ChanneledLogMux {
	outChannelList := make([]chan MuxLogMessage, len(logWriterList))
	for i, logWriter := range logWriterList {
		lBufferSize := logWriter.GetBufferSize()
		if lBufferSize < 1 {
			lBufferSize = int(bufferSize)
		}
		outChannel := make(chan MuxLogMessage, lBufferSize)
		outChannelList[i] = outChannel
		go logWriter.Start(outChannel)
	}
	ls := &ChanneledLogMux{inChannel: make(chan MuxLogMessage, bufferSize), outChannel: outChannelList}
	go ls.start()
	return ls
}

func (ls *ChanneledLogMux) Print(ctx context.Context, msg *LogMessage) {
	ls.inChannel <- MuxLogMessage{
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

type DefaultLogMux struct {
	writer []LogWriter
}

func NewDefaultLogMux(logWriterList ...LogWriter) *DefaultLogMux {
	ls := &DefaultLogMux{writer: logWriterList}
	return ls
}

func (ls *DefaultLogMux) Print(ctx context.Context, msg *LogMessage) {
	for _, w := range ls.writer {
		_ = w.WriteMessage(ctx, msg)
	}
}
