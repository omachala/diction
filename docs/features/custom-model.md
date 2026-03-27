---
title: Use Your Own Model
description: Already running a speech-to-text server? Connect Diction to it. One env var, no extra containers needed. Works with any OpenAI-compatible speech endpoint.
---

<img src="/illustration-custom-model.svg" alt="AI brain intelligence" class="illustration" style="max-width: 480px; margin: 0 auto 2rem; display: block;" />

# Use Your Own Model

Say you already have a speech model running on your homelab. A beefy GPU machine, a fine-tuned model for your language, a domain-specific model, or just a newer version than what ships in our stack. You found Diction, thought "this is exactly what I need," pointed the app at your server, and got a server error.

Here's what's happening and how to fix it.

## Why direct connection doesn't work

The Diction app doesn't talk directly to your speech server. It connects to the gateway first, and the gateway handles everything from there.

The reason is streaming. The app sends audio over a WebSocket in real-time as you speak, and the gateway forwards it to your model when you're done. This is what makes the latency feel instant. Without it, the app would have to record everything you said, then send one big file and wait. You'd feel the pause. With streaming, the transcription comes back before you've finished speaking.

The WebSocket endpoint is part of the gateway. No gateway, no streaming. That's the error you're seeing.

The good news: you don't need to run any of our model containers. Just the gateway, pointed at yours.

## The setup

If you're already running a speech server at, say, `http://192.168.1.50:8000`, your entire `docker-compose.yml` is this:

```yaml
services:
  gateway:
    image: ghcr.io/omachala/diction-gateway:latest
    ports:
      - "8080:8080"
    environment:
      CUSTOM_BACKEND_URL: http://192.168.1.50:8000
      CUSTOM_BACKEND_MODEL: Systran/faster-whisper-large-v3-turbo
```

Start it:

```bash
docker compose up -d
```

Then open the Diction app, go to **Self-Hosted**, and paste your gateway's address:

```
http://192.168.1.50:8080
```

That's it. The gateway connects to your model, streaming works, and text appears as you speak.

## Two common scenarios

### Your server exposes an OpenAI-compatible API

Most self-hosted speech servers speak the same API: `POST /v1/audio/transcriptions`, multipart file upload, JSON response. If your server does this, it works. Point the gateway at it and set the model name your server expects.

Servers like [speaches](https://github.com/speaches-ai/speaches) or [faster-whisper-server](https://github.com/fedirz/faster-whisper-server) are examples, but this applies to any server that follows the same convention.

```yaml
services:
  gateway:
    image: ghcr.io/omachala/diction-gateway:latest
    ports:
      - "8080:8080"
    environment:
      CUSTOM_BACKEND_URL: http://my-server:8000
      CUSTOM_BACKEND_MODEL: your-model-name-here
```

Set `CUSTOM_BACKEND_MODEL` to whatever model name your server expects. If your server only runs one model and doesn't care what name it receives, you can omit it.

### Your model only accepts WAV audio

Some speech models only accept WAV audio input. The gateway can convert for you automatically:

```yaml
services:
  gateway:
    image: ghcr.io/omachala/diction-gateway:latest
    ports:
      - "8080:8080"
    environment:
      CUSTOM_BACKEND_URL: http://my-model:5092
      CUSTOM_BACKEND_NEEDS_WAV: "true"
```

With `CUSTOM_BACKEND_NEEDS_WAV` set, the gateway converts the audio to 16kHz mono WAV before forwarding. Your model gets what it needs.

### Your server requires authentication

If your server is behind an API key:

```yaml
environment:
  CUSTOM_BACKEND_URL: http://my-server:8000
  CUSTOM_BACKEND_MODEL: my-model
  CUSTOM_BACKEND_AUTH: "Bearer sk-your-key-here"
```

The gateway injects the `Authorization` header on every request to your backend.

## All available options

| Variable | Required | Description |
|----------|----------|-------------|
| `CUSTOM_BACKEND_URL` | Yes | Base URL of your server, e.g. `http://192.168.1.50:8000` |
| `CUSTOM_BACKEND_MODEL` | No | Model name to send in the request. Set this if your server has multiple models or requires the field. |
| `CUSTOM_BACKEND_NEEDS_WAV` | No | Set to `true` if your server only accepts WAV audio. Default: `false`. |
| `CUSTOM_BACKEND_AUTH` | No | Authorization header value, e.g. `Bearer sk-xxx`. |

## Latency on a local server

With your model on the same network, the round trip is fast. The streaming setup means words start coming back before you've finished speaking, so even larger, slower models feel responsive. The bottleneck is your hardware, not the connection.

If your server has a dedicated GPU, expect near-instant results regardless of model size.

## What about the built-in models?

Setting `CUSTOM_BACKEND_URL` makes your custom backend the default. The other models in the gateway's built-in stack (small, medium, large) are still registered but won't be running unless you also add their containers. You can mix and match:

```yaml
services:
  gateway:
    image: ghcr.io/omachala/diction-gateway:latest
    ports:
      - "8080:8080"
    environment:
      CUSTOM_BACKEND_URL: http://my-big-model:8000
      CUSTOM_BACKEND_MODEL: Systran/faster-whisper-large-v3-turbo

  whisper-small:
    image: fedirz/faster-whisper-server:latest-cpu
    environment:
      WHISPER__MODEL: Systran/faster-whisper-small
```

Your custom model is the default. The built-in small model is there as a fallback or for testing.

## Requirements

- Docker on any machine
- Your existing speech model reachable from the gateway (same host, same network, or via URL)
- iPhone reachable to the gateway. See [No Public IP?](/features/self-hosting-setup#no-public-ip) if you need a tunnel.
