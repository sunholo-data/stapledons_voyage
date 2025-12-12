# AI Capabilities Reference

**Status**: Active (Updated 2025-12-12)
**Purpose**: Complete reference for AI features available to AILANG game code
**Implementation**: `engine/handlers/ai_gemini*.go`

---

## Quick Reference

| Capability | Trigger | Model | Region |
|------------|---------|-------|--------|
| Text Generation | default | gemini-2.5-flash | europe-west1 |
| Image Generation | `imagen:` prefix | gemini-2.5-flash-image | europe-west1 |
| Image Editing | `edit:` prefix | gemini-2.5-flash-image | europe-west1 |
| Text-to-Speech | `speak:` prefix | gemini-2.5-flash-tts | **us-central1** |

---

## 1. Text Generation

Standard text prompts for NPC dialogue, Archive responses, lore generation.

**Usage:**
```ailang
import std/ai (ai_call)

func ask_archive(question: string) -> string ! {AI} {
    let req = encodeJson({
        "messages": [{"type": "text", "text": question}],
        "system": "You are the Archive, an ancient AI aboard The Spire..."
    });
    let resp = ai_call(req);
    decodeJson(resp).content[0].text
}
```

**Response Format:**
```json
{
  "content": [{"type": "text", "text": "The response..."}],
  "error": ""
}
```

---

## 2. Image Generation

Generate images from text prompts using Gemini 2.5 Flash Image (Nano Banana).

**Trigger:** Prefix prompt with `imagen:`, `generate image`, or `draw`

**Usage:**
```ailang
func generate_portrait(description: string) -> string ! {AI} {
    let req = encodeJson({
        "messages": [{"type": "text", "text": "imagen: " ++ description}],
        "context": {"generate_image": true}
    });
    let resp = ai_call(req);
    decodeJson(resp).content[0].image_ref  -- Returns path to saved image
}
```

### Aspect Ratios

| Ratio | Dimensions | Use Case |
|-------|------------|----------|
| `1:1` | Square | Portraits, icons |
| `2:3` | Portrait | Character art |
| `3:2` | Landscape | Scene backgrounds |
| `3:4` | Portrait | UI panels |
| `4:3` | Landscape | Classic TV ratio |
| `9:16` | Tall | Mobile, vertical scenes |
| `16:9` | Wide | Cinematic, panoramas |
| `21:9` | Ultrawide | Starfield backgrounds |

### Resolution Options

| Resolution | Description |
|------------|-------------|
| 1K | Default (1024px) |
| 2K | High quality |
| 4K | Maximum (Pro model only) |

### Capabilities

- **Text-to-image**: Natural language descriptions
- **Multi-image blending**: Combine multiple reference images
- **Character consistency**: Maintain appearance across generations
- **High-fidelity text**: Legible text in images (logos, signs)
- **Style control**: Photorealistic, artistic, anime, etc.

---

## 3. Image Editing

Edit existing images with natural language instructions.

**Trigger:** Prefix with `edit:` or `modify:`

**Usage:**
```ailang
func edit_scene(image_path: string, instruction: string) -> string ! {AI} {
    let req = encodeJson({
        "messages": [
            {"type": "image", "image_ref": image_path},
            {"type": "text", "text": "edit: " ++ instruction}
        ],
        "context": {"edit_image": true, "reference_image": image_path}
    });
    let resp = ai_call(req);
    decodeJson(resp).content[0].image_ref
}
```

### Editing Capabilities

| Operation | Example Prompt |
|-----------|----------------|
| Add elements | "Add a glowing artifact in the corner" |
| Remove elements | "Remove the person from the background" |
| Style transfer | "Make this look like an oil painting" |
| Color grading | "Add a blue sci-fi color grade" |
| Background blur | "Blur the background" |
| Pose adjustment | "Change the character to face left" |
| Colorization | "Add color to this black and white image" |
| Text overlay | "Add the text 'ARCHIVE' at the top" |

---

## 4. Text-to-Speech (TTS)

Generate spoken audio from text. **Requires us-central1 region** (auto-configured).

**Trigger:** Prefix with `speak:`, `say:`, or `tts:`

**Usage:**
```ailang
func speak_greeting(text: string, voice: string) -> string ! {AI} {
    let req = encodeJson({
        "messages": [{"type": "text", "text": "speak: " ++ text}],
        "context": {"tts": true, "voice": voice}
    });
    let resp = ai_call(req);
    decodeJson(resp).content[0].audio_ref  -- Returns path to WAV file
}
```

### Available Voices (30 total)

| Voice | Characteristic | Suggested Use |
|-------|---------------|---------------|
| **Puck** | Upbeat | Friendly NPCs, guides |
| **Kore** | Firm | Authority figures, commanders |
| **Charon** | Informative | Archive, AI assistants |
| **Zephyr** | Bright | Young characters, announcements |
| **Fenrir** | Excitable | Enthusiastic NPCs, discoveries |
| **Aoede** | Breezy | Casual dialogue, merchants |
| **Enceladus** | Breathy | Mysterious, ethereal beings |
| **Leda** | Youthful | Young characters, children |
| **Orus** | Firm | Military, officials |
| **Callirrhoe** | Easy-going | Relaxed NPCs, travelers |
| **Autonoe** | Bright | Optimistic characters |
| **Iapetus** | Clear | Narration, logs |
| **Umbriel** | Easy-going | Casual NPCs |
| **Algieba** | Smooth | Diplomats, negotiators |
| **Despina** | Smooth | Professional, corporate |
| **Erinome** | Clear | Technical readouts |
| **Algenib** | Gravelly | Gruff characters, veterans |
| **Rasalgethi** | Informative | Scientists, researchers |
| **Laomedeia** | Upbeat | Cheerful NPCs |
| **Achernar** | Soft | Whispers, intimate moments |
| **Alnilam** | Firm | Stern characters |
| **Schedar** | Even | Neutral narration |
| **Gacrux** | Mature | Elder characters, wisdom |
| **Pulcherrima** | Forward | Confident characters |
| **Achird** | Friendly | Welcoming NPCs |
| **Zubenelgenubi** | Casual | Laid-back dialogue |
| **Vindemiatrix** | Gentle | Caring characters, medics |
| **Sadachbia** | Lively | Energetic NPCs |
| **Sadaltager** | Knowledgeable | Experts, historians |
| **Sulafat** | Warm | Comforting characters |

### Voice Selection by Character Type

| Character Type | Recommended Voices |
|----------------|-------------------|
| The Archive (main AI) | Charon, Rasalgethi, Iapetus |
| Ship's Captain | Kore, Orus, Alnilam |
| Alien Elder | Gacrux, Enceladus, Achernar |
| Crew Member (friendly) | Puck, Achird, Aoede |
| Crew Member (technical) | Erinome, Despina, Iapetus |
| Mysterious Entity | Enceladus, Achernar, Vindemiatrix |
| Merchant/Trader | Aoede, Algieba, Zubenelgenubi |
| Military Officer | Kore, Orus, Algenib |
| Scientist | Rasalgethi, Sadaltager, Charon |
| Child/Youth | Leda, Zephyr, Autonoe |

### Voice Variation & Expression Control

Gemini TTS supports multiple methods to control voice expression, emotion, and delivery style.

#### 1. Style Prompts (System Prompt)

Set the overall tone by describing the desired delivery style:

```ailang
func speak_with_style(text: string, voice: string, style: string) -> string ! {AI} {
    let req = encodeJson({
        "messages": [{"type": "text", "text": "speak: " ++ text}],
        "system": style,
        "context": {"tts": true, "voice": voice}
    });
    decodeJson(ai_call(req)).content[0].audio_ref
}

-- Examples:
speak_with_style("The artifact is unstable.", "Charon", "Speak in a calm, measured tone with slight concern")
speak_with_style("We found it!", "Fenrir", "Speak with excitement and wonder")
speak_with_style("The council has decided.", "Kore", "Speak formally and authoritatively")
```

**Effective Style Descriptions:**
| Style | Description |
|-------|-------------|
| Urgent | "Speak quickly with urgency and tension" |
| Mysterious | "Speak slowly with a hushed, mysterious tone" |
| Joyful | "Speak with warmth and happiness" |
| Grave | "Speak solemnly with weight and gravity" |
| Sarcastic | "Speak with dry sarcasm and subtle mockery" |
| Exhausted | "Speak tiredly, as if out of breath" |

#### 2. Emotion Markers `[]`

Insert emotion cues directly in the text using square brackets:

```ailang
-- Emotion markers in text
let text = "[sigh] I've been searching for centuries. [pause] And now... [excited] we finally found it!";
```

**Supported Emotion Markers:**
| Marker | Effect |
|--------|--------|
| `[sigh]` | Audible sigh |
| `[laugh]` | Light laughter |
| `[gasp]` | Surprised gasp |
| `[whisper]` | Whispered delivery |
| `[pause]` | Brief pause |
| `[excited]` | Excited tone shift |
| `[sad]` | Somber tone shift |
| `[angry]` | Tense/angry tone |

#### 3. SSML Tags

SSML (Speech Synthesis Markup Language) provides fine-grained control:

```ailang
-- SSML example
let ssml_text = "<speak>Welcome aboard. <break time=\"500ms\"/> The Archive awaits your questions. <prosody rate=\"slow\" pitch=\"low\">Choose wisely.</prosody></speak>";
```

**Supported SSML Tags:**
| Tag | Purpose | Example |
|-----|---------|---------|
| `<break>` | Insert pause | `<break time="500ms"/>` or `<break strength="medium"/>` |
| `<prosody>` | Control rate, pitch, volume | `<prosody rate="slow" pitch="high">text</prosody>` |
| `<emphasis>` | Add emphasis | `<emphasis level="strong">important</emphasis>` |
| `<say-as>` | Interpret as type | `<say-as interpret-as="characters">AI</say-as>` |

**Prosody Attributes:**
- `rate`: x-slow, slow, medium, fast, x-fast, or percentage (80%, 120%)
- `pitch`: x-low, low, medium, high, x-high, or semitones (+2st, -3st)
- `volume`: silent, x-soft, soft, medium, loud, x-loud, or dB (+6dB)

#### 4. Speaker Vocalizations

The model can produce natural vocalizations when prompted:

```ailang
-- Works: Speaker sounds
"Hmm... let me think about that."     -- Thinking sound
"Ha! You really believed that?"       -- Laughter
"*sigh* If I must explain again..."   -- Sigh
"Shh... they might hear us."          -- Whisper/hush
```

**What Works:**
- Laughter (ha, haha, heh)
- Sighs and exhales
- Thinking sounds (hmm, uh, um)
- Whispers and hushed speech
- Gasps and exclamations

#### 5. Sound Effects Limitations

**⚠️ Environmental sound effects are NOT supported.** Gemini TTS is a speech synthesis model, not a sound effects generator.

**Does NOT work:**
- Ambient sounds (wind, rain, machinery)
- Music or musical tones
- Non-vocal sounds (explosions, doors, alarms)
- Animal sounds (unless mimicked by voice)

**For sound effects, use pre-recorded audio files** loaded through the engine's audio system instead.

---

### Supported Languages

TTS supports 24+ languages with automatic language detection:

| Code | Language | Code | Language |
|------|----------|------|----------|
| en-US | English (US) | ja-JP | Japanese |
| en-GB | English (UK) | ko-KR | Korean |
| en-AU | English (AU) | cmn-CN | Mandarin |
| de-DE | German | ru-RU | Russian |
| fr-FR | French | pt-BR | Portuguese (BR) |
| es-ES | Spanish | it-IT | Italian |
| hi-IN | Hindi | nl-NL | Dutch |
| ar-XA | Arabic | pl-PL | Polish |
| tr-TR | Turkish | th-TH | Thai |
| vi-VN | Vietnamese | id-ID | Indonesian |

---

## 5. Audio Output Format

TTS returns raw PCM audio that is automatically converted to WAV:

- **Format**: WAV (RIFF)
- **Sample Rate**: 24000 Hz
- **Bit Depth**: 16-bit
- **Channels**: Mono
- **Location**: `assets/generated/response_*.wav`

---

## 6. Models Summary

| Model | Purpose | Notes |
|-------|---------|-------|
| `gemini-2.5-flash` | Text generation | Default, fast |
| `gemini-2.5-flash-image` | Image gen/edit | "Nano Banana" |
| `gemini-2.5-flash-tts` | Text-to-speech | Fast, good quality |
| `gemini-2.5-pro-tts` | Text-to-speech | Higher quality |
| `gemini-3-pro-image-preview` | Advanced images | 4K, thinking mode |

---

## 7. Configuration (Go side)

```go
handler, err := handlers.NewGeminiAIHandler(ctx, handlers.GeminiConfig{
    // Models
    Model:       "gemini-2.5-flash",           // Text model
    ImagenModel: "gemini-2.5-flash-image",     // Image model
    TTSModel:    "gemini-2.5-flash-tts",       // TTS model
    TTSVoice:    "Charon",                     // Default voice

    // Locations
    Project:  "your-gcp-project",
    Location: "europe-west1",  // Main region
    // TTS auto-creates us-central1 client

    // Output
    AssetsDir: "assets/generated",
})
```

---

## 8. Error Handling

```ailang
func safe_ai_call(req: string) -> Option[string] ! {AI} {
    let resp = ai_call(req);
    let parsed = decodeJson(resp);
    match parsed.error {
        "" => Some(parsed.content[0].text),
        _  => None
    }
}
```

---

## 9. Game Integration Examples

### Archive Query System
```ailang
type ArchiveQuery = {
    question: string,
    context: string,
    voice: string
}

func query_archive(q: ArchiveQuery) -> (string, string) ! {AI} {
    -- Get text response
    let text_req = encodeJson({
        "messages": [{"type": "text", "text": q.question}],
        "system": "You are the Archive. Context: " ++ q.context
    });
    let text_resp = decodeJson(ai_call(text_req)).content[0].text;

    -- Generate speech
    let audio_req = encodeJson({
        "messages": [{"type": "text", "text": "speak: " ++ text_resp}],
        "context": {"tts": true, "voice": q.voice}
    });
    let audio_path = decodeJson(ai_call(audio_req)).content[0].audio_ref;

    (text_resp, audio_path)
}
```

### Dynamic NPC Portrait
```ailang
func generate_npc_portrait(species: string, age: string, mood: string) -> string ! {AI} {
    let prompt = "Portrait of a " ++ age ++ " " ++ species ++
                 " alien, " ++ mood ++ " expression, sci-fi style, 1:1 aspect ratio";
    let req = encodeJson({
        "messages": [{"type": "text", "text": "imagen: " ++ prompt}],
        "context": {"generate_image": true}
    });
    decodeJson(ai_call(req)).content[0].image_ref
}
```

---

## Sources

- [Gemini TTS Documentation](https://ai.google.dev/gemini-api/docs/speech-generation)
- [Gemini TTS Voice Control](https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/speech-generation) - Style prompts, SSML
- [Gemini Image Generation](https://ai.google.dev/gemini-api/docs/image-generation)
- [Imagen 3 API](https://ai.google.dev/gemini-api/docs/imagen)
- [Gemini 2.5 Flash Image Announcement](https://developers.googleblog.com/en/introducing-gemini-2-5-flash-image/)

---

**Document created**: 2025-12-12
**Last updated**: 2025-12-12 (added voice variation controls)
