<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/logo-light.png">
    <source media="(prefers-color-scheme: light)" srcset="assets/logo-dark.png">
    <img src="assets/logo-dark.png" alt="Diction" height="50">
  </picture>
  <br><br>
  <strong>You talk. We type.</strong>
  <br><br>
  Voice keyboard for iOS. Works in every app.<br>On-device, cloud, or self-hosted transcription. No limits.
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

- **Deep audio engineering.** State-of-the-art audio filtering, a fine-tuned speech recognition model, and context-aware processing, built by a real engineer who goes deep on one problem.
- **Self-hosted in one command.** `docker compose up` and paste the URL. Your server, your models, your data.
- **Works with any Whisper server.** Diction speaks the OpenAI transcription API directly. Point it at any endpoint that implements it, with or without our gateway.
- **Transcripts encrypted in transit.** AES-256-GCM with X25519 key exchange between the app and the gateway. Same primitives used by Signal and WireGuard.
- **Zero tracking in the app.** No analytics, no telemetry, no data collection. Audit the source yourself.
- **On-device.** Whisper runs locally on your iPhone. No network, no server, nothing leaves the device.
- **Free and unlimited.** On-device and self-hosted have no caps, no restrictions, no expiry.

## How It Works

### On-Device (Free, No Setup)

Install the app, add the keyboard, and start dictating. On-device transcription works offline with no server required.

### Self-Hosted

> **Step-by-step walkthrough:** [How to set up Diction — the self-hosted speech-to-text alternative to Wispr Flow](https://dev.to/omachala/how-to-set-up-diction-the-self-hosted-speech-to-text-alternative-to-wispr-flow-20km)

Diction speaks the OpenAI transcription API (`POST /v1/audio/transcriptions`) directly. Any server that implements it works. There are three ways to set it up, depending on what you already have running and how much you care about latency.

#### Path 1: Whisper only (simplest)

One container, no gateway, no extra moving parts. Save this as `docker-compose.yml`:

```yaml
services:
  whisper:
    image: fedirz/faster-whisper-server:latest-cpu
    ports:
      - "8000:8000"
    environment:
      WHISPER__MODEL: Systran/faster-whisper-small
      WHISPER__INFERENCE_DEVICE: cpu
```

Run `docker compose up -d`, paste `http://your-server:8000` into the Diction app's **Self-Hosted** tab, done.

The trade-off: no streaming. The app uploads your recording after you stop talking and waits for whisper to transcribe it. Short phrases feel fine. On longer dictations you'll see a visible pause between the moment you stop and the text appearing.

#### Path 2: Whisper + the Diction gateway (recommended)

Add our open-source gateway in front of whisper. It exposes a WebSocket endpoint that lets the app stream audio as you speak, so the transcript is mostly ready the moment you stop.

```yaml
services:
  gateway:
    image: ghcr.io/omachala/diction-gateway:latest
    ports:
      - "8080:8080"
    environment:
      DEFAULT_MODEL: small

  whisper-small:
    image: fedirz/faster-whisper-server:latest-cpu
    environment:
      WHISPER__MODEL: Systran/faster-whisper-small
      WHISPER__INFERENCE_DEVICE: cpu
```

Run `docker compose up -d`, point the app at `http://your-server:8080`. Short phrases feel about the same as Path 1. Longer dictations are noticeably faster. The longer you talk, the bigger the gap.

Your server needs to be reachable from your phone. See [No Public IP?](#no-public-ip) for Cloudflare Tunnel, Tailscale, and ngrok options.

#### Path 3: You already have a Whisper server

If you already run one (a fine-tuned model, bigger hardware, something domain-specific), you have the same two choices.

**Point Diction straight at it.** Paste your existing server's URL into the app. If your server speaks `POST /v1/audio/transcriptions`, you're done. No extra containers.

**Or wrap it with the gateway** to get streaming on top of your existing setup:

```yaml
services:
  gateway:
    image: ghcr.io/omachala/diction-gateway:latest
    ports:
      - "8080:8080"
    environment:
      CUSTOM_BACKEND_URL: http://your-existing-server:8000
      CUSTOM_BACKEND_MODEL: your-model-name
```

Behind an API key? Add `CUSTOM_BACKEND_AUTH: "Bearer sk-xxx"`. Server only accepts WAV? Add `CUSTOM_BACKEND_NEEDS_WAV: "true"` and the gateway converts with ffmpeg before forwarding. Full reference: [Use Your Own Model](https://diction.one/features/custom-model).

#### Models

Swap `WHISPER__MODEL` in your compose file:

| Model ID | Params | RAM |
|----------|--------|-----|
| `Systran/faster-whisper-small` | 244M | ~850 MB |
| `Systran/faster-whisper-medium` | 769M | ~2.1 GB |
| `deepdml/faster-whisper-large-v3-turbo-ct2` | 809M | ~2.3 GB |

Larger models are more accurate but need more RAM. On a GPU, even the large turbo feels instant. On CPU, small is the sweet spot for everyday dictation.

> If you use Path 2 or 3 with a model other than `small`, set `DEFAULT_MODEL` on the gateway to match (`small`, `medium`, or `large-v3-turbo`) and use the service name the gateway expects: `whisper-small`, `whisper-medium`, or `whisper-large-turbo`.

#### Parakeet (alternative to Whisper)

[Parakeet TDT](https://huggingface.co/nvidia/parakeet-tdt-0.6b-v2) is NVIDIA's speech-to-text engine. More accurate and faster than Whisper for European languages, with lower RAM requirements. The trade-off: it supports 25 languages instead of Whisper's 99.

| | Whisper Large-v3 | Parakeet TDT 0.6B v3 |
|---|---|---|
| WER (English) | 7.4% | 6.34% |
| Speed | Baseline | ~10x faster |
| RAM (INT8) | ~3-4 GB | ~2 GB |
| Languages | 99 | 25 European |

If you mostly dictate in a European language, Parakeet is the better engine. Save this as `docker-compose.yml`:

```yaml
services:
  gateway:
    image: ghcr.io/omachala/diction-gateway:latest
    ports:
      - "8080:8080"
    environment:
      DEFAULT_MODEL: parakeet-v3

  parakeet:
    image: ghcr.io/achetronic/parakeet:latest-int8
```

Run `docker compose up -d`, point the app at `http://your-server:8080`.

Models are baked into the image, so there's no download delay on first start.

**Supported languages:** English, Bulgarian, Croatian, Czech, Danish, Dutch, Estonian, Finnish, French, German, Greek, Hungarian, Italian, Latvian, Lithuanian, Maltese, Polish, Portuguese, Romanian, Slovak, Slovenian, Spanish, Swedish, Russian, Ukrainian.

For non-European languages (Asian, Arabic, etc.) use Whisper instead.

The full compose file in this repo includes Parakeet as a profile:

```bash
docker compose --profile parakeet up -d
```

Set `DEFAULT_MODEL: parakeet-v3` on the gateway to match.

#### NixOS

If your server runs NixOS, the repo ships a flake with a hardened systemd module — no Docker needed. Try it first without committing to anything:

```bash
nix run github:omachala/diction#diction-gateway
```

To run it as a service, import the module and enable it:

```nix
{
  inputs.diction.url = "github:omachala/diction";

  outputs = { nixpkgs, diction, ... }: {
    nixosConfigurations.your-host = nixpkgs.lib.nixosSystem {
      modules = [
        diction.nixosModules.default
        {
          services.diction-gateway = {
            enable = true;
            openFirewall = true;
            # customBackend.url = "http://127.0.0.1:8000";
            # llm.baseUrl = "http://127.0.0.1:11434/v1";
            # llm.model = "gemma2:9b";
            # environmentFile = "/run/secrets/diction-gateway.env";
          };
        }
      ];
    };
  };
}
```

The unit runs under `DynamicUser` with `ProtectSystem=strict`, `NoNewPrivileges`, and a narrow syscall filter. Use `environmentFile` for secrets like `CUSTOM_BACKEND_AUTH`, `LLM_API_KEY`, and `TRIAL_SECRET` so they don't end up in the world-readable Nix store. Full option list: [`nix/module.nix`](nix/module.nix).

### Compatible with OpenAI speech API clients

The gateway implements the OpenAI audio transcription API. Any client that works against `api.openai.com/v1/audio/transcriptions` works against a Diction gateway — the `openai` Python SDK, `openai` Node SDK, LangChain, and any other OpenAI-compatible library. This also makes Diction a drop-in STT replacement for [Speaches](https://github.com/speaches-ai/speaches).

**Supported:**

- `POST /v1/audio/transcriptions` — `file`, `model`, `language`, `prompt`, `response_format=json|text`
- `GET /v1/models` — OpenAI list format with HuggingFace-style model IDs
- HuggingFace model IDs: `Systran/faster-whisper-small`, `nvidia/parakeet-tdt-0.6b-v3`, `nvidia/canary-1b-v2`, etc.

**Not supported:**

- TTS (`/v1/audio/speech`)
- `response_format=verbose_json|srt|vtt` (no word-level timestamps)
- SSE streaming on REST (use Diction's WebSocket `/v1/audio/stream` instead)
- Model download/delete (`POST`/`DELETE /v1/models/{id}`)
- Realtime API (`/v1/realtime`)

**Authentication is off by default** (`AUTH_ENABLED=false`). Self-hosted deployments accept requests without a token. Pass any non-empty string as the API key from the client — the gateway doesn't check it.

**Known limitation:** error responses use Diction's `{"error":"<message>"}` shape, not OpenAI's nested `{"error":{"message":...,"type":...}}`. Most SDKs surface these as raw `HTTPError` rather than `APIError` — catch both.

Quickstart with the Python SDK:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://your-gateway:8080/v1",
    api_key="anything",  # not checked when AUTH_ENABLED=false
)

with open("audio.wav", "rb") as f:
    result = client.audio.transcriptions.create(
        file=f,
        model="Systran/faster-whisper-small",
        response_format="text",
    )
print(result)  # plain string
```

## AI Enhancement (BYO LLM)

You say "so um basically the meeting went well and uh they agreed to the timeline." Your server turns that into "The meeting went well. They agreed to the timeline."

The gateway can pass transcripts through any LLM before returning them. OpenAI, Groq, Anthropic, Ollama on the same machine, or any provider that speaks the OpenAI chat completions format. You write the prompt. You pick the model.

If no LLM is configured, nothing changes. Transcripts come back raw, same as before.

### How it works

```
iPhone → gateway → Whisper → raw transcript
  → your LLM (chat/completions)
  → cleaned text back to the app
```

The app sends `?enhance=true` on the request. The gateway sends your transcript to `{LLM_BASE_URL}/chat/completions` with your system prompt, and returns whatever the LLM sends back.

### Configuration

Four environment variables on the gateway:

| Variable | Required | Description |
|----------|----------|-------------|
| `LLM_BASE_URL` | Yes | OpenAI-compatible endpoint (e.g. `https://api.openai.com/v1`) |
| `LLM_MODEL` | Yes | Model identifier (e.g. `gpt-4o-mini`) |
| `LLM_API_KEY` | No | Bearer token. Your OpenAI/Groq/etc. key, or any string for local Ollama |
| `LLM_PROMPT` | No | System prompt. Supports file paths (e.g. `/config/prompt.txt` via volume mount) |

Both `LLM_BASE_URL` and `LLM_MODEL` must be set. If either is missing, the feature is off.

### Quickstart with OpenAI

Set these on the gateway service in your `docker-compose.yml`:

```yaml
environment:
  LLM_BASE_URL: https://api.openai.com/v1
  LLM_API_KEY: sk-...
  LLM_MODEL: gpt-4o-mini
  LLM_PROMPT: "You are a transcript cleaner. Fix grammar, punctuation, and capitalization. Remove filler words. Correct speech-to-text errors. Return only the cleaned text."
```

Restart the gateway and you're done. Works with any OpenAI-compatible provider. Swap the base URL for Groq (`https://api.groq.com/openai/v1`), Together, Fireworks, or anything else that speaks the same format.

### Local with Ollama

If you'd rather keep everything on your network, the compose file includes an Ollama profile:

```bash
docker compose --profile ollama up -d
docker exec diction-ollama ollama pull gemma2:9b
```

Uncomment the `LLM_*` variables in `docker-compose.yml`, restart the gateway, and transcripts run through your local model. No API key, no external calls.

#### Model recommendations for local

| Model | Size | RAM | Notes |
|-------|------|-----|-------|
| Gemma 2 9B | 9B | ~6 GB | Best editing quality at this size |
| Qwen 2.5 7B | 7B | ~5 GB | Strong instruction following |
| Llama 3.1 8B | 8B | ~5 GB | Most popular, well-tested |
| Gemma 3 4B | 4B | ~3 GB | Limited hardware |

Models under 7B tend to answer questions about the text instead of correcting it. 7B or larger works best.

### Example prompt

```
You are a transcript cleaner. Fix grammar, punctuation, and capitalization.
Remove filler words (um, uh, er, like, you know). Correct common
speech-to-text errors. Return only the cleaned text, nothing else.
```

For longer prompts, use a file: set `LLM_PROMPT=/config/prompt.txt` and mount it into the container.

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

If you don't want to run a server, Diction One gives you a fine-tuned cloud model with advanced audio filters — without the setup. Audio is sent to the Diction endpoint, transcribed, and immediately discarded. Pricing and trial details are in the app.

## Requirements

- **iOS 17.0+** (iPhone)
- For self-hosting: any machine that can run Docker

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT. See [LICENSE](LICENSE).

The iOS app is distributed via the App Store. This repository contains the self-hosting infrastructure and documentation.
