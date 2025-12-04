# AI Audio Generation - Current Blockers

**Date:** 2024-12-04
**Status:** Blocked - awaiting access

## Summary

Text and image generation work. Audio generation (TTS and music) requires additional access.

## What's Working

| Feature | Model | Status |
|---------|-------|--------|
| Text chat | `gemini-2.5-flash` | ✅ Working |
| Image recognition | `gemini-2.5-flash` | ✅ Working |
| Image generation | `gemini-2.5-flash-image` | ✅ Working |

## What's Blocked

### 1. Text-to-Speech (TTS)

**Error:** `You are not allowlisted to request audio output`

**Models tried:**
- `gemini-2.5-flash-tts` - 404 Not Found
- `gemini-2.5-flash` with `ResponseModalities: ["AUDIO"]` - Not allowlisted

**Resolution:** Request allowlisting via:
- https://discuss.ai.google.dev/t/request-allowlist-access-for-audio-output-in-gemini-2-5-pro-flash-tts-vertex-ai/108067
- Or try using Gemini API (not Vertex AI) with `GOOGLE_API_KEY`

**Code ready:** Yes - implementation in `engine/handlers/ai_gemini.go:textToSpeech()`

### 2. Music Generation (Lyria)

**Issue:** No Go SDK support

**Details:**
- Lyria RealTime uses WebSocket streaming
- Only Python and JavaScript SDKs have native support
- Would require implementing raw WebSocket client

**WebSocket endpoint:**
```
wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1alpha.GenerativeService.BidiGenerateMusic
```

**Resolution options:**
1. Wait for Go SDK to add Lyria support
2. Implement raw WebSocket client (complex)
3. Use pre-recorded audio assets (recommended for now)

## CLI Commands (for testing when access is granted)

```bash
# List voices
./bin/voyage ai -list-voices

# Test TTS
./bin/voyage ai -tts -prompt "Hello world"
./bin/voyage ai -tts -voice Puck -prompt "Hello world"

# Available voices: Aoede, Charon, Fenrir, Kore (default), Puck, Zephyr, Enceladus
```

## Environment Variables

```bash
# TTS model override (if different model becomes available)
export GEMINI_TTS_MODEL=gemini-2.5-flash-tts
export GEMINI_TTS_VOICE=Kore
```

## Next Steps

1. **Short term:** Use pre-recorded audio assets for game SFX
2. **Medium term:** Request TTS allowlisting for Vertex AI project
3. **Long term:** Implement Lyria WebSocket client when Go SDK support is available

## References

- [Gemini TTS Docs](https://ai.google.dev/gemini-api/docs/speech-generation)
- [Lyria Music Generation](https://ai.google.dev/gemini-api/docs/music-generation)
- [Vertex AI Allowlist Request](https://discuss.ai.google.dev/t/request-allowlist-access-for-audio-output-in-gemini-2-5-pro-flash-tts-vertex-ai/108067)
