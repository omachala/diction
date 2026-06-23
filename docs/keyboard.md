---
title: "The Keyboard Problem"
description: Why the iPhone keyboard is the right place to put a voice bridge to AI. The keyboard is the one input surface that touches every app you use.
keywords: "iphone keyboard, voice keyboard ios, keyboard extension, qwerty problem, voice input, dictation keyboard"
---

# The Keyboard Problem

QWERTY is from 1873. It was designed for mechanical typewriters with a specific physical problem: stop the hammers from jamming. The layout was a workaround for a constraint that hasn't existed for fifty years.

On a glass touchscreen, it's worse than a workaround. Your fingers cover the letters you're trying to read. The keys are smaller than the tips that hit them. Autocorrect fights you because it can't see what you meant. You type a sentence, look up, find three typos, and fix them with the same wrong fingers. The phone is fast. You are slow. The keyboard is the bottleneck.

## Why the keyboard, not an app

There are dictation apps. Plenty of them. Open them up, talk into the mic, copy the text, switch back to your real app, paste it, fix what didn't come through. By the time you're done, you could have typed it.

The keyboard is different. iOS lets you replace it system-wide. Once installed, your voice button shows up in every text field, in every app, including the ones nobody else can reach — Slack, your bank's chat, the Terminal app, an SSH client, a notes app you found yesterday. The keyboard is the universal input surface of the phone. Anything that takes typed text takes voice the moment you change the keyboard.

That is the only reason to build this as a keyboard. If voice lived in a separate app, you'd switch apps to use it. Most people won't. The friction is too high. So voice never becomes the default input. Putting voice in the keyboard removes the friction. You don't switch. You just talk into whatever you're already using.

## What that takes

Building inside the iOS keyboard extension is harder than building a normal app. The system gives the keyboard a tight box to live in.

- **Memory ceiling.** Keyboards get cut off around 48 MB on most devices. A normal app gets gigabytes. Speech models, audio buffers, language data — all of it has to fit in a closet.
- **Cold start budget.** When you tap into a text field, the keyboard has milliseconds to show up. If it stalls, iOS kills it and falls back to the default. There is no time to warm up.
- **Mic sandbox.** The mic has rules. The orange dot tells the user the mic is live. Background processes can lose access mid-recording. Voice isolation and Bluetooth handoff change the audio pipeline without warning.
- **Text proxy quirks.** The way the keyboard reads and writes text into a field varies app by app. Some apps strip newlines. Some break selection. Some lie about cursor position. The keyboard has to work around all of them.
- **Survives the app switch.** People dictate, hit home, look something up, come back. The keyboard has to be where they left it, with the recording intact.

Most voice tools don't deal with any of this because they don't try to live inside the keyboard. They live in a separate app or a Mac overlay where the rules are easier. The price of that easier path is a worse experience for the user, who now has to switch contexts every time they want to talk to their phone.

Diction takes the harder path because the harder path is the only one that puts the bridge where the user actually is.

## The button

One button. Tap it. Speak. The text appears where your cursor is.

That's it. No layout to learn. No keys to hunt. No autocorrect war. The intelligence is in the audio pipeline, the speech model, and the cleanup pass. The interface is a single button you already know how to use.

When you can't speak — quiet room, public space, a colleague at the next desk — the button stays out of your way and you can drop back to whatever fallback input you prefer. The point isn't to ban typing. The point is that when voice is available, it should be the first thing you reach for, and reaching for it should take exactly one tap inside the app you're already in.

That's the keyboard problem. Diction is the answer.

---

For why this is the only thing we're building, read [Focus Over Spread](/focus). For the engineering underneath, read [How Diction Is Built](/engineering).

[Download for iOS](https://apps.apple.com/app/id6759807364)
