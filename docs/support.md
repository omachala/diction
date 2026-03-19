---
title: Support
description: Get help with Diction voice keyboard for iPhone. Setup guide, troubleshooting, and contact information.
---

# Support

## Getting Started

Setting up Diction takes under a minute:

1. **Open the Diction app** and grant microphone permission when prompted.
2. Go to **Settings > General > Keyboard > Keyboards > Add New Keyboard** and select **Diction**.
3. Tap **Diction** in the keyboard list and enable **Allow Full Access** (required for transcription).
4. Open any app with a text field, tap the **globe icon** to switch to Diction, and tap the mic.

::: tip
The Diction app must be running in the background for the keyboard to work. Launch it once and it stays ready. After a period of inactivity, you may need to open the app again.
:::

## Troubleshooting

### Keyboard does not appear

Make sure Diction is added in **Settings > General > Keyboard > Keyboards**. If it still does not appear, restart your iPhone.

### Microphone not working

Open the Diction app and grant microphone permission. The keyboard extension cannot request mic access on its own. Permission must be granted through the main app first.

### "Open Diction to start" message

The Diction app needs to be running in the background. Open the app, then switch back to your text field and try again.

### "Enable Full Access" message

Go to **Settings > General > Keyboard > Keyboards > Diction** and enable **Allow Full Access**. iOS requires this for any keyboard that uses network access.

### On-device transcription not working

Make sure you have downloaded a speech model in the Diction app. The Standard model downloads automatically on first launch. Check the On-Device section in the app to confirm.

### Transcription fails or times out

Check your internet connection (not needed for on-device mode). If you are using a self-hosted server, verify your endpoint URL is correct and the server is reachable from your phone's network.

## Self-Hosting

Diction works with any speech-to-text server that supports the standard transcription API format. For setup guides, Docker Compose files, and documentation, see the [GitHub repository](https://github.com/omachala/diction).

## Managing Your Subscription

To manage or cancel your Diction One subscription:

**Settings > Apple ID > Subscriptions** on your iPhone.

You can cancel at any time. Your subscription remains active until the end of the current billing period.

## Contact

Need help? Reach out:

- **Email:** [support@diction.one](mailto:support@diction.one)
- **GitHub:** [Open an issue](https://github.com/omachala/diction/issues)
