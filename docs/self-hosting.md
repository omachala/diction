# Self-Hosting Guide

The full self-hosting documentation is in the [README](../README.md).

Quick start:

```bash
git clone https://github.com/omachala/diction.git
cd diction
docker compose up -d gateway whisper-small
```

Then in the Diction app: **Settings → Self-Hosted** → set the endpoint URL to `http://<your-server-ip>:8080`.

See the [README](../README.md) for model options, remote access (Cloudflare Tunnel, Tailscale), and the full self-hosting guide.
