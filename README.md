<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/logo-light.png">
    <source media="(prefers-color-scheme: light)" srcset="assets/logo-dark.png">
    <img src="assets/logo-dark.png" alt="Diction" height="80">
  </picture>
  <br><br>
  <strong>Speech-to-text keyboard for iOS.</strong><br>Self-hosted or cloud. Open source gateway. No lock-in.
</p>

<p align="center">
  <!-- <a href="https://apps.apple.com/app/diction/id000000000"><img src="https://developer.apple.com/assets/elements/badges/download-on-the-app-store.svg" alt="Download on the App Store" height="40"></a> -->
  <a href="https://diction.one">Website</a> &bull;
  <a href="docs/privacy.md">Privacy Policy</a> &bull;
  <a href="https://github.com/omachala/diction/issues">Report a Bug</a>
</p>

<p align="center">
  <img src="https://img.shields.io/github/license/omachala/diction" alt="License">
  <!-- <img src="https://img.shields.io/github/v/release/omachala/diction" alt="Release"> -->
</p>

---

<p align="center">
  <img src="assets/screenshot-settings.png" width="280" alt="Diction settings screen">&nbsp;&nbsp;
  <img src="assets/screenshot-recording.png" width="280" alt="Diction recording screen">
</p>

## What is Diction

An iOS keyboard that transcribes speech to text. Switch to it in any app, tap the mic, speak, and the text is inserted. No QWERTY — dictation only.

It works in two modes:

- **Diction Cloud** — zero setup. Download the app, start dictating. Powered by a hosted API.
- **Self-Hosted** — run your own transcription server. Free, forever. Your audio never leaves your network.

> **Think of it like [Bitwarden](https://bitwarden.com)** — free and self-hosted for those who want control, with a cloud option for convenience.

The iOS app is pure Swift with zero third-party SDKs. No analytics, no tracking, no telemetry. The self-hosting infrastructure — a Go gateway and Docker Compose setup — is open source and lives in this repo.

## Getting Started

### Diction Cloud

1. Download Diction from the App Store
2. Go to **Settings → General → Keyboard → Keyboards → Add New Keyboard → Diction**
3. Enable **Allow Full Access** (required by iOS for network access — [why?](docs/privacy.md#keyboard-extension--full-access))
4. Open any app, tap 🌐 to switch to Diction, tap the mic, speak

That's it. No server, no configuration.

### Self-Hosted

Run a Whisper-compatible transcription server on any machine with Docker:

```bash
git clone https://github.com/omachala/diction.git
cd diction
docker compose up -d gateway whisper-small
```

Your gateway is now running at `http://<your-server-ip>:9000`.

Then in the Diction app: **Settings → Self-Hosted** → set the endpoint URL to your gateway address (e.g. `http://192.168.1.100:9000`). Pick a model and language, and you're done.

## Models

Diction is model-agnostic. It works with **any [OpenAI-compatible](https://platform.openai.com/docs/api-reference/audio/createTranscription) speech-to-text endpoint**. This repo includes a Docker Compose setup with popular models to get you started:

| Service | Model | Port | RAM | Latency (CPU) | Best for |
|---------|-------|------|-----|---------------|----------|
| `whisper-tiny` | [Whisper Tiny](https://huggingface.co/Systran/faster-whisper-tiny) | 9001 | ~350 MB | ~1-2s | Low-power devices, quick notes |
| **`whisper-small`** | **[Whisper Small](https://huggingface.co/Systran/faster-whisper-small)** | **9002** | **~800 MB** | **~3-4s** | **Best starting point for most users** |
| `whisper-medium` | [Whisper Medium](https://huggingface.co/Systran/faster-whisper-medium) | 9003 | ~1.8 GB | ~8-12s | Accents, background noise |
| `whisper-large` | [Whisper Large V3](https://huggingface.co/Systran/faster-whisper-large-v3) | 9004 | ~3.5 GB | ~20-30s | Maximum Whisper accuracy |
| `whisper-distil-large` | [Distil Whisper Large V3](https://huggingface.co/Systran/faster-distil-whisper-large-v3) | 9005 | ~2 GB | ~4-6s | Near-best quality, much faster (English only) |
| `parakeet` | [NVIDIA Parakeet TDT 0.6B](https://huggingface.co/nvidia/parakeet-tdt-0.6b-v2) | 9006 | ~2 GB | ~1-2s | Best speed + accuracy, 25 European languages |

Start any combination:

```bash
# Just one model (gateway optional but recommended)
docker compose up -d gateway whisper-small

# Multiple models — switch between them in the app
docker compose up -d gateway whisper-small parakeet

# Skip the gateway — connect directly to a model
docker compose up -d whisper-small
# Then use http://<ip>:9002 as your endpoint
```

Models download on first start and are cached in a shared Docker volume — subsequent starts are instant. Parakeet models are baked into the image.

You can also point Diction at anything else: [whisper.cpp](https://github.com/ggerganov/whisper.cpp), [OpenAI's API](https://platform.openai.com/docs/api-reference/audio), a custom fine-tuned model, or any future model. If it has a `/v1/audio/transcriptions` endpoint, Diction works with it.

## Gateway

The gateway is a lightweight Go service (~15 MB Docker image) that sits in front of your model backends:

- **Model routing** — one URL, multiple models. Switch from the app without reconfiguring your server.
- **WebSocket streaming** — audio streams to the server during recording. Transcription starts instantly when you stop — no upload wait.
- **Format conversion** — automatically converts audio to the format each backend needs (e.g. WAV for Parakeet). You don't need to think about it.
- **Health monitoring** — checks each backend every 30s. `GET /v1/models` shows which are online.

The gateway is optional. You can always point the app directly at a model backend. But it's recommended — especially if you run multiple models.

### API

The gateway exposes an [OpenAI-compatible](https://platform.openai.com/docs/api-reference/audio/createTranscription) API:

```bash
# Health check
curl http://localhost:9000/health

# List available models with health status
curl http://localhost:9000/v1/models

# Transcribe audio
curl -X POST http://localhost:9000/v1/audio/transcriptions \
  -F file=@recording.wav \
  -F model=small
```

WebSocket streaming for real-time transcription:

```
WS /v1/audio/stream?model=small&language=en

1. Client sends binary frames: raw PCM audio (16-bit LE, mono, 16kHz)
2. Client sends text frame: {"action":"done"}
3. Server replies: {"text":"transcribed text"}
```

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GATEWAY_PORT` | `8080` | Port the gateway listens on (mapped to 9000 in Docker Compose) |
| `DEFAULT_MODEL` | `small` | Model used when no `model` field is specified |
| `MAX_BODY_SIZE` | `10485760` | Max upload size in bytes (10 MB) |

## Remote Access

Your phone needs to reach the server. On the same Wi-Fi network, use the local IP directly. For access from anywhere:

**[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/)** (recommended) — free, outbound-only. No port forwarding, no public IP needed.

```bash
cloudflared tunnel create diction
cloudflared tunnel route dns diction whisper.yourdomain.com
cloudflared tunnel run --url http://localhost:9000 diction
```

**[Tailscale](https://tailscale.com/)** — free WireGuard mesh VPN. Install on server + iPhone, get a stable `100.x.y.z` IP.

**Reverse proxy** — put the gateway behind [Caddy](https://caddyserver.com) for HTTPS:

```
whisper.yourdomain.com {
    reverse_proxy localhost:9000
}
```

WebSocket streaming works through Caddy out of the box.

**Other options:** [ngrok](https://ngrok.com/) (instant public URL), WireGuard (self-managed VPN), port forwarding with DDNS.

## GPU Support

For faster inference, use the CUDA variant of the Whisper image:

```yaml
whisper-small:
  image: fedirz/faster-whisper-server:latest-cuda
  deploy:
    resources:
      reservations:
        devices:
          - driver: nvidia
            count: 1
            capabilities: [gpu]
```

Requires an NVIDIA GPU and the [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html).

## How is Diction different?

<table width="100%">
<tr><th></th><th>Diction (self-hosted)</th><th>Diction Cloud</th><th>Wispr Flow</th><th>Apple Dictation</th></tr>
<tr><td><strong>Price</strong></td><td>Free</td><td>$3.99/mo</td><td>$15/month</td><td>Free</td></tr>
<tr><td><strong>Audio stays on your network</strong></td><td>✅</td><td>❌</td><td>❌</td><td>✅</td></tr>
<tr><td><strong>Choose your model</strong></td><td>✅ Any model, any URL</td><td>✅ Multiple models</td><td>❌ Locked in</td><td>❌ Locked in</td></tr>
<tr><td><strong>Open source</strong></td><td>✅ Gateway + infra</td><td>✅ Same gateway</td><td>❌</td><td>❌</td></tr>
<tr><td><strong>WebSocket streaming</strong></td><td>✅</td><td>✅</td><td>❌</td><td>N/A</td></tr>
<tr><td><strong>Third-party SDKs in app</strong></td><td>None</td><td>None</td><td>Unknown</td><td>N/A</td></tr>
</table>

Diction is pure transcription — what you say is what you get. No AI rewriting, no "smart" corrections, no filler word removal.

## Privacy

This is a keyboard extension. We take it seriously:

- **Self-hosted**: Audio goes only to your server. Full stop.
- **Cloud**: Audio is processed and immediately discarded. Not stored, not used for training.
- **No analytics, no tracking, no telemetry.** Zero third-party SDKs.
- **Full Access** is required by iOS for network access — the keyboard needs to reach the transcription endpoint. There is no QWERTY keyboard to log, no clipboard access.

Read the full [Privacy Policy](docs/privacy.md).

## Troubleshooting

### App issues

**Diction keyboard doesn't appear**
Settings → General → Keyboard → Keyboards → Add New Keyboard → Diction. Make sure **Allow Full Access** is enabled.

**No transcription / timeout**
Check that your endpoint URL is correct and reachable from your phone. In Self-Hosted mode, your phone must be on the same network as your server (or use [remote access](#remote-access)).

**Transcription is slow**
Try a smaller model (Small instead of Large) or enable **Stream Audio** in settings — audio uploads during recording so transcription starts instantly when you stop.

### Self-hosting issues

**Model takes a long time on first start**
Normal. Model weights are downloaded on first launch (~500 MB for Small, ~3 GB for Large V3). They're cached in a Docker volume — subsequent starts are instant.

**Health check failing**
Models need 1-2 minutes to load. Check logs: `docker compose logs -f whisper-small`

**Out of memory**
Run fewer models or pick a smaller one. One model is all you need.

**Updating**

```bash
docker compose pull
docker compose up -d
```

### Report a bug

[Open an issue](https://github.com/omachala/diction/issues/new?template=bug_report.md) with:
- Whether you're using Cloud or Self-Hosted mode
- Your model and language settings
- Steps to reproduce
- For self-hosting: Docker version, OS, and logs (`docker compose logs`)

## Requirements

- **iOS 16.0+** (iPhone)
- For self-hosting: any machine that can run Docker (the gateway uses ~15 MB RAM)

## Contributing

Contributions to the gateway, Docker setup, and documentation are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
</p>
