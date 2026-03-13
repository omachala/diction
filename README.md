<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/logo-light.png">
    <source media="(prefers-color-scheme: light)" srcset="assets/logo-dark.png">
    <img src="assets/logo-dark.png" alt="Diction" height="50">
  </picture>
  <br><br>
  <strong>You talk. We type.</strong>
  <br><br>
  Voice keyboard for iOS. Tap mic, speak, done.<br>Works in any app — on your device or your own server.
</p>

<p align="center">
  <a href="https://apps.apple.com/app/diction-voice-keyboard/id6759807364"><img src="https://developer.apple.com/assets/elements/badges/download-on-the-app-store.svg" alt="Download on the App Store" height="40"></a>
</p>

<p align="center">
  <a href="https://diction.one">Website</a> &bull;
  <a href="docs/self-hosting.md">Self-Hosting Guide</a> &bull;
  <a href="docs/privacy.md">Privacy Policy</a>
</p>

<p align="center">
  <a href="https://github.com/omachala/diction/blob/main/LICENSE"><img src="https://img.shields.io/github/license/omachala/diction?style=for-the-badge" alt="License"></a>
  <a href="https://codecov.io/gh/omachala/diction"><img src="https://img.shields.io/codecov/c/github/omachala/diction?style=for-the-badge" alt="Coverage"></a>
</p>

---

<p align="center">
  <img src="assets/slide-01.png" width="200" alt="You talk. We type.">&nbsp;
  <img src="assets/slide-02.png" width="200" alt="No limits. No word caps. No catch.">&nbsp;
  <img src="assets/slide-03.png" width="200" alt="What you say stays with you.">&nbsp;
  <img src="assets/slide-04.png" width="200" alt="Self-host. Your server, your rules.">
</p>

## Why Diction?

- **Works in every app** — switch to Diction in any text field: Messages, Notes, email, search bars, anything. Tap mic, speak, text appears.
- **Your audio stays where you send it** — on-device works offline, nothing leaves your phone. Self-hosted sends audio to your server only. No third-party routing.
- **On-device and self-hosted are free, no limits** — no word caps, no expiry, no catch. [Diction One](#diction-one) ($5.99/month) adds the hosted option if you don't want to run a server.
- **Three commands to self-host** — clone this repo, run `docker compose up`, paste the URL in the app. Done.
- **Run any model** — point Diction at a model fine-tuned for medical dictation, your language, or your accent. The gateway handles the routing.
- **99 languages** — Whisper handles them all, no extra config.

## How It Works

### On-Device (Free, No Setup)

Install the app, add the keyboard, and start dictating. On-device transcription works offline with no server required.

### Self-Hosted

Save this as `docker-compose.yml` and run `docker compose up -d`:

```yaml
services:
  gateway:
    image: ghcr.io/omachala/diction-gateway:latest
    ports:
      - "8080:8080"

  whisper-small:
    image: fedirz/faster-whisper-server:latest-cpu
    environment:
      WHISPER__MODEL: Systran/faster-whisper-small
      WHISPER__INFERENCE_DEVICE: cpu
```

Your server needs to be reachable from your phone. See [No Public IP?](#no-public-ip) for options like Cloudflare Tunnel, Tailscale, or ngrok.

Once reachable, open the Diction app, go to **Self-Hosted**, paste your server URL. Done.

#### More models

Swap or add models to your compose file. The gateway handles routing and streaming between them.

| Model | RAM | Latency (CPU) |
|-------|-----|---------------|
| `whisper-small` | ~800 MB | ~3-4s |
| `whisper-medium` | ~1.8 GB | ~8-12s |
| `whisper-large` | ~3.5 GB | ~20-30s |

#### Bring your own model

Run a model fine-tuned for your language, licensed for your industry, or trained on your domain. Point Diction at it and it just works.

## No Public IP?

No problem. You don't need to open ports on your router:

- **[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/)** - free, outbound-only connection. No port forwarding needed.
- **[Tailscale](https://tailscale.com/)** - free WireGuard mesh VPN. Install on server + phone, connect from anywhere.
- **[ngrok](https://ngrok.com/)** - instant public URL, great for testing.

See the [Self-Hosting Guide](docs/self-hosting.md) for detailed instructions.

## Privacy

Keyboards can read everything you type. Here's exactly what Diction does with your audio:

- **On-device**: Everything stays on your phone. No network connection made.
- **Self-hosted**: Audio goes to your server only. Nothing else sees it.
- **Diction One**: Audio is transcribed and immediately discarded. Not stored, not used for training.
- **Zero third-party SDKs.** No analytics, no tracking, no telemetry of any kind.
- **Full Access** is required by iOS for any keyboard that makes network requests. Diction has no QWERTY input to log — it only uses the network to reach your transcription endpoint.

Read the full [Privacy Policy](https://diction.one/privacy).

## Diction One

Don't want to run a server? Diction One ($5.99/month, 2-week free trial) gives you cloud transcription without the setup — audio is sent to the Diction endpoint, transcribed, and immediately discarded.

On-device and self-hosted modes are completely free with no word limits.

## Requirements

- **iOS 17.0+** (iPhone)
- For self-hosting: any machine that can run Docker

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT. See [LICENSE](LICENSE).

The iOS app is distributed via the App Store. This repository contains the self-hosting infrastructure and documentation.
