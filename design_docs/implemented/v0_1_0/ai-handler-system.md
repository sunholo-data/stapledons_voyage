# AI Handler System

**Version:** 0.1.0
**Status:** Implemented
**Priority:** P1 (High)
**Complexity:** Medium
**Package:** `engine/handlers`

## Related Documents

- [AI Effect for NPCs](../../reference/ai-effect-npcs.md) - AILANG integration design (planned)
- [AI: The Archive](../../input/ai-the-archive.md) - "Unreliable Archivist" concept
- [Engine Layer Design](engine-layer.md) - Effect handler architecture

## Overview

The AI handler system provides pluggable AI backends for the game. It supports multiple providers (Claude, Gemini, stub) with automatic fallback, multimodal content (text, images, audio), and environment-based configuration.

**What's Implemented:**
- Full Go handler infrastructure with provider abstraction
- Claude API integration (text)
- Gemini API integration (text + image generation + TTS)
- Stub handler for testing with pattern matching
- Factory with auto-detection and chain fallback

**What's Planned:**
- AILANG `AI.decide()` effect integration
- Typed decision contexts (CivContext, CrewContext)
- Philosophy-driven prompts for alien civilizations

## Architecture

### Provider Hierarchy

```
AIHandler interface
    │
    ├── StubAIHandler      (testing, deterministic fallback)
    ├── ClaudeAIHandler    (Anthropic API, text-only)
    └── GeminiAIHandler    (Google API, multimodal)
         ├── Text chat
         ├── Image generation (Imagen)
         └── Text-to-speech (TTS)
```

### Data Flow

```
AILANG AI.decide(input)
    → sim_gen.AIHandler.Call(json)
    → engine/handlers provider
    → API call (Claude/Gemini)
    → JSON response
    → AILANG parses result
```

## Implementation

### Core Types (ai.go)

```go
// Content types for multimodal messages
type ContentType string
const (
    ContentTypeText  ContentType = "text"
    ContentTypeImage ContentType = "image"
    ContentTypeAudio ContentType = "audio"
    ContentTypeVideo ContentType = "video"
)

// ContentBlock represents a piece of content
type ContentBlock struct {
    Type     ContentType `json:"type"`
    Text     string      `json:"text,omitempty"`
    ImageRef string      `json:"image_ref,omitempty"`
    AudioRef string      `json:"audio_ref,omitempty"`
    MimeType string      `json:"mime_type,omitempty"`
    AltText  string      `json:"alt_text,omitempty"`
}

// AIRequest for multimodal input
type AIRequest struct {
    Messages []ContentBlock         `json:"messages"`
    Context  map[string]interface{} `json:"context,omitempty"`
    System   string                 `json:"system,omitempty"`
}

// AIResponse with multimodal output
type AIResponse struct {
    Content []ContentBlock `json:"content"`
    Error   string         `json:"error,omitempty"`
}
```

### Provider Interface

```go
// AIHandler is the interface all providers implement
type AIHandler interface {
    Call(input string) (string, error)
}
```

### Factory Configuration (ai_factory.go)

```go
type AIProvider string
const (
    AIProviderStub   AIProvider = "stub"
    AIProviderClaude AIProvider = "claude"
    AIProviderGemini AIProvider = "gemini"
)

type AIConfig struct {
    Provider AIProvider

    // Claude config
    ClaudeAPIKey string  // or ANTHROPIC_API_KEY env
    ClaudeModel  string  // default: claude-haiku-4-5-20251001

    // Gemini config (Vertex AI preferred)
    GeminiProject     string  // GCP project (uses ADC)
    GeminiLocation    string  // default: us-central1
    GeminiAPIKey      string  // fallback: GOOGLE_API_KEY env
    GeminiModel       string  // default: gemini-2.5-flash
    GeminiImagenModel string  // default: gemini-2.5-flash-image
    GeminiTTSModel    string  // default: gemini-2.5-flash-tts
    GeminiTTSVoice    string  // default: Kore

    AssetsDir string  // where to save generated media
}
```

### Auto-Detection Priority

When no explicit provider is set, the factory auto-detects:

1. **Vertex AI** - `GOOGLE_CLOUD_PROJECT` env (uses ADC for auth)
2. **Gemini API** - `GOOGLE_API_KEY` env
3. **Claude** - `ANTHROPIC_API_KEY` env
4. **Stub** - fallback if nothing configured

### Gemini Multimodal Features

**Image Generation:**
```go
// Triggered by keywords: "generate image", "create image", "draw", "imagen:"
// Or context flag: {"generate_image": true}

// Uses gemini-2.5-flash-image with IMAGE response modality
config := &genai.GenerateContentConfig{
    ResponseModalities: []string{"IMAGE", "TEXT"},
}
```

**Image Editing (iterative refinement):**
```go
// Triggered by keywords: "edit:", "modify:", "change:", "refine:", "fix:", "adjust:"
// Or context flag: {"edit_image": true}
// Plus either: explicit image in messages, context["reference_image"], or "last image" reference

// Reference image sources (priority order):
// 1. Explicit image in request messages
// 2. context["reference_image"] path
// 3. Last generated image (tracked automatically)

// Example usage:
{"messages": [
    {"type": "image", "image_ref": "path/to/image.png"},
    {"type": "text", "text": "edit: make the sky more purple"}
]}

// Or reference last generated image:
{"messages": [{"type": "text", "text": "edit the last image: add more stars"}]}

// Helper methods:
handler.LastGeneratedImage()       // Get path to last generated
handler.SetLastGeneratedImage(p)   // Set reference manually
handler.ClearLastGeneratedImage()  // Clear reference
```

**Text-to-Speech:**
```go
// Triggered by keywords: "speak:", "say:", "tts:", "voice:"
// Or context flag: {"tts": true}

// Available voices: Aoede, Charon, Fenrir, Kore, Puck, Zephyr
config := &genai.GenerateContentConfig{
    ResponseModalities: []string{"AUDIO"},
    SpeechConfig: &genai.SpeechConfig{
        VoiceConfig: &genai.VoiceConfig{
            PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
                VoiceName: "Kore",
            },
        },
    },
}
```

**Multimodal Input:**
- Images: Load from file path, auto-detect MIME type
- Audio: WAV, MP3, OGG, FLAC supported
- Generated media saved to `assets/generated/`

### Stub Handler

Pattern-based responses for testing without API calls:

```go
stub := NewStubAIHandler()
stub.RegisterTextResponse("decide", `{"action": "wait"}`)
stub.RegisterTextResponse("emotion", `{"emotion": "contemplative"}`)
```

Built-in patterns:
- `decide`/`decision` → `{"action": "wait", "reason": "..."}`
- `emotion`/`feel` → `{"emotion": "contemplative", "dialogue": "..."}`
- `describe`/`what do you see` → text + image reference

### Chain Handler (Fallback)

```go
// Try Gemini first, fall back to Claude, then stub
handler := NewChainAIHandler(geminiHandler, claudeHandler, stubHandler)
```

## Usage

### From Environment

```go
// Reads AI_PROVIDER, GOOGLE_CLOUD_PROJECT, ANTHROPIC_API_KEY, etc.
handler, err := NewAIHandlerFromEnv(ctx)
```

### Explicit Configuration

```go
handler, err := NewAIHandler(ctx, AIConfig{
    Provider:      AIProviderGemini,
    GeminiProject: "my-project",
    AssetsDir:     "assets/generated",
})
```

### Registration with sim_gen

```go
sim_gen.Init(sim_gen.Handlers{
    AI: handler,
    // ... other handlers
})
```

## Environment Variables

| Variable | Provider | Description |
|----------|----------|-------------|
| `AI_PROVIDER` | All | Force provider: `claude`, `gemini`, `stub` |
| `ANTHROPIC_API_KEY` | Claude | API key |
| `CLAUDE_MODEL` | Claude | Model override |
| `GOOGLE_CLOUD_PROJECT` | Gemini | Vertex AI project (uses ADC) |
| `GOOGLE_CLOUD_LOCATION` | Gemini | Vertex AI region |
| `GOOGLE_API_KEY` | Gemini | API key (fallback) |
| `GEMINI_MODEL` | Gemini | Model override |
| `GEMINI_TTS_VOICE` | Gemini | TTS voice name |
| `AI_ASSETS_DIR` | All | Where to save generated media |

## File Structure

```
engine/handlers/
├── ai.go           # Core types, StubAIHandler (190 LOC)
├── ai_claude.go    # Claude API handler (123 LOC)
├── ai_gemini.go    # Gemini multimodal handler (663 LOC)
├── ai_factory.go   # Provider factory, auto-detect (187 LOC)
└── ai_test.go      # Tests (293 LOC)
```

**Total:** 1,456 LOC

## Testing

### Unit Tests

```bash
go test ./engine/handlers/... -v
```

### Manual Testing

```bash
# With Vertex AI
export GOOGLE_CLOUD_PROJECT=my-project
make run

# With Gemini API key
export GOOGLE_API_KEY=xxx
make run

# With Claude
export ANTHROPIC_API_KEY=xxx
export AI_PROVIDER=claude
make run

# Stub only (no API)
export AI_PROVIDER=stub
make run
```

## Game Integration

### Current Usage

The AI handler is initialized in `cmd/game/main.go`:

```go
aiHandler, _ := handlers.NewAIHandlerFromEnv(ctx)
sim_gen.Init(sim_gen.Handlers{
    AI: aiHandler,
    // ...
})
```

### Planned: AILANG Integration

```ailang
-- In sim/ai_decisions.ail
func civDecide(civ: Civ, context: CivContext) -> CivAction ! {AI} {
    let input = encodeCivContext(civ, context)
    let output = AI.decide(input)
    decodeCivAction(output)
}
```

### Planned: The Archive (Unreliable AI NPC)

The AI handler will power "The Archive" - an in-game AI that:
- Maintains civilization memory with intentional degradation
- Hallucinates history based on context compression
- Provides unreliable narration as a game mechanic

See [ai-the-archive.md](../../input/ai-the-archive.md) for full design.

## Success Criteria

### Implemented
- [x] Multiple provider support (Claude, Gemini, stub)
- [x] Automatic provider detection from environment
- [x] Multimodal content types
- [x] Image generation (Gemini)
- [x] Image editing with reference images (Gemini)
- [x] Last-image tracking for iterative refinement
- [x] Text-to-speech (Gemini)
- [x] Chain handler with fallback
- [x] Pattern-based stub for testing

### Planned
- [ ] AILANG AI effect integration
- [ ] Typed decision contexts (CivContext, CrewContext)
- [ ] Philosophy-driven prompts
- [ ] Caching layer for determinism
- [ ] Rate limiting for API calls

## Dependencies

| Package | Purpose |
|---------|---------|
| `google.golang.org/genai` | Gemini API client |
| `github.com/anthropics/anthropic-sdk-go` | Claude API client |

---

**Document created**: 2025-12-06
**Last updated**: 2025-12-06 (added image editing)
