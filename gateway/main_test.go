package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildMux_DefaultPort(t *testing.T) {
	t.Setenv("GATEWAY_PORT", "")
	_, port := buildMux()
	if port != "8080" {
		t.Errorf("default port: want 8080, got %s", port)
	}
}

func TestBuildMux_CustomPort(t *testing.T) {
	t.Setenv("GATEWAY_PORT", "9999")
	_, port := buildMux()
	if port != "9999" {
		t.Errorf("custom port: want 9999, got %s", port)
	}
}

func TestBuildMux_CustomDefaultModel(t *testing.T) {
	t.Setenv("DEFAULT_MODEL", "medium")
	mux, _ := buildMux()
	if mux == nil {
		t.Fatal("expected non-nil mux")
	}
}

func TestBuildMux_RoutesRegistered(t *testing.T) {
	t.Setenv("GATEWAY_PORT", "")
	t.Setenv("DEFAULT_MODEL", "")
	mux, _ := buildMux()

	cases := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/health", http.StatusOK},
		{http.MethodGet, "/v1/models", http.StatusOK},
		{http.MethodGet, "/", http.StatusOK},
		{http.MethodGet, "/unknown/path", http.StatusNotFound},
		// POST without body → 405 or 400/413 depending on Content-Type; just ensure route is wired.
		{http.MethodGet, "/v1/audio/transcriptions", http.StatusMethodNotAllowed},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != tc.want {
			t.Errorf("%s %s: want %d, got %d", tc.method, tc.path, tc.want, rr.Code)
		}
	}
}
