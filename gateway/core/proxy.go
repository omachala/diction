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
// - Strips the "model" and "context" fields (routing/post-processing done gateway-side)
// - If forwardModel is non-empty, injects it for the backend
// - If whisperPrompt is non-empty, injects it as a "prompt" field for Whisper vocabulary hinting
// - If convertToWAV is true, converts the audio file to 16kHz mono WAV via ffmpeg (for backends that only accept WAV)
func rewriteMultipart(body []byte, boundary string, convertToWAV bool, forwardModel string, whisperPrompt string) ([]byte, string, error) {
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

	if whisperPrompt != "" {
		if err := writer.WriteField("prompt", whisperPrompt); err != nil {
			return nil, "", fmt.Errorf("write prompt field: %w", err)
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

// buildWhisperPrompt extracts custom words from the transcription context JSON and formats
// them as a Whisper prompt field. Whisper uses this as vocabulary hint during transcription,
// improving accuracy for unusual terms, jargon, and proper nouns before LLM cleanup runs.
func buildWhisperPrompt(contextJSON string) string {
	if contextJSON == "" {
		return ""
	}
	var ctx struct {
		CustomWords []struct {
			Word string `json:"word"`
		} `json:"customWords"`
	}
	if err := json.Unmarshal([]byte(contextJSON), &ctx); err != nil || len(ctx.CustomWords) == 0 {
		return ""
	}
	words := make([]string, 0, len(ctx.CustomWords))
	for _, cw := range ctx.CustomWords {
		if cw.Word != "" {
			words = append(words, cw.Word)
		}
	}
	if len(words) == 0 {
		return ""
	}
	return strings.Join(words, ", ")
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
// postProcess receives (ctx, transcript, contextJSON, intent) and returns (resultText, mode, error).
func (g *Gateway) TranscriptionHandlerWithPostProcess(postProcess func(context.Context, string, string, string) (string, string, error)) http.HandlerFunc {
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

		// Resolve backend — route to best model for the request's language
		contentType := r.Header.Get("Content-Type")
		boundary := parseBoundary(contentType)
		language := ""
		if boundary != "" {
			language = extractFormField(body, boundary, "language")
		}
		model := g.ModelForLanguage(language)
		target, backend := g.resolveBackend(model)
		if target == nil {
			http.Error(w, `{"error":"backend unavailable"}`, http.StatusBadRequest)
			return
		}
		if g.fallbackModel != "" {
			log.Printf("Route: language=%s → model=%s", language, model)
		}
		w.Header().Set("X-Diction-Route-Lang", language)
		w.Header().Set("X-Diction-Route-Model", model)

		// Extract context and build Whisper vocabulary prompt from custom words
		var contextJSON string
		var whisperPrompt string
		if boundary != "" {
			contextJSON = extractFormField(body, boundary, "context")
			whisperPrompt = buildWhisperPrompt(contextJSON)
		}

		// Rewrite multipart body: strip model/context fields (routing done), convert audio if needed, inject Whisper prompt
		proxyBody := body
		proxyContentType := contentType
		if boundary != "" {
			converted, newCT, err := rewriteMultipart(body, boundary, backend.NeedsWAV, backend.ForwardModel, whisperPrompt)
			if err != nil {
				log.Printf("Multipart rewrite failed for %s: %v", backend.Name, err)
				http.Error(w, `{"error":"request processing failed"}`, http.StatusInternalServerError)
				return
			}
			proxyBody = converted
			proxyContentType = newCT
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
				path := "/v1/audio/transcriptions"
				if backend.TargetPath != "" {
					path = backend.TargetPath
				}
				req.URL.Path = path
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

				// Backends with a custom TargetPath (e.g. Canary) may return extra fields
				// (e.g. "timestamps":null) that aren't part of the gateway contract — always
				// normalize their response body to {"text":"..."}.
				needsNormalize := backend.TargetPath != ""
				needsRewrite := (postProcess != nil && enhanceEnabled) || e2eClientKey != "" || needsNormalize
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
				if g.OnTranscription != nil {
					g.OnTranscription(backend.Name, whisperMs, len(transcript), enhanceEnabled, e2eClientKey != "")
				}

				// LLM post-processing (if requested)
				var mode string
				if postProcess != nil && enhanceEnabled {
					intent := r.URL.Query().Get("intent")
					llmStart := time.Now()
					resultText, resultMode, err := postProcess(resp.Request.Context(), transcript, contextJSON, intent)
					llmMs := time.Since(llmStart).Milliseconds()
					resp.Header.Set("X-Diction-LLM-Ms", fmt.Sprintf("%d", llmMs))
					if err != nil {
						log.Printf("post-process error (returning raw): %v", err)
					} else {
						transcript = resultText
						mode = resultMode
					}
				}

				// E2E encrypt transcript if client sent ephemeral pubkey
				if e2eClientKey != "" {
					ct, pk, err := EncryptTranscript(transcript, e2eClientKey)
					if err != nil {
						log.Printf("e2e encrypt error: %v", err)
						// Fall through to plain response
					} else {
						result := map[string]any{
							"e2e": map[string]string{"ct": ct, "pk": pk},
						}
						if mode != "" {
							result["mode"] = mode
						}
						newBody, _ := json.Marshal(result)
						resp.Body = io.NopCloser(bytes.NewReader(newBody))
						resp.ContentLength = int64(len(newBody))
						resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
						return nil
					}
				}

				// Plain response (no E2E or E2E failed)
				result := map[string]string{"text": transcript}
				if mode != "" {
					result["mode"] = mode
				}
				newBody, _ := json.Marshal(result)
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
