package logwriter

import (
	"context"
	"fmt"
	stlLog "log"
	"log/syslog"

	"github.com/sabariramc/goserverbase/v5/log"
)

type SyslogWriter struct {
	logger *stlLog.Logger
}

func NewSyslogWriterWriter(ctx context.Context, syslogTag, prefix string) *SyslogWriter {
	logWriter, err := syslog.New(syslog.LOG_SYSLOG, syslogTag)
	if err != nil {
		panic(fmt.Errorf("NewSyslogWriterWriter: unable to set log file: %w", err))
	}
	syslog := &SyslogWriter{logger: stlLog.New(logWriter, prefix, stlLog.LstdFlags)}
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
	cr := log.ExtractCorrelationParam(ctx)
	c.logger.Printf("[%v] [%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp, l.LogLevelName, cr.CorrelationID, l.ServiceName, l.Message, GetLogObjectType(l.LogObject), l.LogObject)
	return nil
}
