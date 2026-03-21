package core

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// buildMultipart builds a multipart/form-data body with the given fields and an
// optional file part. Returns the body bytes and the content-type header value.
func buildMultipart(t *testing.T, fields map[string]string, fileName, fileContent string) ([]byte, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("WriteField %s: %v", k, err)
		}
	}
	if fileName != "" {
		fw, err := w.CreateFormFile("file", fileName)
		if err != nil {
			t.Fatalf("CreateFormFile: %v", err)
		}
		fw.Write([]byte(fileContent))
	}
	w.Close()
	return buf.Bytes(), w.FormDataContentType()
}

// --- parseBoundary ---

func TestParseBoundary_ExtractsBoundary(t *testing.T) {
	_, ct := buildMultipart(t, map[string]string{}, "audio.m4a", "fake-audio")
	boundary := parseBoundary(ct)
	if boundary == "" {
		t.Error("expected non-empty boundary")
	}
}

func TestParseBoundary_InvalidContentType(t *testing.T) {
	boundary := parseBoundary("application/json")
	if boundary != "" {
		t.Errorf("boundary: want empty for non-multipart, got %s", boundary)
	}
}

func TestParseBoundary_ExplicitBoundary(t *testing.T) {
	boundary := parseBoundary("multipart/form-data; boundary=abc123")
	if boundary != "abc123" {
		t.Errorf("boundary: want abc123, got %s", boundary)
	}
}

// --- rewriteMultipart ---

func TestRewriteMultipart_StripsModelField(t *testing.T) {
	body, ct := buildMultipart(t, map[string]string{"model": "medium", "language": "en"}, "audio.m4a", "fake-audio-data")
	boundary := parseBoundary(ct)

	rewritten, newCT, err := rewriteMultipart(body, boundary, false, "")
	if err != nil {
		t.Fatalf("rewriteMultipart error: %v", err)
	}

	// Parse the rewritten body and verify no model field
	_, params, _ := parseMediaType(newCT)
	newBoundary := params["boundary"]
	reader := multipart.NewReader(bytes.NewReader(rewritten), newBoundary)
	foundModel := false
	foundLanguage := false
	foundFile := false
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read rewritten part: %v", err)
		}
		switch part.FormName() {
		case "model":
			foundModel = true
		case "language":
			foundLanguage = true
		case "file":
			foundFile = true
		}
		part.Close()
	}

	if foundModel {
		t.Error("model field should have been stripped")
	}
	if !foundLanguage {
		t.Error("language field should be preserved")
	}
	if !foundFile {
		t.Error("file part should be preserved")
	}
}

func TestRewriteMultipart_PreservesFileContent(t *testing.T) {
	const audioContent = "this-is-fake-audio-bytes"
	body, ct := buildMultipart(t, map[string]string{"model": "small"}, "audio.m4a", audioContent)
	boundary := parseBoundary(ct)

	rewritten, newCT, err := rewriteMultipart(body, boundary, false, "")
	if err != nil {
		t.Fatalf("rewriteMultipart error: %v", err)
	}

	_, params, _ := parseMediaType(newCT)
	reader := multipart.NewReader(bytes.NewReader(rewritten), params["boundary"])
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read rewritten part: %v", err)
		}
		if part.FormName() == "file" {
			data, _ := io.ReadAll(part)
			if string(data) != audioContent {
				t.Errorf("file content: want %q, got %q", audioContent, string(data))
			}
		}
		part.Close()
	}
}

func TestRewriteMultipart_NoModelField(t *testing.T) {
	// No model field - rewrite should still work cleanly
	body, ct := buildMultipart(t, map[string]string{}, "audio.m4a", "audio")
	boundary := parseBoundary(ct)

	_, _, err := rewriteMultipart(body, boundary, false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRewriteMultipart_ConvertToWAV(t *testing.T) {
	// When convertToWAV=true with already-WAV input (RIFF passthrough), the file
	// part is re-emitted with Content-Type: audio/wav and filename "audio.wav".
	wavData := []byte("RIFF\x00\x00\x00\x00WAVEfmt ")
	body, ct := buildMultipart(t, map[string]string{"language": "en"}, "audio.m4a", string(wavData))
	boundary := parseBoundary(ct)

	rewritten, newCT, err := rewriteMultipart(body, boundary, true, "")
	if err != nil {
		t.Fatalf("rewriteMultipart error: %v", err)
	}

	_, params, _ := parseMediaType(newCT)
	reader := multipart.NewReader(bytes.NewReader(rewritten), params["boundary"])
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read part: %v", err)
		}
		if part.FormName() == "file" {
			if part.Header.Get("Content-Type") != "audio/wav" {
				t.Errorf("Content-Type: want audio/wav, got %s", part.Header.Get("Content-Type"))
			}
			if part.FileName() != "audio.wav" {
				t.Errorf("filename: want audio.wav, got %s", part.FileName())
			}
		}
		part.Close()
	}
}

// --- convertToWAVBytes ---

func TestConvertToWAVBytes_PassthroughWhenAlreadyWAV(t *testing.T) {
	// WAV files start with RIFF magic - should be returned unchanged.
	wavData := []byte("RIFF\x00\x00\x00\x00WAVEfmt ")
	result, err := convertToWAVBytes(wavData, "audio.wav")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(result, wavData) {
		t.Error("expected WAV data to be passed through unchanged")
	}
}

func TestConvertToWAVBytes_FFmpegErrorOnGarbage(t *testing.T) {
	// Non-RIFF garbage - exercises temp dir creation, file write, ffmpeg exec,
	// and the error-return path (ffmpeg can't decode random bytes as .m4a).
	garbage := bytes.Repeat([]byte{0xDE, 0xAD, 0xBE, 0xEF}, 64)
	_, err := convertToWAVBytes(garbage, "audio.m4a")
	if err == nil {
		t.Fatal("expected error when ffmpeg cannot decode garbage audio")
	}
}

func TestConvertToWAVBytes_EmptyFilename(t *testing.T) {
	// Empty filename → falls back to .m4a extension; ffmpeg still fails on garbage.
	garbage := bytes.Repeat([]byte{0xFF, 0xFE}, 32)
	_, err := convertToWAVBytes(garbage, "")
	if err == nil {
		t.Fatal("expected error for undecipherable audio without filename")
	}
}

// --- parseMediaType is a helper to avoid importing mime at the test level.
func parseMediaType(ct string) (string, map[string]string, error) {
	// simple parsing: split on ;
	parts := strings.SplitN(ct, ";", 2)
	params := make(map[string]string)
	if len(parts) == 2 {
		for _, p := range strings.Split(parts[1], ";") {
			p = strings.TrimSpace(p)
			kv := strings.SplitN(p, "=", 2)
			if len(kv) == 2 {
				params[strings.TrimSpace(kv[0])] = strings.Trim(strings.TrimSpace(kv[1]), `"`)
			}
		}
	}
	return strings.TrimSpace(parts[0]), params, nil
}

// --- TranscriptionHandler ---

func TestTranscriptionHandler_MethodNotAllowed(t *testing.T) {
	g := testGateway()
	req := httptest.NewRequest(http.MethodGet, "/v1/audio/transcriptions", nil)
	rr := httptest.NewRecorder()
	g.TranscriptionHandler()(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: want 405, got %d", rr.Code)
	}
}

func TestTranscriptionHandler_NoMatchingBackend(t *testing.T) {
	// defaultModel has no matching backend - gateway returns 400
	g := &Gateway{
		backends:     []Backend{},
		health:       newHealthState(),
		defaultModel: "small",
		maxBodySize:  10 * 1024 * 1024,
	}
	body, ct := buildMultipart(t, map[string]string{}, "audio.m4a", "data")
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	g.TranscriptionHandler()(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: want 400, got %d", rr.Code)
	}
}

func TestTranscriptionHandler_BodyTooLarge(t *testing.T) {
	g := &Gateway{
		backends:     testGateway().backends,
		health:       newHealthState(),
		defaultModel: "small",
		maxBodySize:  10, // tiny limit
	}

	body, ct := buildMultipart(t, map[string]string{"model": "small"}, "audio.m4a", "this-is-way-too-long")
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	g.TranscriptionHandler()(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status: want 413, got %d", rr.Code)
	}
}

func TestTranscriptionHandler_ProxiesToBackend(t *testing.T) {
	// Start a fake whisper backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audio/transcriptions" {
			http.Error(w, "wrong path", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"text":"hello world"}`)
	}))
	defer backend.Close()

	g := &Gateway{
		backends: []Backend{
			{Name: "small", URL: backend.URL, Aliases: []string{"small"}},
		},
		health:       newHealthState(),
		defaultModel: "small",
		maxBodySize:  10 * 1024 * 1024,
	}

	body, ct := buildMultipart(t, map[string]string{"model": "small"}, "audio.m4a", "fake-audio")
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	g.TranscriptionHandler()(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d (body: %s)", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "hello world") {
		t.Errorf("expected transcription in body, got: %s", rr.Body.String())
	}
}

func TestTranscriptionHandler_WAVConversionFails(t *testing.T) {
	// Backend with NeedsWAV=true: gateway runs ffmpeg on the audio; garbage input
	// makes ffmpeg fail → rewriteMultipart returns error → handler returns 500.
	g := &Gateway{
		backends: []Backend{
			{Name: "parakeet-v3", URL: "http://parakeet:5092", Aliases: []string{"parakeet-v3"}, NeedsWAV: true},
		},
		health:       newHealthState(),
		defaultModel: "parakeet-v3",
		maxBodySize:  10 * 1024 * 1024,
	}

	garbage := bytes.Repeat([]byte{0xDE, 0xAD, 0xBE, 0xEF}, 64)
	body, ct := buildMultipart(t, map[string]string{"model": "parakeet-v3"}, "audio.m4a", string(garbage))
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	g.TranscriptionHandler()(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status: want 500, got %d (body: %s)", rr.Code, rr.Body.String())
	}
}

func TestTranscriptionHandler_WithPostProcess_Success(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"text":"raw transcript"}`)
	}))
	defer backend.Close()

	g := &Gateway{
		backends: []Backend{
			{Name: "small", URL: backend.URL, Aliases: []string{"small"}},
		},
		health:       newHealthState(),
		defaultModel: "small",
		maxBodySize:  10 * 1024 * 1024,
	}

	postProcess := func(ctx context.Context, text, contextJSON string) (string, error) {
		return "cleaned: " + text, nil
	}

	body, ct := buildMultipart(t, map[string]string{"model": "small"}, "audio.m4a", "fake-audio")
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions?enhance=true", bytes.NewReader(body))
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	g.TranscriptionHandlerWithPostProcess(postProcess)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d (body: %s)", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "cleaned: raw transcript") {
		t.Errorf("expected cleaned transcript, got: %s", rr.Body.String())
	}
}

func TestTranscriptionHandler_WithPostProcess_ErrorFallback(t *testing.T) {
	// postProcess error → raw transcript returned unchanged
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"text":"raw transcript"}`)
	}))
	defer backend.Close()

	g := &Gateway{
		backends: []Backend{
			{Name: "small", URL: backend.URL, Aliases: []string{"small"}},
		},
		health:       newHealthState(),
		defaultModel: "small",
		maxBodySize:  10 * 1024 * 1024,
	}

	postProcess := func(ctx context.Context, text, contextJSON string) (string, error) {
		return "", fmt.Errorf("llm error")
	}

	body, ct := buildMultipart(t, map[string]string{"model": "small"}, "audio.m4a", "fake-audio")
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions?enhance=true", bytes.NewReader(body))
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	g.TranscriptionHandlerWithPostProcess(postProcess)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "raw transcript") {
		t.Errorf("expected raw transcript fallback, got: %s", rr.Body.String())
	}
}

func TestTranscriptionHandler_ModelStrippedBeforeProxy(t *testing.T) {
	// Verify model field is not forwarded to backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "parse failed", 400)
			return
		}
		if r.FormValue("model") != "" {
			http.Error(w, "model field should not be present", 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"text":"ok"}`)
	}))
	defer backend.Close()

	g := &Gateway{
		backends: []Backend{
			{Name: "small", URL: backend.URL, Aliases: []string{"small"}},
		},
		health:       newHealthState(),
		defaultModel: "small",
		maxBodySize:  10 * 1024 * 1024,
	}

	body, ct := buildMultipart(t, map[string]string{"model": "small", "language": "en"}, "audio.m4a", "audio")
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	g.TranscriptionHandler()(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d (body: %s)", rr.Code, rr.Body.String())
	}
}

func TestTranscriptionHandler_E2EEncryption(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"text":"hello world"}`)
	}))
	defer backend.Close()

	g := &Gateway{
		backends: []Backend{
			{Name: "small", URL: backend.URL, Aliases: []string{"small"}},
		},
		health:       newHealthState(),
		defaultModel: "small",
		maxBodySize:  10 * 1024 * 1024,
	}

	// Generate client ephemeral key
	clientPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	clientPubB64 := base64.RawURLEncoding.EncodeToString(clientPriv.PublicKey().Bytes())

	body, ct := buildMultipart(t, map[string]string{}, "audio.m4a", "fake-audio")
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-Diction-E2E", clientPubB64)
	rr := httptest.NewRecorder()
	g.TranscriptionHandlerWithPostProcess(nil)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d (body: %s)", rr.Code, rr.Body.String())
	}

	// Response should contain e2e envelope
	var resp struct {
		E2E struct {
			CT string `json:"ct"`
			PK string `json:"pk"`
		} `json:"e2e"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.E2E.CT == "" || resp.E2E.PK == "" {
		t.Fatalf("expected e2e envelope with ct and pk, got: %s", rr.Body.String())
	}

	// Decrypt and verify
	serverPubBytes, _ := base64.RawURLEncoding.DecodeString(resp.E2E.PK)
	serverPub, _ := ecdh.X25519().NewPublicKey(serverPubBytes)
	shared, _ := clientPriv.ECDH(serverPub)
	key := hkdfSHA256(shared, nil, []byte("diction-transcript-v1"), 32)

	ctBytes, _ := base64.RawURLEncoding.DecodeString(resp.E2E.CT)
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonce := ctBytes[:gcm.NonceSize()]
	plaintext, err := gcm.Open(nil, nonce, ctBytes[gcm.NonceSize():], nil)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(plaintext) != "hello world" {
		t.Errorf("decrypted: want 'hello world', got %q", string(plaintext))
	}
}

func TestExtractFormField_Found(t *testing.T) {
	body, ct := buildMultipart(t, map[string]string{"context": `{"before":"hello","after":"world"}`, "language": "en"}, "audio.wav", "data")
	boundary := parseBoundary(ct)
	got := extractFormField(body, boundary, "context")
	if got != `{"before":"hello","after":"world"}` {
		t.Errorf("context: want JSON, got %q", got)
	}
}

func TestExtractFormField_NotFound(t *testing.T) {
	body, ct := buildMultipart(t, map[string]string{"language": "en"}, "audio.wav", "data")
	boundary := parseBoundary(ct)
	got := extractFormField(body, boundary, "context")
	if got != "" {
		t.Errorf("context: want empty, got %q", got)
	}
}

func TestRewriteMultipart_InjectsForwardModel(t *testing.T) {
	body, ct := buildMultipart(t, map[string]string{"model": "small", "language": "en"}, "audio.m4a", "fake-audio")
	boundary := parseBoundary(ct)

	rewritten, newCT, err := rewriteMultipart(body, boundary, false, "Systran/faster-whisper-large-v3-turbo")
	if err != nil {
		t.Fatalf("rewriteMultipart error: %v", err)
	}

	_, params, _ := parseMediaType(newCT)
	reader := multipart.NewReader(bytes.NewReader(rewritten), params["boundary"])
	var modelValue string
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read part: %v", err)
		}
		if part.FormName() == "model" {
			data, _ := io.ReadAll(part)
			modelValue = string(data)
		}
		part.Close()
	}

	if modelValue != "Systran/faster-whisper-large-v3-turbo" {
		t.Errorf("model: want Systran/faster-whisper-large-v3-turbo, got %q", modelValue)
	}
}

func TestRewriteMultipart_StripsContextField(t *testing.T) {
	body, ct := buildMultipart(t, map[string]string{"context": `{"before":"x"}`, "language": "en"}, "audio.wav", "audio")
	boundary := parseBoundary(ct)
	rewritten, newCT, err := rewriteMultipart(body, boundary, false, "")
	if err != nil {
		t.Fatalf("rewriteMultipart: %v", err)
	}
	_, params, _ := parseMediaType(newCT)
	reader := multipart.NewReader(bytes.NewReader(rewritten), params["boundary"])
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read part: %v", err)
		}
		if part.FormName() == "context" {
			t.Error("context field should have been stripped")
		}
		part.Close()
	}
}

func TestTranscriptionHandler_ContextFieldPassedToPostProcess(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"text":"raw"}`)
	}))
	defer backend.Close()

	g := &Gateway{
		backends: []Backend{
			{Name: "small", URL: backend.URL, Aliases: []string{"small"}},
		},
		health:       newHealthState(),
		defaultModel: "small",
		maxBodySize:  10 * 1024 * 1024,
	}

	var receivedContext string
	postProcess := func(ctx context.Context, text, contextJSON string) (string, error) {
		receivedContext = contextJSON
		return text, nil
	}

	contextJSON := `{"before":"hello ","after":" world","selected":"test"}`
	body, ct := buildMultipart(t, map[string]string{"context": contextJSON}, "audio.wav", "RIFF"+string(make([]byte, 100)))
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions?enhance=true", bytes.NewReader(body))
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	g.TranscriptionHandlerWithPostProcess(postProcess)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d (body: %s)", rr.Code, rr.Body.String())
	}
	if receivedContext != contextJSON {
		t.Errorf("context: want %q, got %q", contextJSON, receivedContext)
	}
}
