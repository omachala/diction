package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// --- proxyToBackend ---

func TestProxyToBackend_Success(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audio/transcriptions" {
			http.Error(w, "wrong path", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"text":"hello from backend"}`)
	}))
	defer backend.Close()

	g := testGateway()
	target, _ := url.Parse(backend.URL)
	pcm := make([]byte, 3200) // 0.1s of silence at 16kHz 16-bit mono

	text, err := g.proxyToBackend(context.Background(), target, pcm, &Backend{Name: "small"}, "en")
	if err != nil {
		t.Fatalf("proxyToBackend: %v", err)
	}
	if text != "hello from backend" {
		t.Errorf("text: want 'hello from backend', got %q", text)
	}
}

func TestProxyToBackend_NoLanguage(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"text":"no lang"}`)
	}))
	defer backend.Close()

	g := testGateway()
	target, _ := url.Parse(backend.URL)

	text, err := g.proxyToBackend(context.Background(), target, make([]byte, 100), &Backend{Name: "small"}, "")
	if err != nil {
		t.Fatalf("proxyToBackend with empty language: %v", err)
	}
	if text != "no lang" {
		t.Errorf("got: %q", text)
	}
}

func TestProxyToBackend_BackendError(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer backend.Close()

	g := testGateway()
	target, _ := url.Parse(backend.URL)

	_, err := g.proxyToBackend(context.Background(), target, make([]byte, 100), &Backend{Name: "small"}, "")
	if err == nil {
		t.Fatal("expected error for backend 500")
	}
}

func TestProxyToBackend_InvalidJSON(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `not-json`)
	}))
	defer backend.Close()

	g := testGateway()
	target, _ := url.Parse(backend.URL)

	_, err := g.proxyToBackend(context.Background(), target, make([]byte, 100), &Backend{Name: "small"}, "")
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestProxyToBackend_UnreachableHost(t *testing.T) {
	g := testGateway()
	target, _ := url.Parse("http://127.0.0.1:1") // nothing listening there

	_, err := g.proxyToBackend(context.Background(), target, make([]byte, 100), &Backend{Name: "small"}, "")
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
}

// --- StreamingHandler ---

// startStreamingServer creates a Gateway with a mock whisper backend and
// returns a test HTTP server running the StreamingHandler, plus a cleanup func.
func startStreamingServer(t *testing.T, whisperResponse string, whisperStatus int) (*httptest.Server, *Gateway) {
	t.Helper()

	// Mock whisper backend
	whisper := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(whisperStatus)
		if whisperStatus == http.StatusOK {
			fmt.Fprintf(w, `{"text":"%s"}`, whisperResponse)
		}
	}))
	t.Cleanup(whisper.Close)

	g := &Gateway{
		backends: []Backend{
			{Name: "small", URL: whisper.URL, Aliases: []string{"small"}},
		},
		health:       newHealthState(),
		defaultModel: "small",
		maxBodySize:  10 * 1024 * 1024,
	}
	g.health.set("small", true)

	srv := httptest.NewServer(g.StreamingHandler())
	t.Cleanup(srv.Close)
	return srv, g
}

func wsURL(srv *httptest.Server, query string) string {
	u := "ws" + strings.TrimPrefix(srv.URL, "http") + "/v1/audio/stream"
	if query != "" {
		u += "?" + query
	}
	return u
}

func TestStreamingHandler_NoMatchingBackend(t *testing.T) {
	// defaultModel has no matching backend → HTTP 503 before WebSocket upgrade
	g := &Gateway{
		backends:     []Backend{},
		health:       newHealthState(),
		defaultModel: "small",
		maxBodySize:  10 * 1024 * 1024,
	}
	srv := httptest.NewServer(g.StreamingHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/v1/audio/stream")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status: want 503, got %d", resp.StatusCode)
	}
}

func TestStreamingHandler_BackendUnavailable(t *testing.T) {
	srv, g := startStreamingServer(t, "ok", http.StatusOK)
	g.health.set("small", false) // mark as down

	resp, err := http.Get(srv.URL + "/v1/audio/stream")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status: want 503, got %d", resp.StatusCode)
	}
}

func TestStreamingHandler_HappyPath(t *testing.T) {
	srv, _ := startStreamingServer(t, "transcribed text", http.StatusOK)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv, "model=small&language=en"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	// Send PCM audio (silence)
	pcm := make([]byte, 3200)
	if err := conn.Write(ctx, websocket.MessageBinary, pcm); err != nil {
		t.Fatalf("write audio: %v", err)
	}

	// Send done action
	done, _ := json.Marshal(map[string]string{"action": "done"})
	if err := conn.Write(ctx, websocket.MessageText, done); err != nil {
		t.Fatalf("write done: %v", err)
	}

	// Read transcription result
	msgType, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	if msgType != websocket.MessageText {
		t.Errorf("msg type: want text, got %v", msgType)
	}
	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if result.Text != "transcribed text" {
		t.Errorf("text: want 'transcribed text', got %q", result.Text)
	}
}

func TestStreamingHandler_LanguageOverrideInDone(t *testing.T) {
	srv, _ := startStreamingServer(t, "override", http.StatusOK)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv, "model=small"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	conn.Write(ctx, websocket.MessageBinary, make([]byte, 3200))

	// done action with language override
	done, _ := json.Marshal(map[string]string{"action": "done", "language": "fr"})
	conn.Write(ctx, websocket.MessageText, done)

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	var result struct {
		Text string `json:"text"`
	}
	json.Unmarshal(data, &result)
	if result.Text != "override" {
		t.Errorf("text: want 'override', got %q", result.Text)
	}
}

func TestStreamingHandler_NoAudio(t *testing.T) {
	srv, _ := startStreamingServer(t, "ok", http.StatusOK)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv, "model=small"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	// Send done with no preceding audio
	done, _ := json.Marshal(map[string]string{"action": "done"})
	conn.Write(ctx, websocket.MessageText, done)

	// Server should close with wsCloseNoAudio (4004)
	_, _, err = conn.Read(ctx)
	if err == nil {
		t.Fatal("expected connection to be closed by server")
	}
}

func TestStreamingHandler_AudioTooLarge(t *testing.T) {
	srv, g := startStreamingServer(t, "ok", http.StatusOK)
	g.maxBodySize = 100 // tiny limit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv, "model=small"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	// Send more than maxBodySize bytes
	conn.Write(ctx, websocket.MessageBinary, make([]byte, 200))

	_, _, err = conn.Read(ctx)
	if err == nil {
		t.Fatal("expected connection closed due to oversized audio")
	}
}

func TestStreamingHandler_BackendTranscriptionFails(t *testing.T) {
	srv, _ := startStreamingServer(t, "", http.StatusInternalServerError)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv, "model=small"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	conn.Write(ctx, websocket.MessageBinary, make([]byte, 3200))
	done, _ := json.Marshal(map[string]string{"action": "done"})
	conn.Write(ctx, websocket.MessageText, done)

	_, _, err = conn.Read(ctx)
	if err == nil {
		t.Fatal("expected connection closed due to transcription failure")
	}
}

func TestStreamingHandler_NonDoneTextMessage(t *testing.T) {
	// Non-"done" text messages are ignored; server keeps waiting.
	// We send a non-done message then a done to confirm normal flow still works.
	srv, _ := startStreamingServer(t, "after ignore", http.StatusOK)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv, "model=small"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	conn.Write(ctx, websocket.MessageBinary, make([]byte, 3200))
	// Send a text message with unknown action - should be ignored
	conn.Write(ctx, websocket.MessageText, []byte(`{"action":"unknown"}`))
	// Now send done
	done, _ := json.Marshal(map[string]string{"action": "done"})
	conn.Write(ctx, websocket.MessageText, done)

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var result struct {
		Text string `json:"text"`
	}
	json.Unmarshal(data, &result)
	if result.Text != "after ignore" {
		t.Errorf("want 'after ignore', got %q", result.Text)
	}
}

func TestStreamingHandler_InvalidTextIgnored(t *testing.T) {
	// Non-JSON text → json.Unmarshal fails → continue; server keeps waiting.
	srv, _ := startStreamingServer(t, "after garbage", http.StatusOK)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv, "model=small"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	conn.Write(ctx, websocket.MessageBinary, make([]byte, 3200))
	conn.Write(ctx, websocket.MessageText, []byte("not-json-at-all"))
	done, _ := json.Marshal(map[string]string{"action": "done"})
	conn.Write(ctx, websocket.MessageText, done)

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var result struct {
		Text string `json:"text"`
	}
	json.Unmarshal(data, &result)
	if result.Text != "after garbage" {
		t.Errorf("want 'after garbage', got %q", result.Text)
	}
}

func TestStreamingHandler_HealthByAliasName(t *testing.T) {
	// When model query param is an alias whose case differs from the backend Name
	// (e.g. "SMALL" vs stored health key "small"), g.health.get(model) returns
	// false - but the alias-scan loop should find the backend is healthy and allow
	// the upgrade. Covers the backendUp=true branch inside the health-check block.
	srv, _ := startStreamingServer(t, "alias ok", http.StatusOK)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// "SMALL" is an alias for "small"; health key is stored as "small"
	conn, _, err := websocket.Dial(ctx, wsURL(srv, "model=SMALL"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	conn.Write(ctx, websocket.MessageBinary, make([]byte, 3200))
	done, _ := json.Marshal(map[string]string{"action": "done"})
	conn.Write(ctx, websocket.MessageText, done)

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var result struct {
		Text string `json:"text"`
	}
	json.Unmarshal(data, &result)
	if result.Text != "alias ok" {
		t.Errorf("want 'alias ok', got %q", result.Text)
	}
}

func startStreamingServerWithPostProcess(t *testing.T, whisperResponse string, postProcess func(context.Context, string, string) (string, error)) *httptest.Server {
	t.Helper()
	whisper := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"text":"%s"}`, whisperResponse)
	}))
	t.Cleanup(whisper.Close)

	g := &Gateway{
		backends: []Backend{
			{Name: "small", URL: whisper.URL, Aliases: []string{"small"}},
		},
		health:       newHealthState(),
		defaultModel: "small",
		maxBodySize:  10 * 1024 * 1024,
	}
	g.health.set("small", true)

	srv := httptest.NewServer(g.StreamingHandlerWithPostProcess(postProcess))
	t.Cleanup(srv.Close)
	return srv
}

func TestStreamingHandler_WithPostProcess_Success(t *testing.T) {
	postProcess := func(ctx context.Context, text, contextJSON string) (string, error) {
		return "cleaned: " + text, nil
	}
	srv := startStreamingServerWithPostProcess(t, "raw audio", postProcess)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv, "enhance=true"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	conn.Write(ctx, websocket.MessageBinary, make([]byte, 3200))
	done, _ := json.Marshal(map[string]string{"action": "done"})
	conn.Write(ctx, websocket.MessageText, done)

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var result struct {
		Text string `json:"text"`
	}
	json.Unmarshal(data, &result)
	if result.Text != "cleaned: raw audio" {
		t.Errorf("want 'cleaned: raw audio', got %q", result.Text)
	}
}

func TestStreamingHandler_WithPostProcess_ErrorFallback(t *testing.T) {
	// postProcess error → raw transcript returned unchanged
	postProcess := func(ctx context.Context, text, contextJSON string) (string, error) {
		return "", fmt.Errorf("llm error")
	}
	srv := startStreamingServerWithPostProcess(t, "raw audio", postProcess)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv, "enhance=true"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	conn.Write(ctx, websocket.MessageBinary, make([]byte, 3200))
	done, _ := json.Marshal(map[string]string{"action": "done"})
	conn.Write(ctx, websocket.MessageText, done)

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var result struct {
		Text string `json:"text"`
	}
	json.Unmarshal(data, &result)
	if result.Text != "raw audio" {
		t.Errorf("want 'raw audio' fallback, got %q", result.Text)
	}
}

func TestStreamingHandler_DefaultModel(t *testing.T) {
	// No ?model param → uses default model
	srv, _ := startStreamingServer(t, "default model used", http.StatusOK)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv, ""), nil) // no model param
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	conn.Write(ctx, websocket.MessageBinary, make([]byte, 3200))
	done, _ := json.Marshal(map[string]string{"action": "done"})
	conn.Write(ctx, websocket.MessageText, done)

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var result struct {
		Text string `json:"text"`
	}
	json.Unmarshal(data, &result)
	if result.Text != "default model used" {
		t.Errorf("want 'default model used', got %q", result.Text)
	}
}

func TestStreamingHandler_FirstTextFrameAsContext(t *testing.T) {
	var receivedContext string
	postProcess := func(ctx context.Context, text, contextJSON string) (string, error) {
		receivedContext = contextJSON
		return "cleaned: " + text, nil
	}
	srv := startStreamingServerWithPostProcess(t, "raw audio", postProcess)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv, "enhance=true"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	// Send context JSON as first text frame
	contextJSON := `{"before":"hello ","after":" world","selected":"test"}`
	conn.Write(ctx, websocket.MessageText, []byte(contextJSON))

	// Send audio
	conn.Write(ctx, websocket.MessageBinary, make([]byte, 3200))

	// Send done
	done, _ := json.Marshal(map[string]string{"action": "done"})
	conn.Write(ctx, websocket.MessageText, done)

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var result struct {
		Text string `json:"text"`
	}
	json.Unmarshal(data, &result)
	if result.Text != "cleaned: raw audio" {
		t.Errorf("want 'cleaned: raw audio', got %q", result.Text)
	}
	if receivedContext != contextJSON {
		t.Errorf("context: want %q, got %q", contextJSON, receivedContext)
	}
}
