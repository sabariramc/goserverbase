package logwriter

import (
	"context"
	"fmt"
	stlLog "log"
	"log/syslog"

	"github.com/sabariramc/goserverbase/v6/log/message"
	"github.com/sabariramc/goserverbase/v6/trace"
)

// SyslogLogWriter writes log to syslog
type SyslogLogWriter struct {
	logger *stlLog.Logger
}

func NewSyslogWriterWriter(ctx context.Context, syslogTag, prefix string) *SyslogLogWriter {
	logWriter, err := syslog.New(syslog.LOG_SYSLOG, syslogTag)
	if err != nil {
		panic(fmt.Errorf("NewSyslogWriterWriter: unable to set log file: %w", err))
	}
	syslog := &SyslogLogWriter{logger: stlLog.New(logWriter, prefix, stlLog.LstdFlags)}
	return syslog
}

func (s *SyslogLogWriter) GetBufferSize() int {
	return 1
}

func (s *SyslogLogWriter) Start(logChannel chan message.MuxLogMessage) {
	for log := range logChannel {
		_ = s.WriteMessage(log.Ctx, &log.LogMessage)
	}
}

func (s *SyslogLogWriter) WriteMessage(ctx context.Context, l *message.LogMessage) error {
	cr := trace.ExtractCorrelationParam(ctx)
	s.logger.Printf("[%v] [%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp, l.LogLevelName, cr.CorrelationID, l.ServiceName, l.Message, GetLogObjectType(l.LogObject), l.LogObject)
	return nil
}
