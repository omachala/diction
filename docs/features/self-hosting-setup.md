---
title: "Self-Hosting Setup Guide"
description: Run your own Whisper speech-to-text server and connect Diction to it. Three commands to start. Works with any OpenAI-compatible endpoint.
---

<img src="/illustration-self-hosting-setup.svg" alt="Controller" class="illustration" style="max-width: 480px; margin: 0 auto 2rem; display: block;" />

# Self-Hosting Setup Guide

Run a Whisper server on your own hardware. Your audio stays on your network. Three commands to start, works on any machine that runs Docker.

## Quick Start

```bash
git clone https://github.com/omachala/diction.git
cd diction
docker compose up -d whisper-small
```

That's it. Whisper is now running at `http://<your-server-ip>:9002`.

Open the Diction app, switch to **Self-Hosted**, paste the URL, and start dictating.

## Choosing a Model

The Docker Compose setup includes several models at different sizes. Pick the one that fits your hardware:

| Model | Port | RAM | Speed | Best for |
|-------|------|-----|-------|----------|
| `whisper-tiny` | 9001 | ~350 MB | ~1-2s | Low-power hardware, quick tests |
| `whisper-small` | 9002 | ~800 MB | ~3-4s | Recommended starting point |
| `whisper-medium` | 9003 | ~1.8 GB | ~8-12s | Better accuracy, needs more RAM |
| `whisper-large` | 9004 | ~3.5 GB | ~20-30s | Best accuracy, needs serious hardware |
| `whisper-distil-large` | 9005 | ~2 GB | ~4-6s | Near-large accuracy, much faster |

Start any model with:

```bash
docker compose up -d whisper-small    # or whisper-tiny, whisper-medium, etc.
```

You can run multiple models at the same time on different ports.

## Connecting the App

1. Open the Diction app
2. Switch to the **Self-Hosted** tab
3. Paste your server URL into the **Endpoint URL** field:

```
http://192.168.1.100:9002
```

Replace `192.168.1.100` with your server's actual IP address. A green dot in the app confirms the endpoint is reachable.

## No Public IP?

You do not need to open ports on your router. Several free options let you connect your phone to a home server from anywhere:

- **[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/)** -- free, outbound-only connection. No port forwarding needed.
- **[Tailscale](https://tailscale.com/)** -- free WireGuard mesh VPN. Install on server and phone, connect from anywhere.
- **[ngrok](https://ngrok.com/)** -- instant public URL, useful for quick testing.

## Any Whisper Endpoint Works

Diction is not locked to our Docker setup. It works with any [OpenAI-compatible](https://platform.openai.com/docs/api-reference/audio/createTranscription) speech-to-text endpoint:

- [faster-whisper-server](https://github.com/fedirz/faster-whisper-server) (what the Docker Compose setup uses)
- [whisper.cpp](https://github.com/ggerganov/whisper.cpp) with the HTTP server
- OpenAI's own Whisper API
- Any future model that speaks the same protocol

If it accepts `POST /v1/audio/transcriptions` with a file upload and returns a JSON transcript, Diction can use it.

## Optional: API Key

If your server is behind an API key (common with reverse proxies or hosted endpoints), enter it in the **API Key** field in the Self-Hosted settings. It is sent as a Bearer token with every request.

## Requirements

- Any machine that can run Docker (home server, NAS, cloud VM, Raspberry Pi for tiny models)
- iPhone on the same network, or reachable via tunnel/VPN

## Already running your own model?

If you have a Whisper-compatible server running separately and want to connect Diction to it without spinning up any of the containers above, see [Use Your Own Model](/features/custom-model).

## Full Documentation

The complete Docker Compose configuration, model details, and advanced setup options are in the [GitHub repository](https://github.com/omachala/diction).
