# Self-Hosting Guide

The full self-hosting documentation is in the [README](../README.md).

Quick start:

```bash
git clone https://github.com/omachala/diction.git
cd diction
docker compose up -d gateway whisper-small
```

Then in the Diction app: **Settings → Self-Hosted** → set the endpoint URL to `http://<your-server-ip>:9000`.

See the README for:
- [Models](../README.md#models) — choosing the right model for your hardware
- [Gateway](../README.md#gateway) — API reference and configuration
- [Remote Access](../README.md#remote-access) — Cloudflare Tunnel, Tailscale, reverse proxy
- [GPU Support](../README.md#gpu-support) — CUDA setup for faster inference
- [Troubleshooting](../README.md#troubleshooting) — common issues and fixes
