---
title: "Self-Hosting Setup Guide"
description: Run your own Whisper speech-to-text server and connect Diction to it. Two setup paths, direct or with the Diction streaming gateway. Works on any machine with Docker.
---

<img src="/illustration-self-hosting-setup.svg" alt="Controller" class="illustration" style="max-width: 480px; margin: 0 auto 2rem; display: block;" />

# Self-Hosting Setup Guide

Run your own Whisper server, point the Diction app at it, start dictating. Your audio never touches our infrastructure.

Diction speaks the OpenAI transcription API (`POST /v1/audio/transcriptions`). Any server that implements it works. You have two ways to set it up, depending on how much you care about latency.

## Path 1: Whisper only (simplest)

The minimal setup. One container. No gateway, no extra moving parts.

```yaml
# docker-compose.yml
services:
  whisper:
    image: fedirz/faster-whisper-server:latest-cpu
    ports:
      - "8000:8000"
    environment:
      WHISPER__MODEL: Systran/faster-whisper-small
      WHISPER__INFERENCE_DEVICE: cpu
```

```bash
docker compose up -d
```

Open the Diction app, switch to **Self-Hosted**, paste `http://your-server:8000`. A green dot confirms the endpoint is reachable. Start dictating.

**The trade-off:** no streaming. The app waits until you stop speaking, uploads the whole recording to your server, and waits for Whisper to transcribe it. On short phrases that's fine. On longer dictations you'll see a visible pause after you tap stop.

If that's acceptable, you're done. Skip to [Choosing a model](#choosing-a-model).

## Path 2: Whisper + the Diction gateway (streaming)

Adds our open-source gateway in front of whisper. The gateway exposes a WebSocket endpoint the Diction app uses to stream audio live as you speak. By the time you stop talking, the transcript is mostly ready.

```yaml
# docker-compose.yml
services:
  gateway:
    image: ghcr.io/omachala/diction-gateway:latest
    ports:
      - "8080:8080"
    environment:
      DEFAULT_MODEL: small
    depends_on:
      - whisper-small

  whisper-small:
    image: fedirz/faster-whisper-server:latest-cpu
    environment:
      WHISPER__MODEL: Systran/faster-whisper-small
      WHISPER__INFERENCE_DEVICE: cpu
```

```bash
docker compose up -d
```

Paste `http://your-server:8080` into the Diction app's **Self-Hosted** tab. Short phrases feel about the same as Path 1. Longer dictations are noticeably faster. The longer you talk, the bigger the gap.

The Diction gateway is fully open source. It runs as a pure proxy and streaming layer. It does not talk to our servers, does not require a subscription, and does not send any telemetry.

## Choosing a model

Both paths support any model. Pick based on your hardware and what you're dictating.

| Model ID | Params | RAM | Notes |
|----------|--------|-----|-------|
| `Systran/faster-whisper-small` | 244M | ~850 MB | Recommended starting point. Fast on CPU, fine for most dictations. |
| `Systran/faster-whisper-medium` | 769M | ~2.1 GB | Better with accents and background noise. Slow on CPU, good on GPU. |
| `deepdml/faster-whisper-large-v3-turbo-ct2` | 809M | ~2.3 GB | Highest accuracy. Manageable on modern CPUs, near-instant on GPU. |

Swap the model by changing `WHISPER__MODEL` in the service. For Path 2 (gateway), also update `DEFAULT_MODEL` on the gateway service and make sure the whisper service is named to match: `whisper-small`, `whisper-medium`, or `whisper-large-turbo`.

You can run multiple models at the same time in the same compose file. The gateway will route to whichever one you set as the default.

## Connecting the app

1. Open the Diction app
2. Switch to the **Self-Hosted** tab
3. Paste your server URL into **Endpoint URL**:

```
http://192.168.1.100:8080
```

Replace the address with your server's actual IP. A green dot next to the endpoint confirms it's reachable. Tap the mic and start dictating.

## No public IP?

You don't need to open ports on your router. Several free options connect your phone to a home server from anywhere:

- **[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/)**. Free, outbound-only connection. No port forwarding.
- **[Tailscale](https://tailscale.com/)**. Free WireGuard mesh VPN. Install on server and phone, connect from anywhere.
- **[ngrok](https://ngrok.com/)**. Instant public URL. Great for quick testing.

## Optional: API key

If your server is behind an API key (common with reverse proxies or hosted endpoints), enter it in the **API Key** field in the app's Self-Hosted settings. It's sent as a `Bearer` token with every request.

## Any Whisper endpoint works

Neither path locks you to our containers. Both the Diction app (Path 1) and the gateway (Path 2) talk the standard OpenAI transcription API. Anything that accepts `POST /v1/audio/transcriptions` with a file upload and returns a JSON transcript works:

- [faster-whisper-server](https://github.com/fedirz/faster-whisper-server) (used in both paths above)
- [whisper.cpp](https://github.com/ggerganov/whisper.cpp) HTTP server
- OpenAI's own Whisper API
- Any future model that speaks the same protocol

Already running one? See [Use Your Own Model](/features/custom-model).

## Requirements

- Any machine that runs Docker (home server, NAS, cloud VM, Raspberry Pi for tiny models)
- iPhone on the same network, or reachable via tunnel or VPN

## Full configuration

The complete compose file with multiple model profiles, and all gateway environment variables, is in the [public GitHub repository](https://github.com/omachala/diction).
