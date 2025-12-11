# Dialogue Text-to-Speech Voices

**Status**: Planned
**Target**: v0.6.0 (with Dialogue System)
**Priority**: P1 (Enhances Core Interaction)
**Estimated**: 3 days
**Dependencies**: [dialogue-system.md](future/dialogue-system.md), [ai-handler-system.md](../implemented/v0_1_0/ai-handler-system.md)

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | 0 | 0 | Voice doesn't affect time mechanics |
| Civilization Simulation | + | +1 | Alien civs get distinct voices, enhances first contact |
| Philosophical Depth | + | +1 | Voice tone conveys character philosophy in debates |
| Ship & Crew Life | + | +1 | Crew voices build emotional connection before deaths |
| Legacy Impact | 0 | 0 | No direct legacy contribution |
| Hard Sci-Fi Authenticity | N/A | 0 | Audio feature, no physics implications |
| **Net Score** | | **+3** | **Decision: Move forward** |

**Feature type:** Gameplay
- Enhances player emotional connection to crew and civilizations
- Makes dialogue sequences more immersive
- Critical for memorial scenes (crew deaths during journey)

**Reference:** See [game-vision.md](../../docs/game-vision.md)

## Problem Statement

Voice brings characters to life. When crew members speak, players should *hear* them - creating emotional bonds that make journey deaths meaningful.

**Current State:**
- Text-only dialogue (when dialogue system is implemented)
- AI handler already supports TTS via Gemini (6 voices available)
- No voice assignment per character
- No audio caching for dialogue

**Impact:**
- Dialogue feels flat without voice
- Emotional weight of crew deaths diminished
- First contact with aliens loses gravitas
- Journey events (philosophical debates) lack impact

## Goals

**Primary Goal:** Every speaking character has a distinct voice that plays during dialogue.

**Success Metrics:**
- 100% of crew archetypes mapped to voices
- Audio playback latency < 200ms for cached lines
- Seamless integration with dialogue system
- Voice consistency across character lifetime

## Solution Design

### Overview

Integrate Gemini TTS into the dialogue system with:
1. **Voice profiles** per character archetype
2. **Pre-generation** for scripted dialogue nodes
3. **Real-time generation** for AI-driven dialogue
4. **Audio caching** to avoid repeated API calls

### Architecture

```
DialogueNode (AILANG)
    │
    ├── speakerID → VoiceProfile lookup
    ├── text → TTS input
    │
    ▼
VoiceManager (Go)
    │
    ├── Check cache (hash of speaker+text)
    ├── If miss: Call Gemini TTS
    ├── Save to assets/generated/voices/
    │
    ▼
AudioPlayer (existing)
    │
    └── Play WAV/OGG during dialogue
```

**Components:**

1. **VoiceProfile** (AILANG): Maps character archetypes to Gemini voice names
2. **VoiceManager** (Go): Handles TTS generation, caching, and playback
3. **DialogueTTSIntegration**: Hooks voice into dialogue state machine

### Voice Assignment by Archetype

Using archetypes from [crew-psychology.md](future/crew-psychology.md) and Gemini's available voices:

| Archetype | Voice | Rationale |
|-----------|-------|-----------|
| **Engineer** | Charon | Low, measured, laconic |
| **Scientist** | Kore | Clear, analytical |
| **Medic** | Aoede | Warm, reassuring |
| **Diplomat** | Zephyr | Smooth, persuasive |
| **Pilot** | Fenrir | Bold, decisive |
| **Quartermaster** | Charon | Gruff, no-nonsense |
| **Zealot** | Puck | Passionate, intense |
| **Dreamer** | Aoede | Soft, contemplative |
| **Skeptic** | Kore | Questioning tone |
| **Fantasist** | Puck | Energetic, playful |
| **Analyst** | Kore | Precise, measured |
| **Alien (default)** | Fenrir | Otherworldly, deep |

**Available Gemini Voices:** Aoede, Charon, Fenrir, Kore, Puck, Zephyr

### AILANG Types

```ailang
module sim/voice

-- Voice profile for a character
type VoiceProfile = {
    voiceName: string,      -- Gemini voice: "Kore", "Charon", etc.
    pitch: float,           -- Pitch modifier (0.8 - 1.2)
    speed: float,           -- Speed modifier (0.8 - 1.2)
    emotionalRange: float   -- How much emotion affects voice (0.0 - 1.0)
}

-- Map archetype to voice
pure func archetypeVoice(archetype: CrewArchetype) -> VoiceProfile {
    match archetype {
        Engineer => { voiceName: "Charon", pitch: 0.95, speed: 0.9, emotionalRange: 0.3 },
        Scientist => { voiceName: "Kore", pitch: 1.0, speed: 1.0, emotionalRange: 0.5 },
        Medic => { voiceName: "Aoede", pitch: 1.05, speed: 0.95, emotionalRange: 0.8 },
        Diplomat => { voiceName: "Zephyr", pitch: 1.0, speed: 1.0, emotionalRange: 0.7 },
        Pilot => { voiceName: "Fenrir", pitch: 0.9, speed: 1.1, emotionalRange: 0.6 },
        Zealot => { voiceName: "Puck", pitch: 1.0, speed: 1.05, emotionalRange: 0.9 },
        Dreamer => { voiceName: "Aoede", pitch: 1.1, speed: 0.85, emotionalRange: 0.9 },
        _ => { voiceName: "Kore", pitch: 1.0, speed: 1.0, emotionalRange: 0.5 }
    }
}

-- Request TTS for a dialogue line
func requestVoice(speaker: CrewID, text: string, emotion: Emotion) -> VoiceRequest ! {AI} {
    let profile = getCrewVoiceProfile(speaker);
    VoiceRequest {
        speakerID: speaker,
        text: text,
        voice: profile.voiceName,
        emotion: emotion
    }
}
```

### Go Integration

```go
// engine/handlers/voice_manager.go

type VoiceManager struct {
    aiHandler  AIHandler
    cache      map[string]string  // hash -> audio file path
    cacheDir   string
}

func (vm *VoiceManager) GetOrGenerate(speaker, text, voice string) (string, error) {
    // Create cache key from speaker + text hash
    key := hashVoiceRequest(speaker, text, voice)

    if path, ok := vm.cache[key]; ok {
        return path, nil  // Cache hit
    }

    // Generate via Gemini TTS
    request := fmt.Sprintf(`{"tts": true, "voice": "%s", "text": "%s"}`, voice, text)
    response, err := vm.aiHandler.Call(request)
    if err != nil {
        return "", err
    }

    // Parse response, get audio path
    audioPath := parseAudioPath(response)
    vm.cache[key] = audioPath

    return audioPath, nil
}
```

### Implementation Plan

**Phase 1: Voice Profile System** (~4 hours)
- [ ] Define VoiceProfile type in AILANG
- [ ] Create archetype → voice mapping
- [ ] Add voice field to CrewMember type
- [ ] Test with hardcoded dialogue lines

**Phase 2: Voice Manager** (~6 hours)
- [ ] Create VoiceManager in engine/handlers/
- [ ] Implement caching with content hash
- [ ] Integrate with existing Gemini TTS
- [ ] Add pre-generation script for scripted dialogue

**Phase 3: Dialogue Integration** (~6 hours)
- [ ] Hook VoiceManager into dialogue state machine
- [ ] Play audio during dialogue display
- [ ] Handle async generation (show text while generating)
- [ ] Add volume/mute controls

**Phase 4: Polish** (~4 hours)
- [ ] Add emotion modulation (sad, angry, excited)
- [ ] Implement alien voice processing (pitch shift)
- [ ] Add subtitle sync for hearing impaired
- [ ] Test all crew archetypes

### Files to Modify/Create

**New files:**
- `sim/voice.ail` - Voice profiles, archetype mapping (~100 LOC)
- `engine/handlers/voice_manager.go` - TTS caching and playback (~200 LOC)

**Modified files:**
- `sim/dialogue.ail` - Add voice requests to dialogue nodes (~50 LOC)
- `engine/handlers/ai_gemini.go` - Ensure TTS returns proper path (~20 LOC)
- `engine/audio/player.go` - Add dialogue playback queue (~30 LOC)

## Examples

### Example 1: Crew Dialogue

**Before (text only):**
```
┌────────────────────────────────────┐
│ [Chen - Engineer]                  │
│                                    │
│ "The reactor's holding, but        │
│  I don't like the harmonics."      │
│                                    │
│        [Continue]                  │
└────────────────────────────────────┘
```

**After (with voice):**
```
┌────────────────────────────────────┐
│ [Chen - Engineer]     🔊 Playing   │
│                                    │
│ "The reactor's holding, but        │  ← Audio plays (Charon voice)
│  I don't like the harmonics."      │
│                                    │
│        [Continue]                  │
└────────────────────────────────────┘
```

### Example 2: First Contact

```ailang
-- Alien civilization with distinct voice
func alienFirstContactVoice(civ: Civilization) -> VoiceProfile {
    -- Aliens get Fenrir voice with pitch/speed modulation
    {
        voiceName: "Fenrir",
        pitch: 0.7 + (hash(civ.id) % 30) / 100.0,  -- Varies per civ
        speed: 0.8,
        emotionalRange: 0.4  -- More reserved
    }
}
```

### Example 3: Memorial Scene

```ailang
-- Crew death memorial uses their familiar voice
func memorialVoice(deceased: Crew, memory: string) -> [DrawCmd] ! {AI, Audio} {
    let profile = archetypeVoice(deceased.archetype);

    -- Play their voice one last time reading their epitaph
    Audio.play(requestVoice(deceased.id, memory, Melancholy));

    -- Emotional impact through familiar voice saying goodbye
    renderMemorialUI(deceased, memory)
}
```

## Success Criteria

- [ ] All 11 crew archetypes have distinct voice profiles
- [ ] Voice plays during dialogue without blocking UI
- [ ] Audio cached - same line doesn't re-generate
- [ ] Mute/volume controls work
- [ ] Alien civilizations have modified voices
- [ ] Memorial scenes play deceased crew's voice
- [ ] Latency < 500ms for first-time generation
- [ ] Latency < 50ms for cached playback

## Testing Strategy

**Unit tests:**
- Voice profile mapping for all archetypes
- Cache key generation (deterministic)
- Audio path parsing from Gemini response

**Integration tests:**
- Full TTS round-trip with Gemini API
- Cache hit/miss verification
- Audio playback during dialogue

**Manual testing:**
- Listen to each archetype voice
- Verify emotional modulation sounds natural
- Check alien voices feel "other"
- Test memorial scene emotional impact

## Non-Goals

**Not in this feature:**
- Voice cloning for unique crew voices - requires custom training
- Real-time lip sync - too complex for 2D sprites
- Multi-language TTS - English only initially
- Player voice - player remains silent protagonist

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Gemini TTS latency too high | Medium | Pre-generate scripted dialogue; async generation |
| Voice sounds robotic | Medium | Test all 6 voices; use emotion modulation |
| Cache size grows large | Low | Limit to recent conversations; cleanup on journey start |
| API costs for TTS | Low | Cache aggressively; batch generation |
| Voice mismatch with portrait | Medium | Careful archetype-to-voice mapping |

## References

- [ai-handler-system.md](../implemented/v0_1_0/ai-handler-system.md) - Existing TTS infrastructure
- [dialogue-system.md](future/dialogue-system.md) - Dialogue state machine
- [crew-psychology.md](future/crew-psychology.md) - Archetype definitions
- [audio-system.md](../implemented/v0_1_0/audio-system.md) - Audio playback
- [Gemini TTS Voices](https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/text-to-speech) - Voice reference

## Future Work

- **Voice aging**: Crew voices deepen/soften over decades
- **Relationship memory**: AI recalls prior conversations in voice
- **Alien language**: Generated alien phonemes before translation
- **Narrator voice**: Archive AI has distinct narrator voice
- **Voice synthesis**: Train custom voices for main characters

---

**Document created**: 2025-12-11
**Last updated**: 2025-12-11
