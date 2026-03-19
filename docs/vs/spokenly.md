---
title: Diction vs Spokenly
description: Diction vs Spokenly — compare iOS voice dictation apps. Diction is iOS-only with on-device and self-hosted modes. Spokenly is Mac-first with limited iOS support.
---

# Diction vs Spokenly

Spokenly and Diction are both indie apps built around Whisper. They solve similar problems from different angles.

## The key difference

Spokenly is a Mac app with an iOS companion. The Mac version is polished with 500+ ratings and a loyal following. The iOS version has about 30 ratings — it's clearly not the primary focus.

Diction is built exclusively for iOS. The keyboard extension is the whole product. There's no Mac version, no desktop app.

If you want dictation on iPhone as your main use case, you're comparing a Mac app's iOS port against an app built from scratch for iPhone.

## Side-by-side

| | Diction | Spokenly |
|--|---------|---------|
| iPhone keyboard | Yes | Yes |
| Mac | No | Yes (primary platform) |
| On-device (offline) | Yes | Yes |
| Self-hosted server | Yes | No |
| BYOK (own API keys) | No | Yes |
| Open-source server | Yes | No |
| Free tier | On-device free forever | On-device + BYOK free |
| Cloud subscription | Diction One | $9.99/month Pro |
| Language support | 99 (Whisper) | 100+ (Whisper) |
| AI transcript cleanup | Yes (Diction One) | Yes (Pro) |

## Platform focus

Spokenly was built as a macOS app. The iOS version uses the same transcription backend but the experience is designed around the Mac workflow: global hotkey, floating window, system-wide access. On iPhone, it works, but it's a port.

Diction's keyboard extension was built for iOS from day one. The mic button sits in the keyboard area, it works in any text field, and it behaves like a native iOS keyboard.

## Pricing

Spokenly's free tier is genuinely generous: you get on-device transcription (fully offline) and BYOK mode (bring your own OpenAI/Groq/Deepgram API keys) at no cost. You only pay $9.99/month if you want Spokenly's managed cloud models.

Diction's on-device mode is also free with no limits. Self-hosting is free if you run your own Whisper server. Diction One cloud requires a subscription.

If you already have an OpenAI or Groq key, Spokenly's BYOK is worth trying.

## Self-hosting

Diction has a Docker Compose server you can run on any hardware. Point Diction at your server and your audio stays on your infrastructure. This is the route if you want unlimited dictation without any subscription.

Spokenly doesn't have a self-hosted option. You either use their cloud or run transcription fully on-device.

## iOS App Store presence

Spokenly: ~30 iOS ratings (as of March 2026). By comparison, Diction launched in March 2026 and is building iOS presence from the start. Spokenly's iOS version is maintained but secondary.

If you're specifically choosing an iOS voice keyboard, the relative attention each team gives to iOS matters. Spokenly is maintained primarily for Mac.

## When Spokenly is better

If you work on a Mac and want dictation across all your apps, Spokenly's Mac version is excellent — 4.8/5 with 566 ratings. The BYOK mode is unique and useful if you already pay for OpenAI API access.

If you want voice dictation on both Mac and iPhone from one subscription, Spokenly covers both.

## When Diction is better

If iPhone is your main device and you want a keyboard built for that, Diction is the iOS-native choice. On-device mode, self-hosting, and keyboard extension behavior are all tuned for iPhone.

Self-hosting is Diction's strongest differentiator. If you run a home server, you can have unlimited, private, self-hosted iPhone dictation with no subscription. Spokenly doesn't offer this.

---

_Diction is available on the [App Store](https://apps.apple.com/app/id6759807364). The server is open source at [github.com/omachala/diction](https://github.com/omachala/diction)._
