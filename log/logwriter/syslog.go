package logwriter

import (
	"context"
	"fmt"
	dlog "log"
	"log/syslog"

	"github.com/sabariramc/goserverbase/log"
)

type SyslogWriter struct {
	BaseLogWriter
	logger *dlog.Logger
}

func NewSyslogWriterWriter(hostParam log.HostParams, syslogTag, prefix string) *SyslogWriter {
	logWriter, err := syslog.New(syslog.LOG_SYSLOG, syslogTag)
	if err != nil {
		panic(fmt.Errorf("Unable to set logfile: %w", err))
	}
	syslog := &SyslogWriter{logger: dlog.New(logWriter, prefix, dlog.LstdFlags), BaseLogWriter: BaseLogWriter{hostParam: &hostParam}}
	return syslog
}

func (s *SyslogWriter) GetBufferSize() int {
	return 1
}

func (c *SyslogWriter) Start(logChannel chan log.MuxLogMessage) {
	for log := range logChannel {
		_ = c.WriteMessage(log.Ctx, &log.LogMessage)
	}
}

func (c *SyslogWriter) WriteMessage(ctx context.Context, l *log.LogMessage) error {
	cr := GetCorrelationParam(ctx)
	c.logger.Printf("[%v] [%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp, l.LogLevelName, cr.CorrelationId, l.ServiceName, l.ShortMessage, l.FullMessageType, l.FullMessage)
	return nil
}
