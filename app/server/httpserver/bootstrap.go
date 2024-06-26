package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sabariramc/goserverbase/v6/correlation"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// BootstrapServer initializes the HTTP server with the given handler and starts monitoring for shutdown signals.
func (h *HTTPServer) BootstrapServer(ctx context.Context, handler http.Handler) error {
	h.server = &http.Server{Addr: h.GetPort(), Handler: handler, ConnState: h.onStateChange}
	return h.StartSignalMonitor(ctx)
}

// StartServer starts the HTTP server and listens for incoming requests. It logs the startup and handles any server errors.
func (h *HTTPServer) StartServer() {
	corr := &correlation.CorrelationParam{CorrelationID: fmt.Sprintf("%v-HTTP-SERVER", h.c.ServiceName)}
	ctx := correlation.GetContextWithCorrelationParam(context.TODO(), corr)
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

// StartTLSServer starts the HTTPS server using TLS and listens for incoming requests. It logs the startup and handles any server errors.
func (h *HTTPServer) StartTLSServer() {
	corr := &correlation.CorrelationParam{CorrelationID: fmt.Sprintf("%v-HTTP2-SERVER", h.c.ServiceName)}
	ctx := correlation.GetContextWithCorrelationParam(context.TODO(), corr)
	h.log.Notice(ctx, fmt.Sprintf("Server starting at %v", h.GetPort()), nil)
	err := h.BootstrapServer(ctx, h)
	if err != nil {
		h.log.Emergency(context.Background(), "Server bootstrap failed", err, nil)
	}
	err = h.server.ListenAndServeTLS(h.c.TLSConfig.PublicKeyPath, h.c.TLSConfig.PrivateKeyPath)
	if err != nil && err != http.ErrServerClosed {
		h.log.Emergency(context.Background(), "Server crashed", err, nil)
	}
	h.WaitForCompleteShutDown()
}

// StartH2CServer starts the HTTP/2 server in cleartext mode (h2c) and listens for incoming requests. It logs the startup and handles any server errors.
func (h *HTTPServer) StartH2CServer() {
	corr := &correlation.CorrelationParam{CorrelationID: fmt.Sprintf("%v-HTTP2-SERVER", h.c.ServiceName)}
	ctx := correlation.GetContextWithCorrelationParam(context.TODO(), corr)
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
