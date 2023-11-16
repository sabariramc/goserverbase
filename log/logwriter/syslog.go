package logwriter

import (
	"context"
	"fmt"
	stlLog "log"
	"log/syslog"

	"github.com/sabariramc/goserverbase/v4/log"
)

type SyslogWriter struct {
	BaseLogWriter
	logger *stlLog.Logger
}

func NewSyslogWriterWriter(hostParam log.HostParams, syslogTag, prefix string) *SyslogWriter {
	logWriter, err := syslog.New(syslog.LOG_SYSLOG, syslogTag)
	if err != nil {
		panic(fmt.Errorf("unable to set log file: %w", err))
	}
	syslog := &SyslogWriter{logger: stlLog.New(logWriter, prefix, stlLog.LstdFlags), BaseLogWriter: BaseLogWriter{hostParam: &hostParam}}
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
	cr := log.GetCorrelationParam(ctx)
	c.logger.Printf("[%v] [%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp, l.LogLevelName, cr.CorrelationId, l.ServiceName, l.Message, GetLogObjectType(l.LogObject), l.LogObject)
	return nil
}
