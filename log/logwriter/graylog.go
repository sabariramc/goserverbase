package logwriter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"sabariram.com/goserverbase/log"

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

// GraylogWriter represents an established graylog connection
type GraylogWriter struct {
	BaseLogWriter
	writer    gelf.Writer
	backuplog log.LogWriter
}

func NewGraylogTCP(hostParam log.HostParams, backuplog log.LogWriter, e Endpoint) (*GraylogWriter, error) {
	gelfWriter, err := gelf.NewTCPWriter(fmt.Sprintf("%s:%d", e.Address, e.Port))
	if err != nil {
		return nil, err
	}

	return &GraylogWriter{writer: gelfWriter, backuplog: backuplog, BaseLogWriter: BaseLogWriter{hostParam: &hostParam}}, nil
}

func NewGraylogUDP(hostParam log.HostParams, backuplog log.LogWriter, e Endpoint) (*GraylogWriter, error) {
	gelfWriter, err := gelf.NewUDPWriter(fmt.Sprintf("%s:%d", e.Address, e.Port))
	if err != nil {
		return nil, err
	}
	return &GraylogWriter{writer: gelfWriter, backuplog: backuplog, BaseLogWriter: BaseLogWriter{hostParam: &hostParam}}, nil
}

func (g *GraylogWriter) Start(logChannel chan log.MultipluxerLogMessage) {
	for mxMsg := range logChannel {
		msg := mxMsg.LogMessage
		errorMessage := log.LogMessage{
			LogLevelMap:  log.GetLogLevelMap(log.ERROR),
			ShortMessage: msg.ShortMessage,
			Timestamp:    time.Now()}
		err := g.WriteMessage(mxMsg.Ctx, &msg)
		if err != nil {
			_ = g.backuplog.WriteMessage(mxMsg.Ctx, &msg)
			errorMessage.ShortMessage = "Error sending to graylog"
			errorMessage.FullMessage = err
			_ = g.backuplog.WriteMessage(mxMsg.Ctx, &errorMessage)
		}
	}
}

func (g *GraylogWriter) WriteMessage(ctx context.Context, msg *log.LogMessage) (err error) {
	cr := GetCorrelationParam(ctx)
	blob, _ := json.Marshal(msg.FullMessage)
	fullMessage := string(blob)
	return g.writer.WriteMessage(&gelf.Message{
		Version:  g.hostParam.Version,
		Host:     g.hostParam.Host,
		Short:    msg.ShortMessage,
		Full:     fullMessage,
		TimeUnix: float64(msg.Timestamp.UnixMilli()) / 1000,
		Level:    int32(msg.Level),
		Extra: map[string]interface{}{
			"x-correlation-id": cr.CorrelationId,
			"x-scenario-id":    cr.ScenarioId,
			"x-session-id":     cr.SessionId,
			"x-scenario-name":  cr.ScenarioName,
			"service-name":     cr.ServiceName,
		},
	})
}
