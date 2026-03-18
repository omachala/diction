package core

// Backend describes a speech-to-text model backend.
type Backend struct {
	Name            string
	URL             string
	Aliases         []string
	DisplayName     string
	Description     string
	Provider        string // "whisper" or "parakeet"
	NeedsWAV        bool   // if true, gateway converts audio to WAV before proxying
	ForwardModel    string // model name to inject into forwarded request; empty = don't inject
	AuthHeader      string // Authorization header value to inject into backend requests; empty = none
	SkipHealthCheck bool   // if true, skip health polling (custom/external backends)
}

// CustomBackendFromEnv builds a custom backend from environment variables.
// Returns nil if CUSTOM_BACKEND_URL is not set.
//
// Supported env vars:
//
//	CUSTOM_BACKEND_URL       (required) base URL of the backend, e.g. http://192.168.1.50:8000
//	CUSTOM_BACKEND_MODEL     model name to forward in the request, e.g. Systran/faster-whisper-large-v3-turbo
//	CUSTOM_BACKEND_NEEDS_WAV set to "true" if the backend only accepts WAV audio (default: false)
//	CUSTOM_BACKEND_AUTH      Authorization header value, e.g. "Bearer sk-xxx" (default: none)
func CustomBackendFromEnv() *Backend {
	rawURL := EnvOrDefault("CUSTOM_BACKEND_URL", "")
	if rawURL == "" {
		return nil
	}
	return &Backend{
		Name:            "custom",
		URL:             rawURL,
		Aliases:         []string{"custom"},
		DisplayName:     "Custom",
		Description:     "custom backend",
		Provider:        "custom",
		NeedsWAV:        EnvBoolOrDefault("CUSTOM_BACKEND_NEEDS_WAV", false),
		ForwardModel:    EnvOrDefault("CUSTOM_BACKEND_MODEL", ""),
		AuthHeader:      EnvOrDefault("CUSTOM_BACKEND_AUTH", ""),
		SkipHealthCheck: true,
	}
}

// DefaultBackends returns the standard set of STT backends.
func DefaultBackends() []Backend {
	return []Backend{
		// Whisper models (faster-whisper-server, accepts any audio format)
		// Tiny is reserved for on-device inference, not exposed via gateway
		// {Name: "tiny", URL: "http://whisper-tiny:8000", Aliases: []string{"tiny", "Systran/faster-whisper-tiny"}, DisplayName: "Tiny", Description: "fastest, best for quick notes in quiet environments", Provider: "whisper"},
		{Name: "small", URL: "http://whisper-small:8000", Aliases: []string{"small", "Systran/faster-whisper-small"}, DisplayName: "Small", Description: "fast, good for everyday dictation", Provider: "whisper"},
		{Name: "medium", URL: "http://whisper-medium:8000", Aliases: []string{"medium", "Systran/faster-whisper-medium"}, DisplayName: "Medium", Description: "slower, handles accents and background noise better", Provider: "whisper"},
		{Name: "large-v3-turbo", URL: "http://whisper-large-turbo:8000", Aliases: []string{"large-v3-turbo", "turbo", "deepdml/faster-whisper-large-v3-turbo-ct2"}, DisplayName: "Large", Description: "highest accuracy Whisper model, best for difficult audio", Provider: "whisper"},
		{Name: "distil-large-v3", URL: "http://whisper-distil-large:8000", Aliases: []string{"distil-large-v3", "Systran/faster-distil-whisper-large-v3"}, DisplayName: "Large", Description: "highest accuracy Whisper model, English only", Provider: "whisper"},

		// Parakeet (NVIDIA, OpenAI-compatible API, WAV only)
		{Name: "parakeet-v3", URL: "http://parakeet:5092", Aliases: []string{"parakeet-v3", "parakeet", "parakeet-tdt-0.6b-v3"}, DisplayName: "Parakeet", Description: "best overall accuracy and speed, 25 European languages", Provider: "parakeet", NeedsWAV: true},
	}
}
