package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sabariramc/goserverbase/v4/log"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func (h *HTTPServer) BootstrapServer(ctx context.Context) error {
	h.server = &http.Server{Addr: h.GetPort(), Handler: h}
	return h.StartSignalMonitor(ctx)
}

func (h *HTTPServer) StartServer() {
	tracer.Start()
	defer tracer.Stop()
	corr := &log.CorrelationParam{CorrelationId: fmt.Sprintf("%v-HTTP-SERVER", h.c.ServiceName)}
	ctx := log.GetContextWithCorrelation(context.TODO(), corr)
	h.log.Notice(ctx, fmt.Sprintf("Server starting at %v", h.GetPort()), nil)
	err := h.BootstrapServer(ctx)
	if err != nil {
		h.log.Emergency(context.Background(), "Server bootstrap failed", err, nil)
	}
	err = h.server.ListenAndServe()
	time.Sleep(time.Second)
	if err != nil && err != http.ErrServerClosed {
		h.log.Emergency(context.Background(), "Server crashed", err, nil)
	}
}

func (h *HTTPServer) StartTLSServer() {
	tracer.Start()
	defer tracer.Stop()
	corr := &log.CorrelationParam{CorrelationId: fmt.Sprintf("%v-HTTP2-SERVER", h.c.ServiceName)}
	ctx := log.GetContextWithCorrelation(context.TODO(), corr)
	h.log.Notice(ctx, fmt.Sprintf("Server starting at %v", h.GetPort()), nil)
	err := h.BootstrapServer(ctx)
	if err != nil {
		h.log.Emergency(context.Background(), "Server bootstrap failed", err, nil)
	}
	err = h.server.ListenAndServeTLS(h.c.HTTP2Config.PublicKeyPath, h.c.HTTP2Config.PrivateKeyPath)
	time.Sleep(time.Second)
	if err != nil && err != http.ErrServerClosed {
		h.log.Emergency(context.Background(), "Server crashed", err, nil)
	}
}
