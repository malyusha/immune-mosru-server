package status

import (
	"net/http"
	"sync/atomic"
)

type Health struct {
	http.Handler
	ready, live *uint32
}

// newHealthHandler returns new health check HTTP handler.
func newHealthHandler(ready <-chan struct{}) *Health {
	h := &Health{
		ready: new(uint32),
	}
	go h.monitorReady(ready)
	mux := http.NewServeMux()

	mux.HandleFunc("/readiness", h.handleReadiness)
	mux.HandleFunc("/liveness", h.handleLiveness)
	h.Handler = mux
	return h
}

// Ready sets status of health check handler to not ready.
func (h *Health) monitorReady(ready <-chan struct{}) {
	<-ready
	atomic.StoreUint32(h.ready, 1)
}

func (h *Health) handleReadiness(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadUint32(h.ready) == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Health) handleLiveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
