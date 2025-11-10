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

# Start the gateway + one model
docker compose up -d gateway whisper-small
```

Done. The gateway is now running at `http://localhost:9000`.

The gateway handles model routing, health checks, and WebSocket streaming (audio streams during recording so transcription starts instantly when you stop).

> **Direct connection:** You can also skip the gateway and point the app directly at a model (e.g. `http://localhost:9002` for small). The gateway is optional but recommended.

## Choosing a Model

Pick one model based on your hardware and needs:

| Service | Port | RAM | Latency (CPU) | Quality | Recommendation |
|---------|------|-----|---------------|---------|----------------|
| **gateway** | **9000** | **~15 MB** | **—** | **—** | **Recommended entry point** |
| tiny | 9001 | ~350 MB | ~1-2s | Good | Low-power devices, quick notes |
| **small** | **9002** | **~800 MB** | **~3-4s** | **Great** | **Best default for most users** |
| medium | 9003 | ~1.8 GB | ~8-12s | Very good | Better accent/noise handling |
| large-v3 | 9004 | ~3.5 GB | ~20-30s | Best | Maximum accuracy |
| distil-large-v3 | 9005 | ~2 GB | ~4-6s | Near-best | Best speed/accuracy trade-off |

To run the gateway with one model:

```bash
docker compose up -d gateway whisper-small
```

To run multiple models (switch between them in the app):

```bash
docker compose up -d gateway whisper-small whisper-medium whisper-large
```

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
2. Go to **Settings** → switch to **Self-Hosted** mode
3. Set the **Endpoint URL** to your gateway address (e.g. `http://192.168.1.100:9000`)
4. Pick a model and language
5. Toggle **Stream Audio** on for real-time streaming (audio streams during recording, transcription starts instantly on stop)

> **Note:** Your iPhone must be on the same network as your server, or the server must be reachable from the internet (see [Remote Access](#no-public-ip) below).

## Reverse Proxy (HTTPS)

For remote access, put the gateway behind a reverse proxy. Example with [Caddy](https://caddyserver.com):

```
diction.yourdomain.com {
    reverse_proxy localhost:9000
}
```

Caddy automatically handles SSL certificates via Let's Encrypt. WebSocket streaming (`/v1/audio/stream`) works through Caddy out of the box — no extra config needed.

## No Public IP?

If your home network doesn't have a public IP (CGNAT, double NAT, etc.), you can still access your Whisper server from anywhere:

### Cloudflare Tunnel (recommended)

[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) creates an outbound-only connection from your server to Cloudflare's edge — no port forwarding, no public IP needed.

```bash
# Install cloudflared
curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o /usr/local/bin/cloudflared
chmod +x /usr/local/bin/cloudflared

# Authenticate (one-time)
cloudflared tunnel login

# Create a tunnel
cloudflared tunnel create diction

# Route your domain to the tunnel
cloudflared tunnel route dns diction whisper.yourdomain.com

# Run (points to your local gateway)
cloudflared tunnel run --url http://localhost:9000 diction
```

You can also run `cloudflared` as a Docker service alongside your Whisper stack.

### Tailscale

[Tailscale](https://tailscale.com/) creates a private WireGuard mesh network between your devices. Install Tailscale on both your server and iPhone — your server gets a stable IP (e.g. `100.x.y.z`) accessible from your phone anywhere.

```bash
# Install on your server
curl -fsSL https://tailscale.com/install.sh | sh
tailscale up

# Your server's Tailscale IP
tailscale ip -4
# → 100.x.y.z
```

Set the Diction endpoint to `http://100.x.y.z:9000`. Works from anywhere, no domain or SSL needed (traffic is encrypted by WireGuard).

### Other options

- **[ngrok](https://ngrok.com/)** — instant public URL, good for testing (`ngrok http 9000`)
- **WireGuard** — manual VPN setup, same idea as Tailscale but self-managed
- **Port forwarding** — if your ISP gives you a public IP, forward the port on your router and use a DDNS service

## API

The gateway exposes an [OpenAI-compatible](https://platform.openai.com/docs/api-reference/audio/createTranscription) transcription API plus WebSocket streaming:

```bash
# Health check
curl http://localhost:9000/health

# List available models
curl http://localhost:9000/v1/models

# Transcribe audio (HTTP POST)
curl -X POST http://localhost:9000/v1/audio/transcriptions \
  -F file=@recording.wav \
  -F model=small

# WebSocket streaming (connect, send PCM chunks, get transcription)
# ws://localhost:9000/v1/audio/stream?model=small&language=en
#
# Protocol:
#   1. Client sends binary frames: raw PCM audio (16-bit LE, mono, 16kHz)
#   2. Client sends text frame: {"action":"done"}
#   3. Server replies with text frame: {"text":"transcribed text"}
```

## Updating

```bash
docker compose pull
docker compose up -d
```

## Troubleshooting

**Model takes a long time to start the first time**
This is normal. Model weights are downloaded on first launch (~75 MB for tiny, ~500 MB for small, ~1.5 GB for medium, ~3 GB for large-v3). They're cached in a Docker volume, so subsequent starts are instant.

**Health check failing**
Models need 1-2 minutes to load on first start. Check logs: `docker compose logs -f whisper-small`

**Out of memory**
Run fewer models, or pick a smaller one. One model is all you need.
