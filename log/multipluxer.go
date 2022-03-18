package log

import "context"

type LogWriter interface {
	Start(chan MultipluxerLogMessage)
	WriteMessage(context.Context, *LogMessage) error
}

type MultipluxerLogMessage struct {
	Ctx        context.Context
	LogMessage LogMessage
}

type LogMultipluxer struct {
	inChannel  chan MultipluxerLogMessage
	outChannel []chan MultipluxerLogMessage
}

func NewLogMultipluxer(bufferSize uint8, logWriterList ...LogWriter) *LogMultipluxer {
	outChannelList := make([]chan MultipluxerLogMessage, len(logWriterList))
	for i, logWriter := range logWriterList {
		outChannel := make(chan MultipluxerLogMessage, bufferSize)
		outChannelList[i] = outChannel
		go logWriter.Start(outChannel)
	}
	ls := &LogMultipluxer{inChannel: make(chan MultipluxerLogMessage, bufferSize), outChannel: outChannelList}
	go ls.start()
	return ls
}

func (ls *LogMultipluxer) GetChannel() chan MultipluxerLogMessage {
	return ls.inChannel
}

func (ls *LogMultipluxer) start() {
	for log := range ls.inChannel {
		for _, outChannel := range ls.outChannel {
			outChannel <- log
		}
	}
}
