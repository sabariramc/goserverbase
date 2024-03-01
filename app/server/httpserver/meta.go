package httpserver

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
)

func (h *HTTPServer) HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := h.RunHealthCheck(r.Context())
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(204)
}

func (h *HTTPServer) Status(w http.ResponseWriter, r *http.Request) {
	h.WriteJSON(r.Context(), w, h.RunStatusCheck(r.Context()))
}

func (h *HTTPServer) StatusCheck(ctx context.Context) (any, error) {
	res := map[string]any{}
	res["ConnectionCount"] = h.getConnectionCount()
	return res, nil
}

func (h *HTTPServer) onStateChange(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		h.add(1)
	case http.StateHijacked, http.StateClosed:
		h.add(-1)
	}
}

func (h *HTTPServer) getConnectionCount() int {
	return int(atomic.LoadInt64(&h.connectionCount))
}

func (h *HTTPServer) add(c int64) {
	atomic.AddInt64(&h.connectionCount, c)
}
