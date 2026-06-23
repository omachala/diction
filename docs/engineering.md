---
title: "How Diction Is Built"
description: Engineering principles behind Diction. Under 15 MB. Custom audio modules. No analytics. Defence in depth. The same instincts that ship banking infrastructure, applied to an iOS keyboard.
keywords: "diction engineering, keyboard size, custom audio modules, no analytics, ios keyboard memory, banking engineering"
---

# How Diction Is Built

iOS keyboard extensions have tight memory limits and no tolerance for slowness. Most dictation tools route around that constraint by living in a separate app. Diction lives inside the constraint.

This page is the engineering side. What the keyboard is made of, what we won't put in it, and why.

## Under 15 MB

The keyboard fits in less than 15 MB. The whole thing.

Tools in this category routinely ship a hundred MB or more because they bundle large SDKs, multiple speech models, analytics, ad networks, and code for platforms that don't apply on iOS. Diction ships none of that. The keyboard loads instantly, has room to breathe inside the 48 MB process ceiling iOS gives it, and leaves headroom for the actual work — buffering audio, running the model, talking to the host app.

Staying small is a choice, not an accident. Every dependency gets read before it's added. Most of the time it doesn't get added.

## Custom audio modules

A lot of the audio and processing layers are custom. Background noise filtering. Mic warmup. Clipping detection. Silence handling. Voice activity detection. The pipeline that sits between the microphone and the speech model.

The available libraries were either too heavy (pulling in tens of MB of dependencies for the one thing we needed) or too imprecise for an iOS keyboard. Custom modules give us the right footprint and the right behavior. They also give us full visibility when something goes wrong — no black boxes between the mic and the model.

## No analytics in the app

The Diction app has zero analytics, zero tracking, zero data collection.

We don't ship an SDK that phones home. We don't measure which buttons you tap. We don't know how many seconds you dictated last week. This is unusual enough that it's worth saying out loud, but it's not a marketing position. It's a default. There's nothing about building a good keyboard that requires watching the user use it.

The website uses Google Analytics, like most websites. The app does not.

## Defence in depth

Audio is held only as long as needed. Network traffic to Diction One is encrypted. The server processes the audio and drops it. Transcriptions are encrypted with AES-256-GCM before they leave the server, with X25519 used for key exchange. These choices stack: each layer is small on its own, and together they leave no good place for a breach to land.

If you self-host, audio never leaves your network. If you're on-device, audio never leaves your phone. If you're on Diction One, audio reaches our server, gets transcribed, and is immediately dropped. Three different paths, the same principle. The fewer places your data exists, the fewer places it can leak.

The full crypto details are on the [Encryption](/encryption) page.

## The server is open source

The server side of Diction is open source. Every line you depend on for self-hosted mode is public. You can read it, audit it, fork it, and run a version with your own changes. Your own model. Your own infrastructure. Your own changes to the prompt or the cleanup pass.

The iOS app itself is not open source. The TestFlight builds and App Store builds are how the keyboard is distributed. But the part that handles your audio after it leaves the phone — that part is open. If you don't trust our cloud, you don't have to use it. If you don't trust the cloud at all, you can stay on-device.

## What we won't add

A short list of things we are deliberately not building.

- **Telemetry in the keyboard.** Even anonymous. Even useful. Not in the keyboard.
- **Ads.** Ever.
- **Third-party SDKs.** Anything that pulls in code we don't write or audit doesn't ship. The bar is high because the surface is small.
- **A second product on the side.** The keyboard is the product. Not a notes app, not a journaling app, not an AI assistant. Voice input that feeds whatever app you're already in.

The list grows when we say no to something. It is the longest list in the project.

## Why this matters

I work on banking infrastructure for a living. The kind of environment where the reliability and privacy bar is set by regulators, not product managers. Systems don't go down. Data doesn't leak. Those aren't goals, they're the baseline.

Diction inherits the same instincts. No unnecessary data retention. No assumptions about the happy path. Defence in depth. Audio processed and immediately discarded. Open code on the server side so claims about privacy are verifiable, not promises.

You don't need to take the privacy claims on faith. You can read the server. You can self-host. You can stay on-device. The architecture is built so that trust is optional.

---

For who's behind all this, read [About the Author](/about). For why this is the only thing being built, read [Focus Over Spread](/focus).

[Download for iOS](https://apps.apple.com/app/id6759807364)
