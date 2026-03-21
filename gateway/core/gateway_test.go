package core

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// --- resolveBackend ---

func testGateway() *Gateway {
	return &Gateway{
		backends: []Backend{
			{Name: "small", URL: "http://whisper-small:8000", Aliases: []string{"small", "Systran/faster-whisper-small"}},
			{Name: "medium", URL: "http://whisper-medium:8000", Aliases: []string{"medium", "Systran/faster-whisper-medium"}},
			{Name: "parakeet-v3", URL: "http://parakeet:5092", Aliases: []string{"parakeet-v3", "parakeet"}},
		},
		health:       newHealthState(),
		defaultModel: "small",
		maxBodySize:  10 * 1024 * 1024,
	}
}

func TestResolveBackend_ExactMatch(t *testing.T) {
	g := testGateway()
	u, b := g.resolveBackend("small")
	if u == nil {
		t.Fatal("expected non-nil URL for 'small'")
	}
	if b.Name != "small" {
		t.Errorf("backend name: want small, got %s", b.Name)
	}
	if u.Host != "whisper-small:8000" {
		t.Errorf("host: want whisper-small:8000, got %s", u.Host)
	}
}

func TestResolveBackend_AliasMatch(t *testing.T) {
	g := testGateway()
	u, b := g.resolveBackend("Systran/faster-whisper-medium")
	if u == nil {
		t.Fatal("expected non-nil URL for alias")
	}
	if b.Name != "medium" {
		t.Errorf("backend name: want medium, got %s", b.Name)
	}
}

func TestResolveBackend_CaseInsensitive(t *testing.T) {
	g := testGateway()
	u, _ := g.resolveBackend("SMALL")
	if u == nil {
		t.Fatal("expected non-nil URL for uppercase alias")
	}
}

func TestResolveBackend_Whitespace(t *testing.T) {
	g := testGateway()
	u, _ := g.resolveBackend("  small  ")
	if u == nil {
		t.Fatal("expected non-nil URL with surrounding whitespace")
	}
}

func TestResolveBackend_Unknown(t *testing.T) {
	g := testGateway()
	u, b := g.resolveBackend("gpt-4o-transcribe")
	if u != nil {
		t.Error("expected nil URL for unknown model")
	}
	if b != nil {
		t.Error("expected nil backend for unknown model")
	}
}

func TestResolveBackend_InvalidURL(t *testing.T) {
	// url.Parse fails on a control-character URL → resolveBackend returns nil, nil.
	g := &Gateway{
		backends:     []Backend{{Name: "bad", URL: "://\x00invalid", Aliases: []string{"bad"}}},
		health:       newHealthState(),
		defaultModel: "bad",
		maxBodySize:  10 * 1024 * 1024,
	}
	u, b := g.resolveBackend("bad")
	if u != nil || b != nil {
		t.Error("expected nil, nil for invalid backend URL")
	}
}

func TestResolveBackend_Empty(t *testing.T) {
	g := testGateway()
	u, b := g.resolveBackend("")
	if u != nil || b != nil {
		t.Error("expected nil for empty model string")
	}
}

// --- Env helpers ---

func TestEnvOrDefault_Set(t *testing.T) {
	os.Setenv("TEST_KEY", "myvalue")
	defer os.Unsetenv("TEST_KEY")
	if got := EnvOrDefault("TEST_KEY", "fallback"); got != "myvalue" {
		t.Errorf("want myvalue, got %s", got)
	}
}

func TestEnvOrDefault_Unset(t *testing.T) {
	os.Unsetenv("TEST_KEY_MISSING")
	if got := EnvOrDefault("TEST_KEY_MISSING", "fallback"); got != "fallback" {
		t.Errorf("want fallback, got %s", got)
	}
}

func TestEnvIntOrDefault_Set(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")
	if got := EnvIntOrDefault("TEST_INT", 0); got != 42 {
		t.Errorf("want 42, got %d", got)
	}
}

func TestEnvIntOrDefault_Invalid(t *testing.T) {
	os.Setenv("TEST_INT_BAD", "notanumber")
	defer os.Unsetenv("TEST_INT_BAD")
	if got := EnvIntOrDefault("TEST_INT_BAD", 99); got != 99 {
		t.Errorf("want fallback 99, got %d", got)
	}
}

func TestEnvIntOrDefault_Unset(t *testing.T) {
	os.Unsetenv("TEST_INT_MISSING")
	if got := EnvIntOrDefault("TEST_INT_MISSING", 7); got != 7 {
		t.Errorf("want 7, got %d", got)
	}
}

func TestEnvBoolOrDefault_True(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")
	if got := EnvBoolOrDefault("TEST_BOOL", false); !got {
		t.Error("want true, got false")
	}
}

func TestEnvBoolOrDefault_False(t *testing.T) {
	os.Setenv("TEST_BOOL", "false")
	defer os.Unsetenv("TEST_BOOL")
	if got := EnvBoolOrDefault("TEST_BOOL", true); got {
		t.Error("want false, got true")
	}
}

func TestEnvBoolOrDefault_Unset(t *testing.T) {
	os.Unsetenv("TEST_BOOL_MISSING")
	if got := EnvBoolOrDefault("TEST_BOOL_MISSING", true); !got {
		t.Error("want fallback true, got false")
	}
}

// --- HealthHandler ---

func TestHealthHandler(t *testing.T) {
	g := testGateway()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	g.HealthHandler()(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: want application/json, got %s", ct)
	}
	if body := rr.Body.String(); body != `{"status":"ok"}` {
		t.Errorf("body: want {\"status\":\"ok\"}, got %s", body)
	}
}

// --- CatchAllHandler ---

func TestCatchAllHandler_Root(t *testing.T) {
	g := testGateway()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	g.CatchAllHandler()(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: want application/json, got %s", ct)
	}
	body := rr.Body.String()
	if body == "" {
		t.Error("expected non-empty body for /")
	}
}

func TestCatchAllHandler_Unknown(t *testing.T) {
	g := testGateway()
	req := httptest.NewRequest(http.MethodGet, "/unknown/path", nil)
	rr := httptest.NewRecorder()
	g.CatchAllHandler()(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status: want 404, got %d", rr.Code)
	}
}

// --- CustomBackendFromEnv ---

func TestCustomBackendFromEnv_NotSet(t *testing.T) {
	os.Unsetenv("CUSTOM_BACKEND_URL")
	if b := CustomBackendFromEnv(); b != nil {
		t.Error("expected nil when CUSTOM_BACKEND_URL is not set")
	}
}

func TestCustomBackendFromEnv_WithURL(t *testing.T) {
	os.Setenv("CUSTOM_BACKEND_URL", "http://192.168.1.50:8000")
	os.Setenv("CUSTOM_BACKEND_MODEL", "faster-whisper-large-v3")
	os.Setenv("CUSTOM_BACKEND_NEEDS_WAV", "true")
	os.Setenv("CUSTOM_BACKEND_AUTH", "Bearer sk-test")
	defer func() {
		os.Unsetenv("CUSTOM_BACKEND_URL")
		os.Unsetenv("CUSTOM_BACKEND_MODEL")
		os.Unsetenv("CUSTOM_BACKEND_NEEDS_WAV")
		os.Unsetenv("CUSTOM_BACKEND_AUTH")
	}()

	b := CustomBackendFromEnv()
	if b == nil {
		t.Fatal("expected non-nil backend")
	}
	if b.Name != "custom" {
		t.Errorf("name: want custom, got %s", b.Name)
	}
	if b.URL != "http://192.168.1.50:8000" {
		t.Errorf("url: want http://192.168.1.50:8000, got %s", b.URL)
	}
	if b.ForwardModel != "faster-whisper-large-v3" {
		t.Errorf("forward model: want faster-whisper-large-v3, got %s", b.ForwardModel)
	}
	if !b.NeedsWAV {
		t.Error("expected NeedsWAV=true")
	}
	if b.AuthHeader != "Bearer sk-test" {
		t.Errorf("auth: want 'Bearer sk-test', got %s", b.AuthHeader)
	}
	if !b.SkipHealthCheck {
		t.Error("expected SkipHealthCheck=true for custom backend")
	}
}

// --- NewGateway with custom backend ---

func TestNewGateway_CustomBackendOverridesDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	os.Setenv("CUSTOM_BACKEND_URL", srv.URL)
	defer os.Unsetenv("CUSTOM_BACKEND_URL")

	g := NewGateway(Config{
		Backends: []Backend{
			{Name: "small", URL: srv.URL, Aliases: []string{"small"}},
		},
		DefaultModel: "small",
		MaxBodySize:  1 << 20,
	})

	// Custom backend should be prepended and become the default
	u, b := g.resolveBackend("custom")
	if u == nil || b == nil {
		t.Fatal("expected custom backend to be resolvable")
	}
	if g.defaultModel != "custom" {
		t.Errorf("defaultModel: want custom, got %s", g.defaultModel)
	}
}

// --- DefaultBackends ---

func TestDefaultBackends_NonEmpty(t *testing.T) {
	backends := DefaultBackends()
	if len(backends) == 0 {
		t.Fatal("expected at least one backend")
	}
	for _, b := range backends {
		if b.Name == "" {
			t.Error("backend has empty Name")
		}
		if b.URL == "" {
			t.Errorf("backend %s has empty URL", b.Name)
		}
		if len(b.Aliases) == 0 {
			t.Errorf("backend %s has no aliases", b.Name)
		}
	}
}

// --- NewGateway ---

func TestNewGateway_CreatesWithBackends(t *testing.T) {
	// Use a real httptest server so the initial health check doesn't hang.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	g := NewGateway(Config{
		Backends: []Backend{
			{Name: "test", URL: srv.URL, Aliases: []string{"test"}},
		},
		DefaultModel: "test",
		MaxBodySize:  1 << 20,
	})

	if g == nil {
		t.Fatal("expected non-nil gateway")
	}
	u, b := g.resolveBackend("test")
	if u == nil || b == nil {
		t.Error("expected backend to be resolvable after NewGateway")
	}
}
