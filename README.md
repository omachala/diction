<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/logo-light.png">
    <source media="(prefers-color-scheme: light)" srcset="assets/logo-dark.png">
    <img src="assets/logo-dark.png" alt="Diction" height="80">
  </picture>
  <br><br>
  <strong>The free, open-source alternative to <a href="https://wisprflow.ai">Wispr Flow</a>.</strong><br>Self-hosted speech-to-text keyboard for iOS.
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

- **Self-hosted is free** - no subscription, no word limits, no trial that expires. Bring your own server.
- **Your server, your data** - audio goes to a Whisper server you run. Not our cloud. Not anyone's cloud. Your network.
- **Open source infrastructure** - the server setup is right here. Inspect it, modify it, contribute to it.
- **Model agnostic** - point it at any OpenAI-compatible endpoint. Whisper tiny, large-v3, distil, fine-tuned models, future models. You choose.
- **Zero-dependency iOS app** - pure Swift, no third-party SDKs, no analytics, no tracking. Fully auditable.

Don't want to self-host? **Diction Cloud** provides the same experience with zero setup.

> **Think of it like [Bitwarden](https://bitwarden.com)** - free and self-hosted for those who want control, with a hosted cloud option for convenience.

## How It Works

1. Run a Whisper container on any machine (home server, NAS, cloud VM, Raspberry Pi)
2. Make it reachable from your phone (local IP, reverse proxy, or [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/))
3. Paste the URL into the Diction app
4. Switch to the Diction keyboard in any app → tap mic → speak → text appears

That's the entire setup. Three commands to start the server:

```bash
git clone https://github.com/omachala/diction.git
cd diction
docker compose up - d whisper-small
```

Whisper API is now running at `http://<your-server-ip>:9002`. Done.

## How is this different from...

<table width="100%">
<tr><th></th><th>Diction</th><th>Wispr Flow</th><th>Apple Dictation</th></tr>
<tr><td><strong>Price</strong></td><td>Free (self-hosted)</td><td>$15/month</td><td>Free</td></tr>
<tr><td><strong>Audio stays on your network</strong></td><td>✅</td><td>❌ Cloud</td><td>✅</td></tr>
<tr><td><strong>Open source server</strong></td><td>✅</td><td>❌</td><td>❌</td></tr>
<tr><td><strong>iOS keyboard</strong></td><td>✅</td><td>✅</td><td>✅ Built-in</td></tr>
<tr><td><strong>Model agnostic</strong></td><td>✅ Any model, any URL</td><td>❌ Locked in</td><td>❌ Locked in</td></tr>
<tr><td><strong>Zero third-party SDKs</strong></td><td>✅</td><td>❌</td><td>N/A</td></tr>
</table>

Diction is pure transcription: what you say is what you get. No AI rewriting, no filler word removal. If you want that, paid alternatives exist. Diction's trade-off is freedom, privacy, and cost.

## Models

Diction is model agnostic. It works with **any [OpenAI-compatible](https://platform.openai.com/docs/api-reference/audio/createTranscription) speech-to-text endpoint** - public models, private models, fine-tuned models, future models. You're not locked into anything.

This repo includes a Docker Compose setup with popular [faster-whisper](https://github.com/fedirz/faster-whisper-server) models to get you started:

```
docker compose up - d whisper-tiny          # port 9001 - ~350 MB RAM, ~1-2s
docker compose up - d whisper-small         # port 9002 - ~800 MB RAM, ~3-4s  ← recommended
docker compose up - d whisper-medium        # port 9003 - ~1.8 GB RAM, ~8-12s
docker compose up - d whisper-large         # port 9004 - ~3.5 GB RAM, ~20-30s
docker compose up - d whisper-distil-large  # port 9005 - ~2 GB RAM, ~4-6s
```

But you can point Diction at anything: [whisper.cpp](https://github.com/ggerganov/whisper.cpp), OpenAI's API, a custom fine-tuned model for your language or domain, or any future model that speaks the same protocol. If it has an `/v1/audio/transcriptions` endpoint, Diction works with it.

## No Public IP?

No problem. You don't need to open ports on your router:

- **[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/)** - free, outbound-only connection to Cloudflare's edge. No port forwarding needed.
- **[Tailscale](https://tailscale.com/)** - free WireGuard mesh VPN. Install on server + phone, connect from anywhere.
- **[ngrok](https://ngrok.com/)** - instant public URL, great for testing.

See the [Self-Hosting Guide](docs/self-hosting.md) for detailed instructions.

## Privacy

This is a keyboard extension. We take privacy seriously:

- **Self-hosted**: Audio goes only to your server. Full stop.
- **Cloud mode**: Audio is processed and immediately discarded. Not stored, not used for training.
- **No analytics, no tracking, no telemetry.** The app contains zero third-party SDKs.
- **Full Access** is required by iOS for network - the keyboard needs to reach the Whisper endpoint. No keylogging, no clipboard access.

Read the full [Privacy Policy](docs/privacy.md).

## Requirements

- **iOS 17.0+** (iPhone)
- For self-hosting: any machine that can run Docker

## Diction Cloud

Don't want to self-host? **Diction Cloud** is a hosted alternative - same accuracy, zero setup, no server to maintain. Priced to be cheaper than running your own VPS. See [diction.one](https://diction.one) for details.

## Contributing

We welcome contributions to the self-hosting infrastructure, documentation, and Docker setup. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT - see [LICENSE](LICENSE).

The iOS app is distributed via the App Store. This repository contains the self-hosting infrastructure and documentation.
