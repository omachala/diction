---
layout: home
title: "Voice Keyboard for iPhone: Speech to Text"
description: Voice keyboard for iPhone with open-source server. On-device, self-hosted, or cloud transcription. No word limits, no tracking, 99 languages. Free forever on your own server.

hero:
  name: Diction
  text: Voice keyboard for iPhone
  tagline: Tap the mic. Speak. Done. On-device, self-hosted, or cloud. No word limits. No tracking.
  actions:
    - theme: brand
      text: Download for iOS
      link: https://apps.apple.com/app/id6759807364
    - theme: alt
      text: GitHub
      link: https://github.com/omachala/diction

features:
  - icon:
      src: /mic-fill.svg
    title: Any App
    details: Fast and accurate speech to text iOS keyboard that works in any app. No word limits, no daily caps.
  - icon:
      src: /icon-phone.svg
    title: On-Device
    details: Speech recognition that runs locally on your iPhone. No internet required. Audio never leaves your device.
  - icon:
      src: /icon-server.svg
    title: Self-Hosted
    details: Run licensed or industry-specific models on your own server. Your network, your data.
  - icon:
      src: /icon-lock.svg
    title: Privacy First
    details: Zero analytics, zero tracking, zero third-party SDKs. Open source server you can audit.
  - icon:
      src: /icon-key.svg
    title: Encrypted
    details: AES-256 transcription encryption with X25519 key exchange. Signal and WireGuard grade privacy.
  - icon:
      src: /icon-globe.svg
    title: Your Language
    details: Automatic multilingual speech recognition.
  - icon:
      src: /icon-chat.svg
    title: AI Enhancement
    details: Turns natural speech into polished text. Automatic context-aware grammar and formatting upgrade.
  - icon:
      src: /icon-cloud.svg
    title: Diction One
    details: Cloud transcription with zero setup. Highest accuracy, lowest latency. Audio processed and immediately discarded.
---

<div class="content-section faq-section">

## FAQ

<details>
<summary>Is it really free?</summary>
<div>

On-device and self-hosted modes are completely free. No word limits, no daily caps, no restrictions. If you self-host, you run your own server and pay your own infrastructure costs. Diction One cloud requires a subscription. Pricing is shown in the app, with a free trial included.

</div>
</details>

<details>
<summary>What languages does Diction support?</summary>
<div>

99 languages via Whisper. The on-device Standard model handles most languages well. Cloud and self-hosted modes use larger models for even better accuracy across all supported languages.

</div>
</details>

<details>
<summary>Is my voice data stored?</summary>
<div>

Never. On-device mode processes audio in memory and discards it immediately. Self-hosted mode sends audio only to your server. We have no access. Diction One cloud processes and discards. No recordings retained, no model training.

</div>
</details>

<details>
<summary>How do I set it up?</summary>
<div>

Open the app, grant microphone permission, add Diction as a keyboard in iOS Settings, enable Full Access, and start dictating. Under a minute from download to first transcription. [Detailed steps here.](/support)

</div>
</details>

<details>
<summary>What is self-hosting?</summary>
<div>

You run a Whisper speech-to-text server on your own hardware. Diction connects to it over your network. Your audio never touches any third-party service. The server ships as a Docker image. One command to start.

</div>
</details>

<details>
<summary>Why does it need Full Access?</summary>
<div>

iOS requires Full Access for any keyboard extension that uses the network. Diction needs it to send audio to your server or Diction One for transcription. Diction has no QWERTY keys to log, does not read your clipboard, and does not access contacts or any other personal data.

</div>
</details>

</div>
