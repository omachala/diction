package core

// Backend describes a whisper model backend.
type Backend struct {
	Name        string
	URL         string
	Aliases     []string
	DisplayName string
	Description string
}

// DefaultBackends returns the standard set of whisper backends.
func DefaultBackends() []Backend {
	return []Backend{
		{Name: "tiny", URL: "http://whisper-tiny:8000", Aliases: []string{"tiny", "Systran/faster-whisper-tiny"}, DisplayName: "Tiny", Description: "fastest, best for quick notes in quiet environments"},
		{Name: "small", URL: "http://whisper-small:8000", Aliases: []string{"small", "Systran/faster-whisper-small"}, DisplayName: "Small", Description: "fast, good for everyday dictation"},
		{Name: "medium", URL: "http://whisper-medium:8000", Aliases: []string{"medium", "Systran/faster-whisper-medium"}, DisplayName: "Medium", Description: "slower, handles accents and background noise better"},
		{Name: "large-v3", URL: "http://whisper-large:8000", Aliases: []string{"large-v3", "Systran/faster-whisper-large-v3"}, DisplayName: "Large V3", Description: "slowest, best accuracy even with difficult audio"},
		{Name: "distil-large-v3", URL: "http://whisper-distil-large:8000", Aliases: []string{"distil-large-v3", "Systran/faster-distil-whisper-large-v3"}, DisplayName: "Distil Large V3", Description: "near large-v3 accuracy at 6x speed"},
	}
}
