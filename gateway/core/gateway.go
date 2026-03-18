package core

import (
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config holds all gateway configuration.
type Config struct {
	Backends     []Backend
	DefaultModel string
	MaxBodySize  int64
}

// Gateway holds runtime state: backends, health, config.
type Gateway struct {
	backends     []Backend
	health       *healthState
	defaultModel string
	maxBodySize  int64
}

// NewGateway creates a Gateway and starts the background health checker.
// If CUSTOM_BACKEND_URL is set, the custom backend is prepended and becomes the default.
func NewGateway(cfg Config) *Gateway {
	backends := cfg.Backends
	defaultModel := cfg.DefaultModel
	if custom := CustomBackendFromEnv(); custom != nil {
		backends = append([]Backend{*custom}, backends...)
		defaultModel = "custom"
	}
	g := &Gateway{
		backends:     backends,
		health:       newHealthState(),
		defaultModel: defaultModel,
		maxBodySize:  cfg.MaxBodySize,
	}
	g.startHealthChecker()
	return g
}

// resolveBackend maps a model name/alias to a backend URL and its config.
func (g *Gateway) resolveBackend(model string) (*url.URL, *Backend) {
	model = strings.TrimSpace(model)
	for i := range g.backends {
		for _, alias := range g.backends[i].Aliases {
			if strings.EqualFold(model, alias) {
				u, err := url.Parse(g.backends[i].URL)
				if err != nil {
					return nil, nil
				}
				return u, &g.backends[i]
			}
		}
	}
	return nil, nil
}

// HealthHandler returns the handler for GET /health.
func (g *Gateway) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}
}

// CatchAllHandler returns the root / 404 handler.
func (g *Gateway) CatchAllHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service":"diction-gateway","docs":"https://diction.one"}`))
	}
}

// --- Environment helpers ---

func EnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func EnvIntOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func EnvBoolOrDefault(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
