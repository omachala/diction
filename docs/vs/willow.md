---
title: "Diction vs Willow Voice (2026): On-Device & Self-Hosted vs Cloud-Only"
description: "Diction vs Willow Voice: fully offline on-device mode and self-hosted Whisper vs cloud-only. No word caps vs 2,000 words/week free tier. AES-256 text encryption."
keywords: "willow voice alternative, diction vs willow, willow voice self hosted, ios voice keyboard offline, willow voice privacy, willow voice word limit, offline voice keyboard iphone"
---

# Diction vs Willow

Both are iOS voice keyboards. The difference is where your audio goes and what it costs.

## The core difference

Willow is primarily cloud-based. Your audio gets sent to their servers for transcription. The iOS app has a limited offline fallback, but it's not a full on-device mode. There is no way to self-host.

Diction lets you choose. On-device mode runs entirely on your iPhone with zero network requests. Self-hosted mode sends audio to a server you control. Cloud is available if you want it, but never required.

## Side-by-side

| | Diction | Willow |
|--|---------|--------|
| On-device (fully offline) | Yes | Limited fallback |
| Self-hosted | Yes | No |
| Cloud mode | Yes (optional) | Required |
| iOS keyboard | Yes | Yes |
| Open-source server | Yes | No |
| End-to-end text encryption | AES-256 + X25519 | No |
| Free tier | On-device is free, no limits | 2,000 words/week |
| Paid plan | Subscription for cloud | $12-15/month |

## Privacy

Willow says they don't collect transcriptions by default. But cloud-only means your audio always leaves your device. You're trusting their infrastructure and their policies.

With Diction on-device, there is no server involved. Audio is processed locally and discarded. Nothing is transmitted. In self-hosted mode, audio goes to your server only. Either way, transcriptions are encrypted with AES-256-GCM using X25519 key exchange before they travel anywhere.

## Pricing

Willow caps the free tier at 2,000 words per week. After that, it's $12-15/month for unlimited.

Diction's on-device mode is free with no word limits, no weekly caps, no restrictions. Self-hosting is free if you run a server. The Diction One cloud subscription is only needed if you want hosted transcription.

## Why Diction

If your audio staying on your device matters to you, Diction is the clear choice. Willow's offline fallback is limited and there's no path to self-hosting. If you work in healthcare, legal, or any environment where audio can't leave your network, cloud-only is a non-starter.

If you don't want to pay for dictation, Diction's on-device mode has no caps. Willow's free tier runs out after a few emails.

---

_Diction is available on the [App Store](https://apps.apple.com/app/id6759807364). The server is open source at [github.com/omachala/diction](https://github.com/omachala/diction)._
