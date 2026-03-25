---
title: Release Notes
description: What's new in each Diction release. Full changelog for every App Store version
---

# Release Notes

## Diction 3.0

*March 25, 2026*

- Speak to Edit. Select any text in your document, switch to Diction, and dictate an instruction: "make this shorter", "more formal", "fix the grammar". The AI rewrites the selected text in place. No copy-paste, no switching apps back and forth.
- Tone presets. Set a writing style for each app. Casual for Messages, professional for email, technical for Notion. AI Enhancement picks up the tone automatically so you don't have to adjust anything between apps.
- My Words. A new vocabulary screen in Preferences where you teach Diction names, acronyms, and technical terms it should always get right. Add them once and they stick.
- Smarter AI Enhancement. Diction now builds context across your dictation session. If you've been talking about a specific topic, the AI cleanup understands the domain and makes better decisions for the next transcription.
- Time saved. Insights now shows how much time Diction has saved you, calculated from your dictation speed versus typical typing speed.
- Haptic feedback. Subtle taps when you start dictating, stop, and insert text. If you prefer silence, toggle it off in Preferences.
- AirPods fix. Music no longer pauses when you start dictating with AirPods connected.
- Better history search. Results now highlight the matching text so you can spot what you're looking for at a glance. Tap any entry to open the full transcript detail view.
- Mid-sentence insertion fix. Capitalization and spacing now work correctly when you dictate into the middle of existing text.

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
