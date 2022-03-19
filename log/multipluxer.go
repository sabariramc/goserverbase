package log

import "context"

type ChanneledLogWriter interface {
	Start(chan MultipluxerLogMessage)
	WriteMessage(context.Context, *LogMessage) error
	GetBufferSize() int
}

type LogWriter interface {
	WriteMessage(context.Context, *LogMessage) error
}

type AuditLogWriter interface {
	WriteAuditMessage(context.Context, *AuditLogMessage) error
}

type LogMultipluxer interface {
	Print(context.Context, *LogMessage)
}

type MultipluxerLogMessage struct {
	Ctx        context.Context
	LogMessage LogMessage
}

type ChanneledLogMultipluxer struct {
	inChannel  chan MultipluxerLogMessage
	outChannel []chan MultipluxerLogMessage
}

func NewChanneledLogMultipluxer(bufferSize uint8, logWriterList ...ChanneledLogWriter) *ChanneledLogMultipluxer {
	outChannelList := make([]chan MultipluxerLogMessage, len(logWriterList))
	for i, logWriter := range logWriterList {
		lbufferSize := logWriter.GetBufferSize()
		if lbufferSize < 1 {
			lbufferSize = int(bufferSize)
		}
		outChannel := make(chan MultipluxerLogMessage, lbufferSize)
		outChannelList[i] = outChannel
		go logWriter.Start(outChannel)
	}
	ls := &ChanneledLogMultipluxer{inChannel: make(chan MultipluxerLogMessage, bufferSize), outChannel: outChannelList}
	go ls.start()
	return ls
}

func (ls *ChanneledLogMultipluxer) Print(ctx context.Context, msg *LogMessage) {
	ls.inChannel <- MultipluxerLogMessage{
		Ctx:        ctx,
		LogMessage: *msg,
	}
}

func (ls *ChanneledLogMultipluxer) start() {
	for log := range ls.inChannel {
		for _, outChannel := range ls.outChannel {
			outChannel <- log
		}
	}
}

type SequenctialLogMultipluxer struct {
	writer []LogWriter
}

func NewSequenctialLogMultipluxer(logWriterList ...LogWriter) *SequenctialLogMultipluxer {
	ls := &SequenctialLogMultipluxer{writer: logWriterList}
	return ls
}

func (ls *SequenctialLogMultipluxer) Print(ctx context.Context, msg *LogMessage) {
	for _, w := range ls.writer {
		_ = w.WriteMessage(ctx, msg)
	}
}
