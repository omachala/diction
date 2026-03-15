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
// - Strips the "model" field (each backend container serves one model, the field is only used for gateway routing)
// - If convertToWAV is true, converts the audio file to 16kHz mono WAV via ffmpeg (for backends that only accept WAV)
func rewriteMultipart(body []byte, boundary string, convertToWAV bool) ([]byte, string, error) {
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

		// Strip model field — gateway already routed to the correct backend
		if part.FormName() == "model" {
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

// TranscriptionHandler returns the handler for POST /v1/audio/transcriptions.
func (g *Gateway) TranscriptionHandler() http.HandlerFunc {
	return g.TranscriptionHandlerWithPostProcess(nil)
}

// TranscriptionHandlerWithPostProcess is like TranscriptionHandler but calls postProcess
// on the transcript when ?enhance=true is requested. Pass nil for no post-processing.
func (g *Gateway) TranscriptionHandlerWithPostProcess(postProcess func(context.Context, string) (string, error)) http.HandlerFunc {
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
			converted, newCT, err := rewriteMultipart(body, boundary, backend.NeedsWAV)
			if err != nil {
				log.Printf("Multipart rewrite failed for %s: %v", backend.Name, err)
				http.Error(w, `{"error":"request processing failed"}`, http.StatusInternalServerError)
				return
			}
			proxyBody = converted
			proxyContentType = newCT
		}

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
			},
			ModifyResponse: func(resp *http.Response) error {
				whisperMs := time.Since(whisperStart).Milliseconds()
				resp.Header.Set("X-Diction-Whisper-Ms", fmt.Sprintf("%d", whisperMs))

				if postProcess == nil || !enhanceEnabled || resp.StatusCode != http.StatusOK {
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

				// Post-process transcript
				llmStart := time.Now()
				cleaned, err := postProcess(resp.Request.Context(), transcription.Text)
				llmMs := time.Since(llmStart).Milliseconds()
				resp.Header.Set("X-Diction-LLM-Ms", fmt.Sprintf("%d", llmMs))

				if err != nil {
					log.Printf("post-process error (returning raw): %v", err)
					resp.Body = io.NopCloser(bytes.NewReader(body))
					return nil
				}

				// Rewrite response with cleaned text
				newBody, _ := json.Marshal(map[string]string{"text": cleaned})
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
