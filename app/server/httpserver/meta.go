package httpserver

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
)

// HealthCheck handles the HTTP request for the health check endpoint. It runs the health check and returns a 500 status code if there is an error, otherwise it returns a 204 status code.
func (h *HTTPServer) HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := h.RunHealthCheck(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Status handles the HTTP request for the status endpoint. It runs the status check and writes the JSON response.
func (h *HTTPServer) Status(w http.ResponseWriter, r *http.Request) {
	h.WriteJSON(r.Context(), w, h.RunStatusCheck(r.Context()))
}

// StatusCheck performs the status check of the server and returns a map containing the current connection count.
func (h *HTTPServer) StatusCheck(ctx context.Context) (any, error) {
	res := map[string]any{}
	res["ConnectionCount"] = h.getConnectionCount()
	return res, nil
}

// onStateChange is called when the connection state changes. It increments or decrements the connection count based on the state.
func (h *HTTPServer) onStateChange(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		h.add(1)
	case http.StateHijacked, http.StateClosed:
		h.add(-1)
	}
}

// getConnectionCount returns the current number of connections.
func (h *HTTPServer) getConnectionCount() int {
	return int(atomic.LoadInt64(&h.connectionCount))
}

// add adjusts the connection count by the specified value.
func (h *HTTPServer) add(c int64) {
	atomic.AddInt64(&h.connectionCount, c)
}
