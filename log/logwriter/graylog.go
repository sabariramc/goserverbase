package logwriter

import (
	"context"
	"fmt"
	"time"

	"github.com/sabariramc/goserverbase/v5/log"

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
	Host              string
	Version           string
	Address           string
	Port              int
	ShortMessageLimit uint
	LongMessageLimit  uint
}

// GraylogWriter represents an established graylog connection
type GraylogWriter struct {
	writer   gelf.Writer
	errorLog log.LogWriter
	c        GraylogConfig
}

func NewGraylogTCP(errorLog log.LogWriter, c GraylogConfig) (*GraylogWriter, error) {
	gelfWriter, err := gelf.NewTCPWriter(fmt.Sprintf("%s:%d", c.Address, c.Port))
	if err != nil {
		return nil, err
	}

	return &GraylogWriter{writer: gelfWriter, errorLog: errorLog, c: c}, nil
}

func NewGraylogUDP(errorLog log.LogWriter, c GraylogConfig) (*GraylogWriter, error) {
	gelfWriter, err := gelf.NewUDPWriter(fmt.Sprintf("%s:%d", c.Address, c.Port))
	if err != nil {
		return nil, err
	}
	return &GraylogWriter{writer: gelfWriter, errorLog: errorLog, c: c}, nil
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
	cr := log.ExtractCorrelationParam(ctx)
	errorMessage := log.LogMessage{
		LogLevel:  log.GetLogLevelMap(log.ERROR),
		Message:   msg.Message,
		LogObject: msg.LogObject,
		Timestamp: time.Now()}
	fullMessage := ParseLogObject(msg.LogObject, false)
	err = g.writer.WriteMessage(&gelf.Message{
		Version:  g.c.Version,
		Host:     g.c.Host,
		Short:    truncate(&msg.Message, g.c.ShortMessageLimit),
		Full:     truncate(&fullMessage, g.c.LongMessageLimit),
		TimeUnix: float64(msg.Timestamp.UnixMilli()) / 1000,
		Level:    int32(msg.Level),
		Extra: map[string]interface{}{
			"x-correlation-id":    cr.CorrelationID,
			"x-scenario-id":       cr.ScenarioID,
			"x-session-id":        cr.SessionID,
			"x-scenario-name":     cr.ScenarioName,
			"service-name":        msg.ServiceName,
			"module-name":         msg.ModuleName,
			"x-full-message-type": GetLogObjectType(msg.LogObject),
		},
	})
	if err != nil {
		_ = g.errorLog.WriteMessage(ctx, msg)
		errorMessage.Message = "Error sending to graylog"
		errorMessage.LogObject = []any{err.Error()}
		_ = g.errorLog.WriteMessage(ctx, &errorMessage)
		return fmt.Errorf("GraylogWriter.WriteMessage: %w", err)
	}
	return nil
}
