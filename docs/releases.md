---
title: Release Notes
description: What's new in each Diction release. Full changelog for every App Store version
---

<img src="/illustration-releases.svg" alt="Releases" class="illustration" style="max-width: 480px; margin: 0 auto 2rem; display: block;" />

# Release Notes

## Diction 4.0

*April 2026*

- Speak to Edit. Select any text, say what you want changed, and it's done. Works for simple replacements and editing instructions like "translate to Czech" or "make this shorter."
- Your custom words now improve transcription accuracy directly. Names and jargon get recognized correctly even without AI Enhancement.
- Dictate for as long as you need. Improved reliability for long recordings, no more cut-off transcripts.
- Profile lets you tell Diction who you are and how you write, so AI Enhancement matches your style.
- New guided onboarding walks you through setup step by step instead of throwing dialogs at you on first launch.
- Improved on-device model setup. Smoother download, faster preparation, automatically ready when done.
- The mic no longer activates when you open the app manually. Orange dot only when you're actually dictating.
- Improved AI Enhancement accuracy across apps.
- Various UI polish across the keyboard, history, tones, and settings.

## Diction 3.0

*March 25, 2026*

- AI Enhancement now understands where you are in your document. Dictate into the middle of a sentence and it inserts correctly. Select text and tell Diction what to do with it, and it rewrites the selection in place.
- A small pulse now lights up in the action bar when Diction hears your voice. You always know exactly when it is listening.
- Added a dominant hand setting. If you prefer left-handed use, flip the keyboard layout so the controls are on your side.
- Redesigned History with a tabbed view separating recent and all transcriptions. Search works across all of them, and tapping any entry copies it instantly.
- Carefully revisited what happens when a transcription fails. We now save your audio automatically and show a retry button on the keyboard so you never have to say it twice.
- Added a dedicated Keyboard Preferences screen. Easier to find keyboard settings, and auto-detect language is now on by default.
- Added a clear explanation screen for when Full Access is missing. No more silent failures if the keyboard is not fully set up.
- Fixed capitalization and spacing when dictating into the middle of existing text, and a rare issue where cloud transcriptions could fail silently when multiple keyboard instances were active.

## Diction 2.0

*March 15, 2026*

- AI Enhancement is now available for cloud mode. After transcription, Diction can optionally clean up your text. It removes filler words, fixes grammar, and polishes the result. Only the transcript is sent for cleanup, never the audio. Off by default, toggle it in Settings.
- Added a setup guide that walks you through keyboard installation and permissions before your first dictation. No more guessing why things are not working.
- Large model downloads now wait for WiFi by default. No surprise data bills from downloading on mobile.
- The cloud subscription is now Diction One, with a redesigned offer screen that makes pricing and what is included much clearer.
- Improved dictation reliability. Fixed the tap-to-reconnect loop, globe key skipping past iOS keyboards, and stale heartbeat issues.
- On-device models now pre-warm after download so your first dictation is fast.
- When something goes wrong, you now see a clear explanation screen instead of a silent failure.
- Added a support screen with troubleshooting steps and a way to reach us directly.
- Various UI polish across the keyboard and settings.

## Diction 1.0

*March 11, 2026*

The first public release. Everything that makes Diction what it is:

- Dictation-only keyboard for iPhone. Tap the mic, speak, and text appears wherever your cursor is in any app. No QWERTY, no distractions.
- Three transcription modes out of the box: on-device for complete offline use, self-hosted to point at your own server, and Diction cloud.
- On-device models in three tiers. The standard model downloads automatically on first launch. Larger models are available for better accuracy.
- Self-hosted mode works with any server running the Whisper API format. One Docker command to get started.
- 99 languages with automatic detection. Speak in your language and Diction figures it out.
- No word limits, no daily caps, no session timeouts. Dictate as much as you want.
- Cloud transcriptions are encrypted before they leave the server. Your audio is processed and immediately discarded.
- The Diction app contains no analytics and no tracking code.
- Configurable idle timeout for hands-free dictation.
