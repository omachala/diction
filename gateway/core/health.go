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

func (g *Gateway) startHealthChecker() {
	check := func() {
		client := &http.Client{Timeout: 5 * time.Second}
		for _, b := range g.backends {
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
