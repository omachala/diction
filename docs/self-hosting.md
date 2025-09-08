# Self-Hosting Guide

Run your own Whisper transcription server for Diction. Your audio stays on your network — nothing is sent to any cloud service.

## Requirements

- **Docker** and **Docker Compose** v2+
- At least **1 GB RAM** free (for the `small` model)
- Any Linux, macOS, or Windows machine

## Quick Start

```bash
git clone https://github.com/omachala/diction.git
cd diction
cp .env.example .env

# Start with just one model (recommended)
docker compose up -d whisper-small gateway
```

The gateway will be available at `http://localhost:9000`.

## Choosing a Model

You don't need to run all models. Pick one based on your needs:

| Model | RAM | Latency (CPU) | Quality | Recommendation |
|-------|-----|---------------|---------|----------------|
| tiny | ~350 MB | ~1-2s | Good | Low-power devices, quick notes |
| **small** | **~800 MB** | **~3-4s** | **Great** | **Best default for most users** |
| medium | ~1.8 GB | ~8-12s | Very good | Better accent/noise handling |
| large-v3 | ~3.5 GB | ~20-30s | Best | Maximum accuracy |
| distil-large-v3 | ~2 GB | ~4-6s | Near-best | Best speed/accuracy trade-off |

To run a single model:

```bash
docker compose up -d whisper-small gateway
```

To run multiple models:

```bash
docker compose up -d whisper-small whisper-medium gateway
```

The gateway automatically detects which models are running and routes requests accordingly.

## GPU Support

For significantly faster inference, use the GPU variant of the Whisper image:

```yaml
# In docker-compose.yml, change the image for any model:
whisper-small:
  image: fedirz/faster-whisper-server:latest-cuda
  # ...
  deploy:
    resources:
      reservations:
        devices:
          - driver: nvidia
            count: 1
            capabilities: [gpu]
```

Requirements: NVIDIA GPU with CUDA support, [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html) installed.

## Connecting the App

1. Open the **Diction** app on your iPhone
2. Go to **Settings**
3. Set the **Endpoint URL** to your server address: `http://<your-server-ip>:9000`
4. Select your preferred model
5. Test with the built-in test button

> **Note:** If your server is on a local network, make sure your iPhone is on the same network. For access outside your LAN, set up a reverse proxy with HTTPS.

## Reverse Proxy (HTTPS)

For production or remote access, put the gateway behind a reverse proxy. Example with [Caddy](https://caddyserver.com):

```
whisper.yourdomain.com {
    reverse_proxy localhost:9000
}
```

Caddy automatically handles SSL certificates via Let's Encrypt.

## API

The gateway exposes an OpenAI-compatible transcription API:

```bash
# Transcribe audio
curl -X POST http://localhost:9000/v1/audio/transcriptions \
  -F file=@audio.m4a \
  -F model=small

# Check available models
curl http://localhost:9000/v1/models

# Health check
curl http://localhost:9000/health
```

## Updating

```bash
docker compose pull
docker compose up -d
```

## Troubleshooting

**Models take a long time to start the first time**
This is normal. The model weights are downloaded on first launch (~75 MB for tiny, ~500 MB for small, ~1.5 GB for medium, ~3 GB for large-v3). They're cached in a Docker volume, so subsequent starts are instant.

**Gateway shows models as unavailable**
The gateway health-checks each model every 30 seconds. If a model just started, wait for it to finish loading. Check logs: `docker compose logs -f whisper-small`

**Out of memory**
Reduce the number of running models. One model at a time is fine — the gateway routes all requests to whatever's available.
