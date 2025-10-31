// Diction Gateway — community edition (no auth).
//
// Routes transcription requests to the correct whisper backend by model name.
// Supports both HTTP POST and WebSocket streaming.
//
// Endpoints:
//   GET  /health                      → {"status":"ok"}
//   GET  /v1/models                   → list available models with health status
//   POST /v1/audio/transcriptions     → proxy multipart audio to whisper backend
//   WS   /v1/audio/stream?model=small → stream PCM audio, get transcription back
package main

import (
	"log"
	"net/http"

	"github.com/omachala/diction/gateway/core"
)

func main() {
	port := core.EnvOrDefault("GATEWAY_PORT", "8080")
	defaultModel := core.EnvOrDefault("DEFAULT_MODEL", "small")
	maxBodySize := int64(core.EnvIntOrDefault("MAX_BODY_SIZE", 10485760))

	gw := core.NewGateway(core.Config{
		Backends:     core.DefaultBackends(),
		DefaultModel: defaultModel,
		MaxBodySize:  maxBodySize,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/health", gw.HealthHandler())
	mux.HandleFunc("/v1/models", gw.ModelsHandler())
	mux.HandleFunc("/v1/audio/transcriptions", gw.TranscriptionHandler())
	mux.HandleFunc("/v1/audio/stream", gw.StreamingHandler())
	mux.HandleFunc("/", gw.CatchAllHandler())

	log.Printf("Diction Gateway starting on :%s (default_model=%s)", port, defaultModel)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
