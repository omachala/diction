---
layout: home
title: "Voice Keyboard for iPhone: Speech to Text"
description: Voice keyboard for iPhone with open-source server. On-device, self-hosted, or cloud transcription. No word limits, no tracking, 99 languages. Free forever on your own server.

hero:
  name: Diction
  text: "Powerful<br>speech to text<br>iPhone keyboard"
  tagline: "Your voice. Your data. On your device or self-hosted.<br>No limits · Private · Encrypted · 99 languages"
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
      src: /icon-infinity.svg
    title: No Limits
    details: No weekly caps, no word restrictions, no catch. Dictate as much as you want, whenever you want.
  - icon:
      src: /icon-phone.svg
    title: On-Device
    details: Speech recognition that runs locally on your iPhone. No internet required. Audio never leaves your device.
  - icon:
      src: /icon-github.svg
    title: Self-Hosted
    details: Run licensed or industry-specific models on your own server. Your network, your data.
    link: /features/self-hosting-setup
  - icon:
      src: /icon-shield.svg
    title: Private
    details: AES-256 end-to-end text encryption with X25519 key exchange. Zero analytics, zero tracking, zero data collection.
  - icon:
      src: /icon-globe.svg
    title: Your Language
    details: Automatic multilingual speech recognition. Speak in your language, or translate to another - Diction handles both.
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

## Questions

<details>
<summary>Is it really free?</summary>
<div>

On-device and self-hosted modes are completely free with basic models. No word limits, no daily caps. Pro unlocks premium on-device models for better accuracy, plus Diction One cloud. Free trial included.

</div>
</details>

<details>
<summary>How is it better than Apple Dictation?</summary>
<div>

Diction uses Whisper-based models that are significantly more accurate. No session time limits, no word caps. Works identically across all apps. Choose between on-device, cloud, or your own server.

</div>
</details>

<details>
<summary>Does it work offline?</summary>
<div>

Yes. On-device mode works without internet once the model is downloaded. Cloud and self-hosted modes require network access.

</div>
</details>

<details>
<summary>What languages does Diction support?</summary>
<div>

99 languages via Whisper. On-device base model handles most languages well. Cloud and self-hosted modes use larger models for even better accuracy across all supported languages.

</div>
</details>

<details>
<summary>Is my voice data stored?</summary>
<div>

Never. On-device mode processes audio in memory and discards it immediately. Self-hosted mode sends audio only to your server - we have no access. Diction One cloud processes and discards. No recordings retained, no model training.

</div>
</details>

<details>
<summary>What is AI Enhancement?</summary>
<div>

After transcription, Diction can optionally clean up your text - removing filler words, fixing grammar, and polishing the result. Only the text is sent to the AI, never the audio. Off by default.

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
