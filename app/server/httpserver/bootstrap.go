package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sabariramc/goserverbase/v5/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func (h *HTTPServer) BootstrapServer(ctx context.Context, handler http.Handler) error {
	h.server = &http.Server{Addr: h.GetPort(), Handler: handler, ConnState: h.onStateChange}
	return h.StartSignalMonitor(ctx)
}

func (h *HTTPServer) StartServer() {
	corr := &log.CorrelationParam{CorrelationID: fmt.Sprintf("%v-HTTP-SERVER", h.c.ServiceName)}
	ctx := log.GetContextWithCorrelation(context.TODO(), corr)
	h.log.Notice(ctx, fmt.Sprintf("Server starting at %v", h.GetPort()), nil)
	err := h.BootstrapServer(ctx, h)
	if err != nil {
		h.log.Emergency(context.Background(), "Server bootstrap failed", err, nil)
	}
	err = h.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		h.log.Emergency(context.Background(), "Server crashed", err, nil)
	}
	h.WaitForCompleteShutDown()
}

func (h *HTTPServer) StartTLSServer() {
	corr := &log.CorrelationParam{CorrelationID: fmt.Sprintf("%v-HTTP2-SERVER", h.c.ServiceName)}
	ctx := log.GetContextWithCorrelation(context.TODO(), corr)
	h.log.Notice(ctx, fmt.Sprintf("Server starting at %v", h.GetPort()), nil)
	err := h.BootstrapServer(ctx, h)
	if err != nil {
		h.log.Emergency(context.Background(), "Server bootstrap failed", err, nil)
	}
	err = h.server.ListenAndServeTLS(h.c.HTTP2Config.PublicKeyPath, h.c.HTTP2Config.PrivateKeyPath)
	if err != nil && err != http.ErrServerClosed {
		h.log.Emergency(context.Background(), "Server crashed", err, nil)
	}
	h.WaitForCompleteShutDown()
}

func (h *HTTPServer) StartH2CServer() {
	corr := &log.CorrelationParam{CorrelationID: fmt.Sprintf("%v-HTTP2-SERVER", h.c.ServiceName)}
	ctx := log.GetContextWithCorrelation(context.TODO(), corr)
	h.log.Notice(ctx, fmt.Sprintf("Server starting at %v", h.GetPort()), nil)
	h2s := &http2.Server{}
	err := h.BootstrapServer(ctx, h2c.NewHandler(h, h2s))
	if err != nil {
		h.log.Emergency(context.Background(), "Server bootstrap failed", err, nil)
	}
	err = h.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		h.log.Emergency(context.Background(), "Server crashed", err, nil)
	}
	h.WaitForCompleteShutDown()
}
