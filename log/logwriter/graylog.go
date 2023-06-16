package logwriter

import (
	"context"
	"fmt"
	"time"

	"github.com/sabariramc/goserverbase/v3/log"

	"gopkg.in/Graylog2/go-gelf.v2/gelf"
)

// Transport represents a transport type enum
type Transport string

const (
	UDP Transport = "udp"
	TCP Transport = "tcp"
)

// Endpoint represents a graylog endpoint
type Endpoint struct {
	Transport Transport
	Address   string
	Port      uint
}

type GraylogConfig struct {
	Address           string
	Port              int
	ShortMessageLimit uint
	LongMessageLimit  uint
}

// GraylogWriter represents an established graylog connection
type GraylogWriter struct {
	BaseLogWriter
	writer   gelf.Writer
	errorLog log.LogWriter
	c        GraylogConfig
}

func NewGraylogTCP(hostParam log.HostParams, errorLog log.LogWriter, c GraylogConfig) (*GraylogWriter, error) {
	gelfWriter, err := gelf.NewTCPWriter(fmt.Sprintf("%s:%d", c.Address, c.Port))
	if err != nil {
		return nil, err
	}

	return &GraylogWriter{writer: gelfWriter, errorLog: errorLog, c: c, BaseLogWriter: BaseLogWriter{hostParam: &hostParam}}, nil
}

func NewGraylogUDP(hostParam log.HostParams, errorLog log.LogWriter, c GraylogConfig) (*GraylogWriter, error) {
	gelfWriter, err := gelf.NewUDPWriter(fmt.Sprintf("%s:%d", c.Address, c.Port))
	if err != nil {
		return nil, err
	}
	return &GraylogWriter{writer: gelfWriter, errorLog: errorLog, c: c, BaseLogWriter: BaseLogWriter{hostParam: &hostParam}}, nil
}

func (g *GraylogWriter) GetBufferSize() int {
	return -1
}

func (g *GraylogWriter) Start(logChannel chan log.MuxLogMessage) {
	for mxMsg := range logChannel {
		_ = g.WriteMessage(mxMsg.Ctx, &mxMsg.LogMessage)
	}
}

func truncate(s *string, l uint) string {
	if uint(len(*s)) <= l {
		return *s
	}
	return (*s)[:l]
}

func (g *GraylogWriter) WriteMessage(ctx context.Context, msg *log.LogMessage) (err error) {
	cr := log.GetCorrelationParam(ctx)
	errorMessage := log.LogMessage{
		LogLevel:     log.GetLogLevelMap(log.ERROR),
		ShortMessage: msg.ShortMessage,
		FullMessage:  msg.FullMessage,
		Timestamp:    time.Now()}
	err = g.writer.WriteMessage(&gelf.Message{
		Version:  g.hostParam.Version,
		Host:     g.hostParam.Host,
		Short:    truncate(&msg.ShortMessage, g.c.ShortMessageLimit),
		Full:     truncate(&msg.FullMessage, g.c.LongMessageLimit),
		TimeUnix: float64(msg.Timestamp.UnixMilli()) / 1000,
		Level:    int32(msg.Level),
		Extra: map[string]interface{}{
			"x-correlation-id":    cr.CorrelationId,
			"x-scenario-id":       cr.ScenarioId,
			"x-session-id":        cr.SessionId,
			"x-scenario-name":     cr.ScenarioName,
			"service-name":        msg.ServiceName,
			"module-name":         msg.ModuleName,
			"x-full-message-type": msg.FullMessageType,
		},
	})
	if err != nil {
		_ = g.errorLog.WriteMessage(ctx, msg)
		errorMessage.ShortMessage = "Error sending to graylog"
		errorMessage.FullMessage = err.Error()
		_ = g.errorLog.WriteMessage(ctx, &errorMessage)
		return fmt.Errorf("GraylogWriter.WriteMessage : %w", err)
	}
	return nil
}
