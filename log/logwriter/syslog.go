package logwriter

import (
	"context"
	"encoding/json"
	"fmt"
	dlog "log"
	"log/syslog"
	"reflect"

	"sabariram.com/goserverbase/log"
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

func (c *SyslogWriter) Start(logChannel chan log.MultipluxerLogMessage) {
	for log := range logChannel {
		_ = c.WriteMessage(log.Ctx, &log.LogMessage)
	}
}

func (c *SyslogWriter) WriteMessage(ctx context.Context, l *log.LogMessage) error {
	cr := GetCorrelationParam(ctx)
	b, err := json.Marshal(l.FullMessage)
	var fullMessage string
	if err != nil {
		fullMessage = parseErrorMsg
	} else {
		fullMessage = string(b)
	}
	c.logger.Printf("[%v] [%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp, l.LogLevelName, cr.CorrelationId, cr.ServiceName, l.ShortMessage, reflect.TypeOf(l.FullMessage), fullMessage)
	return nil
}
