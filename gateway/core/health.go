package core

import (
	"net/http"
	"sync"
	"time"
)

type healthState struct {
	mu        sync.RWMutex
	available map[string]bool
}

func newHealthState() *healthState {
	return &healthState{
		available: make(map[string]bool),
	}
}

func (h *healthState) set(name string, ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.available[name] = ok
}

func (h *healthState) get(name string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.available[name]
}

// SetServiceHealth sets the health status of a named service. This is the
// public API for external code (e.g. private gateway wiring) to register
// additional services in the health state alongside the built-in backends.
func (g *Gateway) SetServiceHealth(name string, ok bool) {
	g.health.set(name, ok)
}

// IsServiceHealthy returns whether a named service is currently healthy.
func (g *Gateway) IsServiceHealthy(name string) bool {
	return g.health.get(name)
}

func (g *Gateway) startHealthChecker() {
	check := func() {
		client := &http.Client{Timeout: 5 * time.Second}
		for _, b := range g.backends {
			if b.Disabled {
				continue // not deployed — don't waste a request on a dead host
			}
			resp, err := client.Get(b.URL + "/health")
			ok := err == nil && resp != nil && resp.StatusCode == http.StatusOK
			if resp != nil {
				resp.Body.Close()
			}
			g.health.set(b.Name, ok)
		}
	}
	check()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			check()
		}
	}()
}
