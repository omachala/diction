---
title: "Diction vs Superwhisper: iOS-Native Keyboard vs Mac-First Port"
description: "Diction vs Superwhisper: purpose-built iOS keyboard vs Mac app ported to iOS. Self-hosted Whisper, AES-256 encryption, open-source server. No recording bugs."
keywords: "superwhisper alternative, diction vs superwhisper, superwhisper ios alternative, superwhisper self hosted, ios voice keyboard, whisper ios keyboard, superwhisper iphone bugs"
---

# Diction vs Superwhisper

Superwhisper started on Mac. Diction started on iPhone. That difference shows up everywhere.

## The core difference

Superwhisper is a Mac dictation app that added an iOS version later. Diction is built from scratch as an iOS keyboard. Every design decision, every performance optimization, every UX detail is iOS-first.

iOS keyboard extensions run in a tight sandbox with strict memory limits and no background privileges. Building a reliable one takes deep platform knowledge. Porting a Mac app into that sandbox is a different exercise entirely.

## Side-by-side

| | Diction | Superwhisper |
|--|---------|-------------|
| Built for iOS | Yes, iOS-only | Mac-first, iOS added later |
| iOS keyboard extension | Purpose-built | Ported |
| Self-hosted | Yes | No |
| Custom dictionary | Yes (My Words) | No |
| Per-app tone presets | Yes | No |
| End-to-end text encryption | AES-256 + X25519 | No |
| Open-source server | Yes | No |
| Free tier | On-device, no limits | 15 min recording limit |
| Paid plan | Subscription for cloud | $8.49/month or $249 lifetime |

## iOS experience

Diction is one button. Tap mic, speak, text appears. No modes to choose, no settings to configure before you start. It works the same in every app because it's a native iOS keyboard extension.

Superwhisper on iOS inherits its Mac complexity. Multiple modes, model selection, prompt configuration. Features that make sense on a Mac with a big screen and a pointer. On an iPhone keyboard, simplicity wins.

## Privacy and encryption

Superwhisper processes audio locally on Mac, which is good. But there's no end-to-end encryption for transcriptions, no self-hosting option, and no way to audit the server.

Diction encrypts transcription text with AES-256-GCM and X25519 key exchange. The server is open source. Self-hosted mode means your audio never touches infrastructure you don't control.

## Pricing

Superwhisper's free tier gives you 15 minutes of recording. After that, $8.49/month or $249 for lifetime.

Diction's on-device mode is free with no time limits, no word caps, no restrictions. You pay only if you want Diction One cloud transcription.

## Why Diction

If you dictate on iPhone, you want a keyboard built for iPhone. Not a Mac app squeezed into an iOS keyboard extension. Diction is purpose-built for the platform, with the constraints and polish that requires.

Diction also lets you add your own words to a custom dictionary and set a different writing tone per app. Names, jargon, and product terms come through correctly. Superwhisper has no equivalent for either.

Self-hosting, text encryption, and an open-source server are things Superwhisper doesn't offer at any price.

---

_Diction is available on the [App Store](https://apps.apple.com/app/id6759807364). The server is open source at [github.com/omachala/diction](https://github.com/omachala/diction)._
