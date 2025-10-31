package core

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

// TranscriptionHandler returns the handler for POST /v1/audio/transcriptions.
func (g *Gateway) TranscriptionHandler() http.HandlerFunc {
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

		// Extract model from multipart
		model := g.defaultModel
		contentType := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(contentType)
		if err == nil && strings.HasPrefix(mediaType, "multipart/") {
			boundary := params["boundary"]
			if boundary != "" {
				reader := multipart.NewReader(bytes.NewReader(body), boundary)
				for {
					part, err := reader.NextPart()
					if err != nil {
						break
					}
					if part.FormName() == "model" {
						val, err := io.ReadAll(part)
						if err == nil && len(val) > 0 {
							model = string(val)
						}
					}
					part.Close()
				}
			}
		}

		// Resolve backend
		target, ok := g.resolveBackend(model)
		if !ok {
			http.Error(w, fmt.Sprintf(`{"error":"unknown model: %s"}`, model), http.StatusBadRequest)
			return
		}

		// Proxy via httputil.ReverseProxy
		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.URL.Path = "/v1/audio/transcriptions"
				req.Host = target.Host
				req.Body = io.NopCloser(bytes.NewReader(body))
				req.ContentLength = int64(len(body))
			},
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		}

		proxy.ServeHTTP(w, r)
	}
}
