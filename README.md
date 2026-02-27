<p align="center">
  <!-- <img src="assets/icon.png" width="128" height="128" alt="Diction app icon"> -->
  <h1 align="center">Diction</h1>
  <p align="center"><strong>The free, open-source alternative to <a href="https://wisprflow.ai">Wispr Flow</a>.</strong><br>Self-hosted speech-to-text keyboard for iOS.</p>
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
</p>

---

<!-- TODO: Add screenshot or demo GIF
<p align="center">
  <img src="assets/demo.gif" width="300" alt="Diction demo">
</p>
-->

## Why Diction?

Voice-to-text keyboards like [Wispr Flow](https://wisprflow.ai) cost **$15/month** and send your audio to their cloud. Apple's built-in dictation is free but unreliable.

**Diction is different:**

- **Free forever** for self-hosted -no subscription, no word limits, no trial that expires
- **Your server, your data** -audio goes to a Whisper server you run. Not our cloud. Not anyone's cloud. Your network.
- **Open source infrastructure** -the server setup is right here. Inspect it, modify it, contribute to it.
- **Zero-dependency iOS app** -pure Swift, no third-party SDKs, no analytics, no tracking. Fully auditable.

Don't want to self-host? **Diction Cloud** provides the same experience with zero setup.

> **Think of it like [Bitwarden](https://bitwarden.com)** -free and self-hosted for those who want control, with a hosted cloud option for convenience.

## How It Works

1. Run a Whisper container on any machine (home server, NAS, cloud VM, Raspberry Pi)
2. Make it reachable from your phone (local IP, reverse proxy, or [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/))
3. Paste the URL into the Diction app
4. Switch to the Diction keyboard in any app → tap mic → speak → text appears

That's the entire setup. Three commands to start the server:

```bash
git clone https://github.com/omachala/diction.git
cd diction
docker compose up -d whisper-small
```

Whisper API is now running at `http://<your-server-ip>:9002`. Done.

## How is this different from...

| | Diction | Wispr Flow | Apple Dictation |
|---|---|---|---|
| **Price** | Free (self-hosted) | $15/month | Free |
| **Audio stays on your network** | ✅ | ❌ Cloud | ✅ |
| **Open source server** | ✅ | ❌ | ❌ |
| **iOS keyboard** | ✅ | ✅ | ✅ Built-in |
| **Custom Whisper endpoint** | ✅ Any URL | ❌ | ❌ |
| **Accuracy** | ✅ Whisper | ✅ Whisper + AI | ❌ Poor |
| **Zero third-party SDKs** | ✅ | ❌ | N/A |

**Wispr Flow** is a polished product with AI editing features. If you want filler word removal, grammar correction, and context-aware tone -Wispr Flow does that. Diction is pure transcription: what you say is what you get. The trade-off is freedom, privacy, and cost.

## Models

Pick the model that fits your hardware. Each runs on its own port:

```
docker compose up -d whisper-tiny          # port 9001 -~350 MB RAM, ~1-2s
docker compose up -d whisper-small         # port 9002 -~800 MB RAM, ~3-4s  ← recommended
docker compose up -d whisper-medium        # port 9003 -~1.8 GB RAM, ~8-12s
docker compose up -d whisper-large         # port 9004 -~3.5 GB RAM, ~20-30s
docker compose up -d whisper-distil-large  # port 9005 -~2 GB RAM, ~4-6s
```

Models are downloaded automatically on first start and cached. Subsequent starts are instant.

Works with any [OpenAI-compatible](https://platform.openai.com/docs/api-reference/audio/createTranscription) Whisper endpoint -[faster-whisper-server](https://github.com/fedirz/faster-whisper-server), [whisper.cpp](https://github.com/ggerganov/whisper.cpp), or OpenAI's own API.

## No Public IP?

No problem. You don't need to open ports on your router:

- **[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/)** -free, outbound-only connection to Cloudflare's edge. No port forwarding needed.
- **[Tailscale](https://tailscale.com/)** -free WireGuard mesh VPN. Install on server + phone, connect from anywhere.
- **[ngrok](https://ngrok.com/)** -instant public URL, great for testing.

See the [Self-Hosting Guide](docs/self-hosting.md) for detailed instructions.

## Privacy

This is a keyboard extension. We take privacy seriously:

- **Self-hosted**: Audio goes only to your server. Full stop.
- **Cloud mode**: Audio is processed and immediately discarded. Not stored, not used for training.
- **No analytics, no tracking, no telemetry.** The app contains zero third-party SDKs.
- **Full Access** is required by iOS for network -the keyboard needs to reach the Whisper endpoint. No keylogging, no clipboard access.

Read the full [Privacy Policy](docs/privacy.md).

## Requirements

- **iOS 17.0+** (iPhone)
- For self-hosting: any machine that can run Docker

## Contributing

We welcome contributions to the self-hosting infrastructure, documentation, and Docker setup. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT -see [LICENSE](LICENSE).

The iOS app is distributed via the App Store. This repository contains the self-hosting infrastructure and documentation.
