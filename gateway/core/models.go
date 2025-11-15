package core

import (
	"encoding/json"
	"net/http"
)

type modelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Available   bool   `json:"available"`
}

type providerInfo struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Models []modelInfo `json:"models"`
}

type modelsResponse struct {
	Providers []providerInfo `json:"providers"`
}

// provider display names
var providerNames = map[string]string{
	"whisper":  "Faster Whisper",
	"parakeet": "NVIDIA Parakeet",
}

// ModelsHandler returns the handler for GET /v1/models.
func (g *Gateway) ModelsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// Group backends by provider
		grouped := make(map[string][]modelInfo)
		providerOrder := make([]string, 0)
		for _, b := range g.backends {
			m := modelInfo{
				ID:          b.Name,
				Name:        b.DisplayName,
				Description: b.Description,
				Available:   g.health.get(b.Name),
			}
			p := b.Provider
			if p == "" {
				p = "whisper"
			}
			if _, exists := grouped[p]; !exists {
				providerOrder = append(providerOrder, p)
			}
			grouped[p] = append(grouped[p], m)
		}

		providers := make([]providerInfo, 0, len(grouped))
		for _, p := range providerOrder {
			name := providerNames[p]
			if name == "" {
				name = p
			}
			providers = append(providers, providerInfo{
				ID:     p,
				Name:   name,
				Models: grouped[p],
			})
		}

		resp := modelsResponse{Providers: providers}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
