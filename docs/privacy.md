# Privacy Policy

**Last updated:** February 2025

Diction is built with privacy as a core principle. Here's exactly what happens with your data.

## Self-Hosted Mode

When you use Diction with your own Whisper server:

- **Audio is sent only to your server.** No data touches any third-party service.
- **Nothing is stored.** Audio is processed in memory and discarded immediately after transcription.
- **We have no access** to your audio, transcriptions, or any other data.

You are in complete control.

## Cloud Mode (Diction Cloud)

When you use the hosted Diction Cloud endpoint:

- **Audio is processed and immediately discarded.** We do not store recordings.
- **Transcriptions are not stored.** The text is returned to your device and deleted from server memory.
- **No training data.** Your audio is never used to train or improve any models.

## What We Don't Do

- No analytics or tracking SDKs
- No user behavior tracking
- No data collection of any kind
- No advertising
- No selling or sharing of data with third parties

## Keyboard Extension & Full Access

Diction requests **Full Access** for the keyboard extension. This is required by iOS for any keyboard that needs network access. Here's what Full Access means for Diction:

- **Network access** — the keyboard needs to reach the Whisper endpoint to transcribe audio
- **No keylogging** — Diction has no QWERTY keyboard, no text input to log
- **No clipboard access** — Diction does not read your clipboard
- **No contact access** — Diction does not access your contacts or any other personal data

## Data Flow

```
Your voice → iPhone mic → Diction keyboard → Your Whisper server → Transcribed text → Your app
                                              (or Diction Cloud)
```

That's it. No side channels, no analytics endpoints, no tracking pixels.

## Contact

Questions about privacy? Open an issue on [GitHub](https://github.com/omachala/diction/issues) or email [ondrej@diction.one](mailto:ondrej@diction.one).
