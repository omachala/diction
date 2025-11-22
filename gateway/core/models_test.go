package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestModelsHandler_MethodNotAllowed(t *testing.T) {
	g := testGateway()
	req := httptest.NewRequest(http.MethodPost, "/v1/models", nil)
	rr := httptest.NewRecorder()
	g.ModelsHandler()(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: want 405, got %d", rr.Code)
	}
}

func TestModelsHandler_ReturnsJSON(t *testing.T) {
	g := testGateway()
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rr := httptest.NewRecorder()
	g.ModelsHandler()(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: want application/json, got %s", ct)
	}

	var resp modelsResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Providers) == 0 {
		t.Error("expected at least one provider")
	}
}

func TestModelsHandler_GroupsByProvider(t *testing.T) {
	g := &Gateway{
		backends: []Backend{
			{Name: "small", Aliases: []string{"small"}, Provider: "whisper", DisplayName: "Small"},
			{Name: "medium", Aliases: []string{"medium"}, Provider: "whisper", DisplayName: "Medium"},
			{Name: "parakeet-v3", Aliases: []string{"parakeet-v3"}, Provider: "parakeet", DisplayName: "Parakeet"},
		},
		health:       newHealthState(),
		defaultModel: "small",
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rr := httptest.NewRecorder()
	g.ModelsHandler()(rr, req)

	var resp modelsResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// Expect 2 providers: whisper (2 models) and parakeet (1 model)
	if len(resp.Providers) != 2 {
		t.Fatalf("providers: want 2, got %d", len(resp.Providers))
	}

	whisperProvider := resp.Providers[0]
	if whisperProvider.ID != "whisper" {
		t.Errorf("first provider ID: want whisper, got %s", whisperProvider.ID)
	}
	if len(whisperProvider.Models) != 2 {
		t.Errorf("whisper models: want 2, got %d", len(whisperProvider.Models))
	}

	parakeetProvider := resp.Providers[1]
	if parakeetProvider.ID != "parakeet" {
		t.Errorf("second provider ID: want parakeet, got %s", parakeetProvider.ID)
	}
	if len(parakeetProvider.Models) != 1 {
		t.Errorf("parakeet models: want 1, got %d", len(parakeetProvider.Models))
	}
}

func TestModelsHandler_ProviderDisplayName(t *testing.T) {
	g := &Gateway{
		backends: []Backend{
			{Name: "small", Aliases: []string{"small"}, Provider: "whisper"},
		},
		health:       newHealthState(),
		defaultModel: "small",
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rr := httptest.NewRecorder()
	g.ModelsHandler()(rr, req)

	var resp modelsResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.Providers[0].Name != "Faster Whisper" {
		t.Errorf("provider name: want 'Faster Whisper', got %s", resp.Providers[0].Name)
	}
}

func TestModelsHandler_AvailabilityReflectsHealth(t *testing.T) {
	g := &Gateway{
		backends: []Backend{
			{Name: "small", Aliases: []string{"small"}, Provider: "whisper"},
		},
		health:       newHealthState(),
		defaultModel: "small",
	}

	// Mark backend as available
	g.health.set("small", true)

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rr := httptest.NewRecorder()
	g.ModelsHandler()(rr, req)

	var resp modelsResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if !resp.Providers[0].Models[0].Available {
		t.Error("expected model to be available")
	}

	// Mark backend as unavailable
	g.health.set("small", false)

	req = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rr = httptest.NewRecorder()
	g.ModelsHandler()(rr, req)
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.Providers[0].Models[0].Available {
		t.Error("expected model to be unavailable")
	}
}

func TestModelsHandler_EmptyProvider_DefaultsToWhisper(t *testing.T) {
	g := &Gateway{
		backends: []Backend{
			{Name: "small", Aliases: []string{"small"}, Provider: ""}, // empty provider
		},
		health:       newHealthState(),
		defaultModel: "small",
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rr := httptest.NewRecorder()
	g.ModelsHandler()(rr, req)

	var resp modelsResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.Providers[0].ID != "whisper" {
		t.Errorf("empty provider should default to 'whisper', got %s", resp.Providers[0].ID)
	}
}
