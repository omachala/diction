package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// parseBoundary extracts the multipart boundary from the Content-Type header.
func parseBoundary(contentType string) string {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		return ""
	}
	return params["boundary"]
}

// rewriteMultipart rewrites the multipart body:
// - Strips the "model" field (routing is done); if forwardModel is non-empty, injects it for the backend
// - If convertToWAV is true, converts the audio file to 16kHz mono WAV via ffmpeg (for backends that only accept WAV)
func rewriteMultipart(body []byte, boundary string, convertToWAV bool, forwardModel string) ([]byte, string, error) {
	reader := multipart.NewReader(bytes.NewReader(body), boundary)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("read multipart: %w", err)
		}

		// Strip model and context fields - gateway-only, not forwarded to backend
		if part.FormName() == "model" || part.FormName() == "context" {
			part.Close()
			continue
		}

		if part.FormName() == "file" {
			audioData, err := io.ReadAll(part)
			filename := part.FileName()
			part.Close()
			if err != nil {
				return nil, "", fmt.Errorf("read audio: %w", err)
			}

			if convertToWAV {
				audioData, err = convertToWAVBytes(audioData, filename)
				if err != nil {
					return nil, "", err
				}
				filename = "audio.wav"
			}

			partHeader := make(textproto.MIMEHeader)
			partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
			if convertToWAV {
				partHeader.Set("Content-Type", "audio/wav")
			}
			dst, err := writer.CreatePart(partHeader)
			if err != nil {
				return nil, "", fmt.Errorf("create file part: %w", err)
			}
			if _, err := dst.Write(audioData); err != nil {
				return nil, "", fmt.Errorf("write audio: %w", err)
			}
		} else {
			// Copy non-file parts as-is
			data, err := io.ReadAll(part)
			fieldName := part.FormName()
			part.Close()
			if err != nil {
				return nil, "", fmt.Errorf("read part %s: %w", fieldName, err)
			}
			if err := writer.WriteField(fieldName, string(data)); err != nil {
				return nil, "", fmt.Errorf("write field %s: %w", fieldName, err)
			}
		}
	}

	if forwardModel != "" {
		if err := writer.WriteField("model", forwardModel); err != nil {
			return nil, "", fmt.Errorf("write model field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("close multipart writer: %w", err)
	}

	return buf.Bytes(), writer.FormDataContentType(), nil
}

// convertToWAVBytes converts audio bytes to 16kHz mono WAV. Passes through data that's already WAV.
func convertToWAVBytes(audioData []byte, filename string) ([]byte, error) {
	// Skip conversion if already WAV
	if len(audioData) >= 4 && string(audioData[:4]) == "RIFF" {
		return audioData, nil
	}

	tmpDir, err := os.MkdirTemp("", "diction-convert-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	ext := ".m4a"
	if filename != "" {
		if e := filepath.Ext(filename); e != "" {
			ext = e
		}
	}

	inPath := filepath.Join(tmpDir, "input"+ext)
	outPath := filepath.Join(tmpDir, "output.wav")

	if err := os.WriteFile(inPath, audioData, 0644); err != nil {
		return nil, fmt.Errorf("write temp input: %w", err)
	}

	cmd := exec.Command("ffmpeg", "-y", "-i", inPath, "-ar", "16000", "-ac", "1", "-f", "wav", outPath)
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg convert: %w", err)
	}

	wavData, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("read wav output: %w", err)
	}

	return wavData, nil
}

// extractFormField reads a single field value from a multipart body without consuming it.
func extractFormField(body []byte, boundary, fieldName string) string {
	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		part, err := reader.NextPart()
		if err != nil {
			break
		}
		if part.FormName() == fieldName {
			data, _ := io.ReadAll(part)
			part.Close()
			return string(data)
		}
		part.Close()
	}
	return ""
}

// TranscriptionHandler returns the handler for POST /v1/audio/transcriptions.
func (g *Gateway) TranscriptionHandler() http.HandlerFunc {
	return g.TranscriptionHandlerWithPostProcess(nil)
}

// TranscriptionHandlerWithPostProcess is like TranscriptionHandler but calls postProcess
// on the transcript when ?enhance=true is requested. Pass nil for no post-processing.
// postProcess receives (ctx, transcript, contextJSON).
func (g *Gateway) TranscriptionHandlerWithPostProcess(postProcess func(context.Context, string, string) (string, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// Read body (bounded)
		body, err := io.ReadAll(io.LimitReader(r.Body, g.maxBodySize+1))
		if err != nil {
			http.Error(w, `{"error":"failed to read request body"}`, http.StatusBadRequest)
			return
		}
		if int64(len(body)) > g.maxBodySize {
			http.Error(w, fmt.Sprintf(`{"error":"request body exceeds %d bytes"}`, g.maxBodySize), http.StatusRequestEntityTooLarge)
			return
		}

		// Resolve backend
		contentType := r.Header.Get("Content-Type")
		boundary := parseBoundary(contentType)
		target, backend := g.resolveBackend(g.defaultModel)
		if target == nil {
			http.Error(w, `{"error":"backend unavailable"}`, http.StatusBadRequest)
			return
		}

		// Rewrite multipart body: strip model field (routing is done), convert audio if needed
		proxyBody := body
		proxyContentType := contentType
		if boundary != "" {
			converted, newCT, err := rewriteMultipart(body, boundary, backend.NeedsWAV, backend.ForwardModel)
			if err != nil {
				log.Printf("Multipart rewrite failed for %s: %v", backend.Name, err)
				http.Error(w, `{"error":"request processing failed"}`, http.StatusInternalServerError)
				return
			}
			proxyBody = converted
			proxyContentType = newCT
		}

		// Extract context from multipart form field (stripped during rewrite)
		var contextJSON string
		if boundary != "" {
			contextJSON = extractFormField(body, boundary, "context")
		}

		// Capture E2E client key before proxy (X-Diction-E2E header carries client ephemeral X25519 pubkey)
		e2eClientKey := r.Header.Get("X-Diction-E2E")

		// Proxy via httputil.ReverseProxy
		enhanceEnabled := r.URL.Query().Get("enhance") == "true"
		whisperStart := time.Now()
		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.URL.Path = "/v1/audio/transcriptions"
				req.Host = target.Host
				req.Header.Set("Content-Type", proxyContentType)
				req.Body = io.NopCloser(bytes.NewReader(proxyBody))
				req.ContentLength = int64(len(proxyBody))
				if backend.AuthHeader != "" {
					req.Header.Set("Authorization", backend.AuthHeader)
				}
			},
			ModifyResponse: func(resp *http.Response) error {
				whisperMs := time.Since(whisperStart).Milliseconds()
				resp.Header.Set("X-Diction-Whisper-Ms", fmt.Sprintf("%d", whisperMs))

				needsRewrite := (postProcess != nil && enhanceEnabled) || e2eClientKey != ""
				if !needsRewrite || resp.StatusCode != http.StatusOK {
					return nil
				}

				// Read Whisper response body
				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					resp.Body = io.NopCloser(bytes.NewReader(body))
					return nil
				}

				// Parse {"text": "..."}
				var transcription struct {
					Text string `json:"text"`
				}
				if err := json.Unmarshal(body, &transcription); err != nil || transcription.Text == "" {
					resp.Body = io.NopCloser(bytes.NewReader(body))
					return nil
				}

				transcript := transcription.Text

				// LLM post-processing (if requested)
				if postProcess != nil && enhanceEnabled {
					llmStart := time.Now()
					cleaned, err := postProcess(resp.Request.Context(), transcript, contextJSON)
					llmMs := time.Since(llmStart).Milliseconds()
					resp.Header.Set("X-Diction-LLM-Ms", fmt.Sprintf("%d", llmMs))
					if err != nil {
						log.Printf("post-process error (returning raw): %v", err)
					} else {
						transcript = cleaned
					}
				}

				// E2E encrypt transcript if client sent ephemeral pubkey
				if e2eClientKey != "" {
					ct, pk, err := EncryptTranscript(transcript, e2eClientKey)
					if err != nil {
						log.Printf("e2e encrypt error: %v", err)
						// Fall through to plain response
					} else {
						newBody, _ := json.Marshal(map[string]any{
							"e2e": map[string]string{"ct": ct, "pk": pk},
						})
						resp.Body = io.NopCloser(bytes.NewReader(newBody))
						resp.ContentLength = int64(len(newBody))
						resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
						return nil
					}
				}

				// Plain response (no E2E or E2E failed)
				newBody, _ := json.Marshal(map[string]string{"text": transcript})
				resp.Body = io.NopCloser(bytes.NewReader(newBody))
				resp.ContentLength = int64(len(newBody))
				resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
				return nil
			},
			Transport: &http.Transport{
				MaxIdleConns:          20,
				MaxIdleConnsPerHost:   5,
				IdleConnTimeout:       90 * time.Second,
				ResponseHeaderTimeout: 10 * time.Minute,
			},
		}

		proxy.ServeHTTP(w, r)
	}
}
