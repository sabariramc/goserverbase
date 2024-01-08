package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sabariramc/goserverbase/v4/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func (h *HTTPServer) BootstrapServer(ctx context.Context) error {
	h.server = &http.Server{Addr: h.GetPort(), Handler: h}
	return h.StartSignalMonitor(ctx)
}

func (h *HTTPServer) StartServer() {
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

func (h *HTTPServer) StartH2CServer() {
	corr := &log.CorrelationParam{CorrelationId: fmt.Sprintf("%v-HTTP2-SERVER", h.c.ServiceName)}
	ctx := log.GetContextWithCorrelation(context.TODO(), corr)
	h.log.Notice(ctx, fmt.Sprintf("Server starting at %v", h.GetPort()), nil)
	h2s := &http2.Server{}
	h.server = &http.Server{Addr: h.GetPort(), Handler: h2c.NewHandler(h, h2s)}
	err := h.StartSignalMonitor(ctx)
	if err != nil {
		h.log.Emergency(context.Background(), "Server bootstrap failed", err, nil)
	}
	err = h.server.ListenAndServe()
	time.Sleep(time.Second)
	if err != nil && err != http.ErrServerClosed {
		h.log.Emergency(context.Background(), "Server crashed", err, nil)
	}
}
