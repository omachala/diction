<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/logo-light.png">
    <source media="(prefers-color-scheme: light)" srcset="assets/logo-dark.png">
    <img src="assets/logo-dark.png" alt="Diction" height="50">
  </picture>
  <br><br>
  <strong>You talk. We type.</strong>
  <br><br>
  Open-source voice keyboard for iOS. Works in every app.<br>Transcription runs on your device or your own server.
</p>

<p align="center">
  <a href="https://apps.apple.com/app/id6759807364"><img src="https://developer.apple.com/assets/elements/badges/download-on-the-app-store.svg" alt="Download on the App Store" height="40"></a>
</p>

<p align="center">
  <a href="https://diction.one">Website</a> &bull;
  <a href="https://diction.one/self-hosted">Self-Hosting Guide</a> &bull;
  <a href="https://diction.one/privacy">Privacy Policy</a>
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

- **Self-hosted.** `docker compose up` and paste the URL. Your server, your models, your data.
- **Any Whisper-compatible model.** Point Diction at any endpoint. Medical, legal, accent-tuned — run whatever you want.
- **End-to-end text encrypted.** AES-256-GCM text encryption with X25519 key exchange. Same encryption used by Signal and WireGuard.
- **Zero tracking.** No analytics, no telemetry, no data collection. Audit the source yourself.
- **On-device.** Whisper runs locally on your iPhone. No network, no server, nothing leaves the device.
- **AI enhancement.** Optional LLM cleanup — only the transcript text is sent, never the audio.
- **Free and unlimited.** On-device and self-hosted have no caps, no restrictions, no expiry.

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

| Model | Parameters | RAM |
|-------|-----------|-----|
| `whisper-small` | 244M | ~850 MB |
| `whisper-medium` | 769M | ~2.1 GB |
| `whisper-large-v3` | 1.5B | ~3.9 GB |
| `whisper-large-v3-turbo` | 809M | ~2.3 GB |

#### Bring your own model

Already running a speech model on your homelab? You don't need to run ours. Set `CUSTOM_BACKEND_URL` to point the gateway at your existing server:

```yaml
services:
  gateway:
    image: ghcr.io/omachala/diction-gateway:latest
    ports:
      - "8080:8080"
    environment:
      CUSTOM_BACKEND_URL: http://my-server:8000
      CUSTOM_BACKEND_MODEL: your-model-name  # model name your server expects
```

If your server only accepts WAV audio, add `CUSTOM_BACKEND_NEEDS_WAV: "true"` and the gateway converts automatically. For servers behind an API key, add `CUSTOM_BACKEND_AUTH: "Bearer sk-xxx"`.

Works with any server that implements `POST /v1/audio/transcriptions`. See the [full guide](https://diction.one/features/custom-model) for more examples.

## No Public IP?

You don't need to open ports on your router:

- **[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/)** - free, outbound-only connection. No port forwarding needed.
- **[Tailscale](https://tailscale.com/)** - free WireGuard mesh VPN. Install on server + phone, connect from anywhere.
- **[ngrok](https://ngrok.com/)** - instant public URL, great for testing.

See the [Self-Hosting Guide](https://diction.one/self-hosted) for detailed instructions.

## Privacy

Keyboards can read everything you type. Here's exactly what Diction does with your audio:

- **On-device**: Everything stays on your phone. No network connection made.
- **Self-hosted**: Audio goes to your server only. Nothing else sees it.
- **Diction One**: Audio is transcribed and immediately discarded. Not stored, not used for training.
- **Zero third-party SDKs.** No analytics, no tracking, no telemetry of any kind.
- **Full Access** is required by iOS for any keyboard that makes network requests. Diction has no QWERTY input to log. It only uses the network to reach your transcription endpoint.

Read the full [Privacy Policy](https://diction.one/privacy).

## Diction One

On-device and self-hosted are completely free with no word limits.

If you don't want to run a server, Diction One gives you cloud transcription without the setup. Audio is sent to the Diction endpoint, transcribed, and immediately discarded. Pricing and trial details are in the app.

## Requirements

- **iOS 17.0+** (iPhone)
- For self-hosting: any machine that can run Docker

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT. See [LICENSE](LICENSE).

The iOS app is distributed via the App Store. This repository contains the self-hosting infrastructure and documentation.
