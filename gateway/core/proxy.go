package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
// - Strips the "model", "context", and "response_format" fields (gateway-owned, not forwarded)
// - If forwardModel is non-empty, injects it for the backend
// - If whisperPrompt is non-empty, injects it as a "prompt" field for Whisper vocabulary hinting
// - If convertToWAV is true, converts the audio file to 16kHz mono WAV via ffmpeg (for backends that only accept WAV)
func rewriteMultipart(body []byte, boundary string, convertToWAV bool, forwardModel string, whisperPrompt string) ([]byte, string, int64, error) {
	reader := multipart.NewReader(bytes.NewReader(body), boundary)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	var audioDurationMs int64

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", 0, fmt.Errorf("read multipart: %w", err)
		}

		// Strip model, context, response_format fields - gateway owns these, not forwarded to backend
		if part.FormName() == "model" || part.FormName() == "context" || part.FormName() == "response_format" {
			part.Close()
			continue
		}

		if part.FormName() == "file" {
			audioData, err := io.ReadAll(part)
			filename := part.FileName()
			part.Close()
			if err != nil {
				return nil, "", 0, fmt.Errorf("read audio: %w", err)
			}

			if convertToWAV {
				audioData, err = convertToWAVBytes(audioData, filename)
				if err != nil {
					return nil, "", 0, err
				}
				filename = "audio.wav"
			}

			audioDurationMs = ParseWAVDurationMs(audioData)

			partHeader := make(textproto.MIMEHeader)
			partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
			if convertToWAV {
				partHeader.Set("Content-Type", "audio/wav")
			}
			dst, err := writer.CreatePart(partHeader)
			if err != nil {
				return nil, "", 0, fmt.Errorf("create file part: %w", err)
			}
			if _, err := dst.Write(audioData); err != nil {
				return nil, "", 0, fmt.Errorf("write audio: %w", err)
			}
		} else {
			// Copy non-file parts as-is
			data, err := io.ReadAll(part)
			fieldName := part.FormName()
			part.Close()
			if err != nil {
				return nil, "", 0, fmt.Errorf("read part %s: %w", fieldName, err)
			}
			if err := writer.WriteField(fieldName, string(data)); err != nil {
				return nil, "", 0, fmt.Errorf("write field %s: %w", fieldName, err)
			}
		}
	}

	if forwardModel != "" {
		if err := writer.WriteField("model", forwardModel); err != nil {
			return nil, "", 0, fmt.Errorf("write model field: %w", err)
		}
	}

	if whisperPrompt != "" {
		if err := writer.WriteField("prompt", whisperPrompt); err != nil {
			return nil, "", 0, fmt.Errorf("write prompt field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", 0, fmt.Errorf("close multipart writer: %w", err)
	}

	return buf.Bytes(), writer.FormDataContentType(), audioDurationMs, nil
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

// writeTranscriptionResponse rewrites the proxied response body into the shape
// requested by the client. Called only on a successful (2xx) backend response;
// backend errors pass through untouched in whatever shape the backend produced.
//
// responseFormat must be "json" (default) or "text" (OpenAI-compatible plain body).
// e2eClientKey is the raw X-Diction-E2E header — when set, JSON output is encrypted.
// "text" + e2eClientKey != "" is rejected upstream in the handler.
func writeTranscriptionResponse(resp *http.Response, transcript, mode, responseFormat, e2eClientKey string) {
	// Plain text — OpenAI response_format=text. Only reachable when e2eClientKey == "".
	if responseFormat == "text" {
		resp.Header.Set("Content-Type", "text/plain; charset=utf-8")
		body := []byte(transcript)
		resp.Body = io.NopCloser(bytes.NewReader(body))
		resp.ContentLength = int64(len(body))
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
		return
	}

	// E2E JSON — wrap transcript in encrypted envelope.
	if e2eClientKey != "" {
		ct, pk, err := EncryptTranscript(transcript, e2eClientKey)
		if err == nil {
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
			return
		}
		log.Printf("e2e encrypt error: %v", err)
		if OnError != nil {
			OnError(resp.Request.Context(), ErrorEvent{
				Source:     "e2e",
				Kind:       "e2e_encrypt",
				Endpoint:   "/v1/audio/transcriptions",
				HTTPStatus: resp.StatusCode,
				Hint:       "transcript e2e encrypt failed; falling back to plain",
			})
		}
		// fall through to plain JSON
	}

	// Plain JSON (default).
	result := map[string]string{"text": transcript}
	if mode != "" {
		result["mode"] = mode
	}
	newBody, _ := json.Marshal(result)
	resp.Body = io.NopCloser(bytes.NewReader(newBody))
	resp.ContentLength = int64(len(newBody))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
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
		responseFormat := "json"
		if boundary != "" {
			language = extractFormField(body, boundary, "language")
			if rf := extractFormField(body, boundary, "response_format"); rf != "" {
				responseFormat = rf
			}
		}

		// Only OpenAI formats we can actually produce. verbose_json/srt/vtt require
		// word timestamps which aren't universally available; fail loud rather than
		// silently returning a different shape than the client asked for.
		switch responseFormat {
		case "json", "text":
		default:
			http.Error(w, fmt.Sprintf(`{"error":"response_format '%s' not supported; gateway supports 'json' and 'text' only"}`, responseFormat), http.StatusBadRequest)
			return
		}

		// response_format=text returns a raw string body — no room for the E2E envelope.
		// iOS never sends response_format; Speaches clients never send X-Diction-E2E;
		// this is a configuration error, not a silent-fallback situation.
		if responseFormat == "text" && r.Header.Get("X-Diction-E2E") != "" {
			http.Error(w, `{"error":"response_format=text is incompatible with X-Diction-E2E (E2E requires JSON envelope)"}`, http.StatusBadRequest)
			return
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
		var audioDurationMs int64
		if boundary != "" {
			converted, newCT, durationMs, err := rewriteMultipart(body, boundary, backend.NeedsWAV, backend.ForwardModel, whisperPrompt)
			if err != nil {
				log.Printf("Multipart rewrite failed for %s: %v", backend.Name, err)
				if OnError != nil {
					OnError(r.Context(), ErrorEvent{
						Source:     "stt",
						Kind:       "stt_multipart",
						Endpoint:   "/v1/audio/transcriptions",
						Provider:   backend.Name,
						HTTPStatus: http.StatusInternalServerError,
						Hint:       "multipart rewrite failed",
					})
				}
				http.Error(w, `{"error":"request processing failed"}`, http.StatusInternalServerError)
				return
			}
			proxyBody = converted
			proxyContentType = newCT
			audioDurationMs = durationMs
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

				if resp.StatusCode != http.StatusOK {
					return nil
				}

				// Always read the response body to extract the transcript text for metrics,
				// even when no rewrite is needed. The body is restored below.
				respBody, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					resp.Body = io.NopCloser(bytes.NewReader(respBody))
					return nil
				}

				// Parse {"text": "..."}
				var transcription struct {
					Text string `json:"text"`
				}
				if err := json.Unmarshal(respBody, &transcription); err != nil || transcription.Text == "" {
					resp.Body = io.NopCloser(bytes.NewReader(respBody))
					return nil
				}

				transcript := transcription.Text
				if g.OnTranscription != nil {
					g.OnTranscription(resp.Request.Context(), backend.Name, whisperMs, len(transcript), audioDurationMs, enhanceEnabled, e2eClientKey != "")
				}

				// Backends with a custom TargetPath (e.g. Canary) may return extra fields
				// (e.g. "timestamps":null) that aren't part of the gateway contract — always
				// normalize their response body to {"text":"..."}.
				needsNormalize := backend.TargetPath != ""
				needsRewrite := (postProcess != nil && enhanceEnabled) || e2eClientKey != "" || needsNormalize || responseFormat != "json"
				if !needsRewrite {
					resp.Body = io.NopCloser(bytes.NewReader(respBody))
					return nil
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
						if OnError != nil {
							OnError(resp.Request.Context(), ErrorEvent{
								Source:     "stt",
								Kind:       "stt_post_process",
								Endpoint:   "/v1/audio/transcriptions",
								Provider:   backend.Name,
								HTTPStatus: resp.StatusCode,
								InputChars: len(transcript),
								LatencyMs:  time.Since(llmStart).Milliseconds(),
								Hint:       "post-process failed; returning raw transcript",
							})
						}
					} else {
						transcript = resultText
						mode = resultMode
					}
				}

				writeTranscriptionResponse(resp, transcript, mode, responseFormat, e2eClientKey)
				return nil
			},
			Transport: &http.Transport{
				MaxIdleConns:          20,
				MaxIdleConnsPerHost:   5,
				IdleConnTimeout:       90 * time.Second,
				ResponseHeaderTimeout: 10 * time.Minute,
			},
			ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
				// Distinguish upstream-canceled (user started a new recording
				// mid-flight, client disconnect) from genuine transport
				// failures. Without this hook these failures silently write
				// a 502 and leave no `errors` point — confirmed 30d sweep
				// showed zero error points matched against `http: proxy
				// error: context canceled` log lines.
				kind := "stt_backend_error"
				hint := "backend transport failed"
				if errors.Is(err, context.Canceled) {
					kind = "stt_upstream_canceled"
					hint = "upstream canceled (client disconnect or new request)"
				}
				log.Printf("http: proxy error: %v", err)
				if OnError != nil {
					OnError(req.Context(), ErrorEvent{
						Source:     "stt",
						Kind:       kind,
						Endpoint:   "/v1/audio/transcriptions",
						Provider:   backend.Name,
						HTTPStatus: http.StatusBadGateway,
						Hint:       hint,
					})
				}
				if OnRequestFailed != nil {
					OnRequestFailed(req.Context(), errTypeSTTError)
				}
				rw.WriteHeader(http.StatusBadGateway)
			},
		}

		proxy.ServeHTTP(w, r)
	}
}
