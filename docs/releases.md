---
title: Release Notes
description: What's new in each Diction release. Full changelog for every App Store version
---

# Release Notes

## Diction 2.0

*March 15, 2026*

- Added AI Enhancement for cloud mode. After transcription, Diction can optionally clean up your text, removing filler words like "um" and "uh", fixing grammar, and polishing the result. Only the text is sent for cleanup, never the audio. Off by default, toggle it in settings.
- Introduced a mandatory setup guide that walks you through keyboard installation and permissions before you start dictating. No more guessing why things aren't working.
- Cellular download guard for on-device models. Large model downloads now only start on WiFi unless you explicitly allow cellular.
- Unified branding under "Diction One" for the cloud subscription.
- Redesigned the subscription offer card with clearer pricing and what you get.
- Improved dictation reliability with fixes for the "tap to reconnect" loop, globe key skipping past iOS keyboards, and stale heartbeat issues.
- On-device model warmup now works reliably. Models pre-warm after download so your first dictation is fast.
- Better error handling. When something goes wrong, you now see a clear full-screen message explaining what happened instead of a silent failure.
- Added a support screen with troubleshooting steps and direct contact options.
- Various UI polish across the keyboard and settings.

## Diction 1.0

*March 11, 2026*

The first public release. Everything that makes Diction what it is:

- Dictation-only keyboard for iPhone. Tap the mic, speak, text appears in any app. No QWERTY keys, no distractions.
- Three transcription modes out of the box: on-device (completely offline), self-hosted (point it at your own Whisper server), and Diction cloud.
- On-device models in three tiers. Standard downloads automatically on first launch. Larger models available for better accuracy.
- Self-hosted mode connects to any server running the Whisper API format. One Docker command to start your own.
- 99 languages with automatic detection. Speak in your language and Diction figures it out.
- No word limits, no daily caps, no session timeouts. Dictate as much as you want.
- AES-256 encryption with X25519 key exchange for cloud transcriptions.
- Zero analytics, zero tracking in the app. Your voice data is processed and immediately discarded.
- Configurable idle timeout for hands-free use.
- App Store screenshots and website at [diction.one](https://diction.one).
