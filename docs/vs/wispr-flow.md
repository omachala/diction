---
title: Diction vs Wispr Flow
description: Diction vs Wispr Flow — compare privacy, pricing, and offline support for iOS voice dictation. Diction runs on-device or self-hosted. Wispr Flow requires cloud.
---

# Diction vs Wispr Flow

Both apps let you dictate into any text field on iOS. The difference is what happens to your audio.

## The core difference

Wispr Flow is 100% cloud. Your audio goes to their servers, gets transcribed, gets cleaned up by an AI model, and comes back. This works well — they've built solid infrastructure and the accuracy is good.

Diction gives you a choice. On-device mode runs entirely on your iPhone with no network requests at all. Self-hosted mode sends audio to a server you control. Diction One cloud mode is available if you want it, but it's never forced on you.

## Side-by-side

| | Diction | Wispr Flow |
|--|---------|------------|
| On-device (fully offline) | Yes | No |
| Self-hosted | Yes | No |
| Cloud mode | Yes (optional) | Required |
| iOS keyboard | Yes | Yes |
| Open-source server | Yes | No |
| Free tier | On-device is free forever | 2,000 words/week |
| Paid plan | Diction One subscription | $15/month |
| macOS | No | Yes |
| Windows | No | Yes |
| Android | No | Yes |

## Privacy

Wispr Flow has a documented history with privacy. In 2024, it was discovered they were using customer audio to train their models without explicit opt-in. The CTO apologized publicly, and training is now opt-out by default. They've added a Privacy Mode (zero data retention) and achieved SOC 2.

To be clear: Wispr Flow has fixed the worst issues. But if you work in a regulated environment, or just don't want your words leaving your device, "fixed" isn't the same as "never happened."

With Diction in on-device mode, the question doesn't arise. Audio is processed locally by WhisperKit and never transmitted anywhere. There is no server to breach, no policy to trust.

## Pricing

Wispr Flow is $15/month for unlimited dictation on Pro. There's a free tier capped at 2,000 words per week (roughly a few emails a day).

Diction's on-device mode is free with no word limits. The Diction One cloud subscription is required only if you want cloud transcription with AI enhancement. Self-hosting is free if you can run a server.

## When Wispr Flow is better

Wispr Flow has been around longer, has more platforms (macOS, Windows, Android), and a larger team building it. If you spend most of your time on a Mac and want one app that works everywhere, Wispr Flow is the more complete solution today.

Their accuracy numbers are genuinely impressive. They've published that 90% of outputs need no edits, and users send dictated text within 0.5 seconds of seeing it.

## When Diction is better

If you're on iPhone and want your audio to stay on your iPhone, Diction is the only option. No other iOS dictation keyboard offers on-device transcription through WhisperKit.

If you're self-hosting infrastructure (Home Assistant, Nextcloud, Jellyfin, anything), adding a Whisper server takes minutes and Diction just points to it. You get unlimited dictation, no subscription, and complete data control.

If you're in healthcare, legal, or anywhere that can't send audio to third-party servers, on-device or self-hosted is the practical answer.

---

_Diction is available on the [App Store](https://apps.apple.com/app/id6759807364). The server is open source at [github.com/omachala/diction](https://github.com/omachala/diction)._
