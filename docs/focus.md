---
title: "Focus Over Spread"
description: Why Diction only ships for iOS. The case against spreading thin across every platform and why deep beats wide for voice input.
keywords: "ios only, focused product, voice keyboard ios only, why not mac, why not android, depth over breadth"
---

# Focus Over Spread

Every other voice tool ships for everything. Mac. Windows. Android. iOS. A browser extension. A desktop overlay. A CLI. An MCP server. Some of them ship for platforms most of their users will never touch. They go wide because going wide looks impressive on a homepage.

Diction goes the other way. iOS keyboard. That's all.

## The math of going wide

Voice input is hard. The hard parts are different on every platform.

- On Mac you fight the accessibility APIs, system audio, and overlay rendering.
- On Windows you fight three different accessibility frameworks and a fragmented audio stack.
- On Android you fight 12,000 device variations, vendor-specific audio drivers, and a permissions model that changes every release.
- On iOS you fight the keyboard sandbox, the mic sandbox, the 48 MB ceiling, and the text proxy.
- On the web you fight browser audio APIs that change quarterly.

A team that ships on all of them is doing five different jobs. Each one demands deep knowledge of the platform. Each one has its own bug surface. Each one breaks when the OS updates. The same engineer who fixed the iOS keyboard yesterday is fighting a Windows audio driver today and a Mac overlay tomorrow.

The result is predictable. The platform the team uses most works well. The rest are rough. The iOS app is usually the roughest, because iOS is the hardest surface and the team has the least time to spend on it. Read reviews of any cross-platform dictation tool and the pattern is consistent: the Mac version works, the iOS version is half-finished.

That math means a focused tool wins on the platform it focuses on. Always. Because the focused tool spends 100% of its time on the surface the broad tool is treating as a side project.

## What we don't build

A short list, kept as a forcing function.

- **No Mac app.** Mac dictation is a different product. Different surfaces, different competitors, different problem.
- **No Android app.** Same logic. Android voice has its own ecosystem and its own gaps. We don't pretend to know them.
- **No web version.** Browser audio is a different beast and a different audience.
- **No desktop overlay.** The overlay model is what Wispr Flow and Willow built. It's fine for them. It's not what we are.
- **No agent mode or MCP server.** Voice input feeds agents. We don't host them.

When something on this list starts to look tempting, the question to ask is: would building this make the iPhone keyboard better? If no, it doesn't get built.

## Why iOS

iOS is the hardest surface. That's the reason.

The hardest surface is where the biggest gap lives. Mac dictation is mostly solved by the existing players. Windows has a strong native dictation. Android has Google's voice typing baked in. iOS has Apple's dictation, which is unreliable, and a market of dictation apps that mostly skip the keyboard layer because it's hard.

The keyboard layer is the layer that matters. Anywhere you type, the keyboard is what you reach for. Voice input that lives anywhere else is a context switch the user has to opt into. Voice input that lives in the keyboard is always there.

Diction is for the user who already wants voice on their phone and is tired of the trade-offs. The keyboard you already use, replaced with one that works.

## What focus buys you

A few concrete things, all of which are the result of one team building one product for one surface.

- **The keyboard fits in 15 MB.** It loads instantly. It doesn't lag when you open a text field. Most apps in this category are an order of magnitude bigger because they ship with code for surfaces they don't need on iOS.
- **The keyboard survives app switches.** You can dictate, hit home, swap apps, come back. The recording is intact. This works because we spent the engineering hours on the iOS lifecycle instead of the Windows audio driver.
- **The keyboard works in apps no other tool touches.** Terminal apps. SSH clients. Bank chat. The kind of fields the standard iOS keyboard already struggles with. Diction works in them because we tested them.
- **The audio pipeline is custom.** Background noise, mic warmup, clipping, silence handling, voice isolation. Each one tuned for the iPhone microphone and the iOS audio session, not a generic cross-platform abstraction.
- **The keyboard polish keeps going.** Long-press behavior, haptics, key sizing, dark mode, accessibility. Small things that compound. None of it would happen if half the engineering budget went to the Mac version.

The list is not exhaustive. It's the surface of a longer one. The point is the same: focus turns into things the user actually feels.

---

This focus is the bet. So far it has been the right one.

For what the bet looks like on the inside, read [How Diction Is Built](/engineering).

[Download for iOS](https://apps.apple.com/app/id6759807364)
