package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coder/websocket"
)

const (
	streamTimeout   = 90 * time.Minute
	wsCloseUnknown  = 4000
	wsCloseDown     = 4001
	wsCloseFailed   = 4002
	wsCloseTooLarge = 4003
	wsCloseNoAudio  = 4004
)

type streamAction struct {
	Action   string `json:"action"`
	Language string `json:"language,omitempty"`
}

type streamResult struct {
	Text string `json:"text"`
}

// StreamingHandler returns the handler for WS /v1/audio/stream.
//
// Protocol:
//
//	Client connects: ws(s)://host/v1/audio/stream?model=small&language=en
//	Client → Server: binary frames of PCM audio (16-bit LE, mono, 16kHz)
//	Client → Server: text frame {"action":"done"}
//	Server → Client: text frame {"text":"transcribed text"}
//	Server closes connection.
func (g *Gateway) StreamingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate model before upgrade
		model := r.URL.Query().Get("model")
		if model == "" {
			model = g.defaultModel
		}
		target, _ := g.resolveBackend(model)
		if target == nil {
			http.Error(w, fmt.Sprintf(`{"error":"unknown model: %s"}`, model), http.StatusBadRequest)
			return
		}
		if !g.health.get(strings.TrimSpace(model)) {
			// Try matching by iterating backends
			backendUp := false
			for _, b := range g.backends {
				for _, alias := range b.Aliases {
					if strings.EqualFold(model, alias) {
						backendUp = g.health.get(b.Name)
						break
					}
				}
				if backendUp {
					break
				}
			}
			if !backendUp {
				http.Error(w, `{"error":"backend unavailable"}`, http.StatusServiceUnavailable)
				return
			}
		}

		language := r.URL.Query().Get("language")

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // allow any origin for now
		})
		if err != nil {
			log.Printf("ws accept: %v", err)
			return
		}
		defer conn.CloseNow()

		ctx, cancel := context.WithTimeout(r.Context(), streamTimeout)
		defer cancel()

		// Collect PCM chunks
		var pcmBuf bytes.Buffer
		maxPCM := g.maxBodySize

		for {
			msgType, data, err := conn.Read(ctx)
			if err != nil {
				log.Printf("ws read: %v", err)
				return
			}

			if msgType == websocket.MessageBinary {
				if int64(pcmBuf.Len())+int64(len(data)) > maxPCM {
					conn.Close(wsCloseTooLarge, "audio too large")
					return
				}
				pcmBuf.Write(data)
				continue
			}

			// Text message — check for done action
			if msgType == websocket.MessageText {
				var action streamAction
				if err := json.Unmarshal(data, &action); err != nil {
					continue
				}
				if action.Action == "done" {
					if action.Language != "" {
						language = action.Language
					}
					break
				}
			}
		}

		if pcmBuf.Len() == 0 {
			conn.Close(wsCloseNoAudio, "no audio received")
			return
		}

		// Wrap PCM in WAV header and POST to backend
		text, err := g.proxyToBackend(ctx, target, pcmBuf.Bytes(), model, language)
		if err != nil {
			log.Printf("ws proxy: %v", err)
			conn.Close(wsCloseFailed, "transcription failed")
			return
		}

		result, _ := json.Marshal(streamResult{Text: text})
		if err := conn.Write(ctx, websocket.MessageText, result); err != nil {
			log.Printf("ws write result: %v", err)
			return
		}

		conn.Close(websocket.StatusNormalClosure, "")
	}
}

// proxyToBackend wraps raw PCM in a WAV and POSTs multipart to the whisper backend.
func (g *Gateway) proxyToBackend(ctx context.Context, target *url.URL, pcm []byte, model, language string) (string, error) {
	// Build WAV
	var wav bytes.Buffer
	if err := WriteWAVHeader(&wav, len(pcm)); err != nil {
		return "", fmt.Errorf("write wav header: %w", err)
	}
	wav.Write(pcm)

	// Build multipart body
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	filePart, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(filePart, &wav); err != nil {
		return "", fmt.Errorf("copy wav: %w", err)
	}

	writer.WriteField("model", model)
	if language != "" {
		writer.WriteField("language", language)
	}
	writer.Close()

	// POST to backend
	backendURL := target.String() + "/v1/audio/transcriptions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, backendURL, &body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("backend request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("backend returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Text, nil
}
