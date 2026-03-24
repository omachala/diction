---
title: "About Diction"
description: Diction is built by a hands-on engineering lead who works daily on critical banking infrastructure. Privacy and reliability are not product values. They are engineering constraints he lives by.
keywords: "about diction, who built diction, diction developer, voice keyboard engineer"
---

# About

Diction is an engineering project.

I'm [Ondrej Machala](https://www.linkedin.com/in/ondrejmachala). Hands-on engineering lead. My 9-to-5 is building systems at the core of banking infrastructure. The kind where a bug doesn't mean a bad user review. It means real money moves in the wrong direction, or doesn't move at all. Privacy and reliability aren't things I put on a slide. They're constraints I engineer against every single day.

I use speech to text constantly. Commuting. On a walk with my son. Apple's built-in dictation is well known for being unreliable, and I needed something that works in any app. And I mean any app. I use terminal apps, SSH clients, things where the standard iOS keyboard already struggles. I was just missing something that solves the problem properly. A keyboard you tap, speak into, and it works. Every time, everywhere.

So I built it properly. Diction is under 15 MB. I wanted it lean and focused, one thing that works rather than ten things that mostly work. A lot of the core audio and processing modules are custom built because the available libraries either pulled in too many dependencies or weren't precise enough for what I needed. iOS keyboard extensions have tight memory limits and no tolerance for slowness. I've spent a lot of time on memory usage, battery draw, and making sure the keyboard is ready the moment you need it.

The same instincts I bring to banking systems are in Diction. No unnecessary data retention. No assumptions about the happy path. Defence in depth. Audio is processed and immediately discarded. The server is open source. You can read every line.

On-device and self-hosted are free. If you don't want to run a server, there's Diction One. It runs on my own infrastructure with fine-tuned speech models and low-latency processing. Transcription text is encrypted with AES-256-GCM before it leaves the server. Context awareness carries terminology and names across your session. AI polishing cleans up the raw transcript before it hits your text field. The result is clean, ready-to-send text with no setup on your end. Priced like a developer built it, not like a startup needs to hit revenue targets.

---

Find me on [GitHub](https://github.com/omachala) or [LinkedIn](https://www.linkedin.com/in/ondrejmachala). Questions and feedback welcome: [reach out.](/support)
