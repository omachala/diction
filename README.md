<p align="center">
  <!-- <img src="assets/icon.png" width="128" height="128" alt="Diction app icon"> -->
  <h1 align="center">Diction</h1>
  <p align="center">The dictation-only iOS keyboard.<br>Tap. Speak. Done.</p>
</p>

<p align="center">
  <!-- <a href="https://apps.apple.com/app/diction/id000000000"><img src="https://developer.apple.com/assets/elements/badges/download-on-the-app-store.svg" alt="Download on the App Store" height="40"></a> -->
  <a href="https://diction.one">Website</a> &bull;
  <a href="docs/self-hosting.md">Self-Hosting Guide</a> &bull;
  <a href="docs/privacy.md">Privacy Policy</a>
</p>

<p align="center">
  <img src="https://img.shields.io/github/license/omachala/diction" alt="License">
  <!-- <img src="https://img.shields.io/github/v/release/omachala/diction" alt="Release"> -->
  <!-- <img src="https://ghcr-badge.egpl.dev/omachala/diction-gateway/size" alt="Docker image size"> -->
</p>

---

<!-- TODO: Add screenshot or demo GIF
<p align="center">
  <img src="assets/demo.gif" width="300" alt="Diction demo">
</p>
-->

## What is Diction?

Diction is an iOS keyboard that replaces your QWERTY layout with a single purpose: **speech-to-text**. It uses [OpenAI Whisper](https://github.com/openai/whisper) to transcribe your voice — and it works with any app on your iPhone.

- Tap the mic, speak, tap stop. Text appears.
- No typing. No autocorrect. No QWERTY.
- Works in every app — Messages, Notes, Safari, Slack, anything with a text field.

### Why?

Most people already have a keyboard they like. What they don't have is a keyboard that's *only* for dictation — fast, focused, no distractions.

Diction is also **self-hosted-first**. Your audio can go to a Whisper server you run on your own hardware. No cloud required. No data leaves your network.

## Features

- **Dictation-only keyboard** — no QWERTY, no distractions
- **Self-hosted Whisper** — point it at your own server, keep audio on your network
- **Diction Cloud** — or use our hosted API with no setup required
- **Multiple Whisper models** — tiny, small, medium, large-v3, distil-large-v3
- **OpenAI-compatible API** — works with faster-whisper, whisper.cpp, or OpenAI's API
- **Zero dependencies** — pure Swift, no third-party SDKs, fully auditable
- **Privacy-first** — no analytics, no tracking, no data collection

## How It Works

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  iOS Device  │────▶│  Diction Gateway  │────▶│  Whisper Model   │
│  (keyboard)  │◀────│  (Go, routes by   │◀────│  (faster-whisper) │
│              │     │   model field)    │     │                  │
└─────────────┘     └──────────────────┘     └─────────────────┘
```

1. You tap the mic in any app
2. Diction records audio (m4a, 16kHz mono)
3. Audio is sent to a Whisper-compatible endpoint (self-hosted or cloud)
4. Transcribed text is inserted into the active text field

## Self-Hosting

Run your own Whisper transcription server with Docker. Your audio stays on your network.

### Quick Start

```bash
# Clone and start
git clone https://github.com/omachala/diction.git
cd diction

# Copy and edit environment variables
cp .env.example .env

# Start with a single model (recommended to start)
docker compose up -d whisper-small gateway

# Or start all models
docker compose up -d
```

The gateway will be available at `http://localhost:9000`.

Point your Diction app's endpoint setting to your server's address and you're done.

### Models

| Model | RAM | Latency | Best for |
|-------|-----|---------|----------|
| tiny | ~350 MB | ~1-2s | Quick notes in quiet environments |
| small | ~800 MB | ~3-4s | Everyday dictation |
| medium | ~1.8 GB | ~8-12s | Accents and background noise |
| large-v3 | ~3.5 GB | ~20-30s | Best accuracy, difficult audio |
| distil-large-v3 | ~2 GB | ~4-6s | Near large-v3 accuracy at 6x speed |

**No public IP?** No problem — use [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) or [Tailscale](https://tailscale.com/) to reach your server from anywhere without port forwarding.

See the [Self-Hosting Guide](docs/self-hosting.md) for detailed setup instructions, GPU support, remote access, and more.

## Requirements

- **iOS 17.0+**
- **iPhone** (iPad support planned)
- For self-hosting: any machine that can run Docker (Linux, macOS, Windows)

## Privacy

Diction is designed with privacy as a core principle:

- **Self-hosted mode**: Audio is sent only to your server. Nothing touches the internet.
- **Cloud mode**: Audio is processed and immediately discarded. No storage, no training data.
- **No analytics**: The app contains zero tracking or analytics SDKs.
- **No data collection**: We don't collect, store, or sell any user data.
- **Full Access**: The keyboard requests Full Access only for network — it needs to reach the Whisper endpoint. No keylogging, no clipboard access.

Read the full [Privacy Policy](docs/privacy.md).

## Contributing

We welcome contributions to the self-hosting infrastructure, documentation, and Docker setup. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License — see [LICENSE](LICENSE) for details.

The iOS app source code is not included in this repository. The app is distributed via the App Store.
