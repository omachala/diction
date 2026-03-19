---
title: Self-Hosted Transcription
description: Run your own Whisper server and connect Diction to it. Audio stays on your network. Free, unlimited, open-source server. Docker Compose setup.
---

# Self-Hosted Transcription

Your server. Your models. Your rules. Point Diction at any Whisper-compatible server you run and your audio never leaves your network.

Need a model for your language? A licensed model for your industry? Full privacy with full power? Self-host it.

## How It Works

1. Clone the repo and start the server:

```bash
git clone https://github.com/omachala/diction.git
cd diction/server
docker compose up
```

2. Enter your server's address in the Diction app settings.
3. Tap the mic and speak. Audio goes to your server, gets transcribed, and the text comes back.

That's it. The server code is [open source](https://github.com/omachala/diction). You can audit every line.

## What You Get

- **Audio stays on your network**: from your phone to your server and back. No third party in the middle.
- **Free, unlimited, no restrictions**: no Diction subscription needed. No word limits. No daily caps. No trial that expires.
- **Any Whisper-compatible endpoint**: works with any server that supports the OpenAI-compatible transcription API format. Not locked to our server.
- **Run anywhere**: home server, NAS, Raspberry Pi, cloud VM, behind a reverse proxy, over a VPN. If Docker runs there, Diction connects to it.
- **Pick your own models**: run whatever speech model fits your use case. Your language, your accuracy requirements, your hardware.
- **Open-source server**: the server infrastructure is fully open source on GitHub. Inspect it, modify it, contribute to it.

## Best For

- You already run Docker at home and want transcription on your own hardware
- You work in a regulated industry (medical, legal, finance) where audio cannot leave your network
- You want full control over the transcription pipeline, the models, and the infrastructure
- You need a specific model for a specific language or domain
- You refuse to send voice data to someone else's cloud
