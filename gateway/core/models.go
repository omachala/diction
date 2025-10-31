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

// ModelsHandler returns the handler for GET /v1/models.
func (g *Gateway) ModelsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		models := make([]modelInfo, 0, len(g.backends))
		for _, b := range g.backends {
			models = append(models, modelInfo{
				ID:          b.Name,
				Name:        b.DisplayName,
				Description: b.Description,
				Available:   g.health.get(b.Name),
			})
		}

		resp := modelsResponse{
			Providers: []providerInfo{
				{
					ID:     "faster-whisper",
					Name:   "Faster Whisper",
					Models: models,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
