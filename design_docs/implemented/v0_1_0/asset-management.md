# Asset Management System

**Version:** 0.2.0
**Status:** Implemented
**Priority:** P0 (High)
**Complexity:** Medium
**Package:** `engine/assets`

## Related Documents

- [Engine Layer Design](../../implemented/v0_1_0/engine-layer.md) - Current stub implementation
- [Architecture Overview](../../implemented/v0_1_0/architecture.md) - System context

## Problem Statement

The v0.1.0 engine uses placeholder rectangles for sprites and has no font or sound loading. To render actual game graphics, we need a proper asset management system.

**Current State:**
- AssetManager is a stub returning nil
- Sprites render as white rectangles
- No font loading (uses Ebiten debug font)
- No sound support

**What's Needed:**
- Load sprites from `assets/sprites/` directory
- Load fonts from `assets/fonts/`
- Load sounds from `assets/sounds/`
- Cache loaded assets by ID
- Handle missing assets gracefully

## Design

### Directory Structure

```
assets/
├── sprites/
│   ├── manifest.json       # ID → filename mapping
│   ├── player.png
│   ├── npc_001.png
│   └── tiles/
│       ├── water.png
│       ├── grass.png
│       └── mountain.png
├── fonts/
│   ├── manifest.json
│   └── pixel.ttf
└── sounds/
    ├── manifest.json
    ├── step.wav
    └── ambient.ogg
```

### Manifest Format

```json
{
  "sprites": {
    "1": {"file": "player.png", "width": 32, "height": 32},
    "2": {"file": "npc_001.png", "width": 32, "height": 32},
    "100": {"file": "tiles/water.png", "width": 16, "height": 16}
  }
}
```

### Go Implementation

```go
package assets

type Manager struct {
    sprites   map[int]*ebiten.Image
    fonts     map[string]*text.GoTextFace
    sounds    map[int]*audio.Player
    basePath  string
}

func NewManager(basePath string) (*Manager, error)
func (m *Manager) LoadManifests() error
func (m *Manager) GetSprite(id int) *ebiten.Image
func (m *Manager) GetFont(name string) *text.GoTextFace
func (m *Manager) GetSound(id int) *audio.Player
```

### Loading Strategy

1. **Eager loading:** Load all assets at startup
2. **Manifest-driven:** IDs defined in JSON, not hardcoded
3. **Fallback sprites:** Return placeholder for missing assets
4. **Error reporting:** Log missing assets, don't crash

### Integration with AILANG

AILANG's `DrawCmd.Sprite(id, x, y, z)` uses integer IDs. The manifest maps these IDs to actual files:

```
AILANG: Sprite(1, 100.0, 200.0, 0)
  ↓
Engine: manager.GetSprite(1) → player.png image
  ↓
Ebiten: DrawImage at (100, 200)
```

## Implementation Plan

### Files to Create

| File | Purpose |
|------|---------|
| `engine/assets/manager.go` | Core Manager struct and loading |
| `engine/assets/sprites.go` | Sprite loading and caching |
| `engine/assets/fonts.go` | Font loading (TTF support) |
| `engine/assets/sounds.go` | Sound loading (WAV/OGG) |
| `assets/sprites/manifest.json` | Sprite ID mappings |

### Changes to Existing Files

| File | Change |
|------|--------|
| `engine/render/draw.go` | Use Manager instead of placeholders |
| `cmd/game/main.go` | Initialize Manager at startup |

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/hajimehoshi/ebiten/v2` | Image loading |
| `github.com/hajimehoshi/ebiten/v2/text/v2` | Font rendering |
| `github.com/hajimehoshi/ebiten/v2/audio` | Sound playback |
| `github.com/hajimehoshi/ebiten/v2/audio/wav` | WAV decoding |
| `github.com/hajimehoshi/ebiten/v2/audio/vorbis` | OGG decoding |

## Testing Strategy

### Unit Tests

```go
func TestLoadSprite(t *testing.T)
func TestMissingSpriteFallback(t *testing.T)
func TestManifestParsing(t *testing.T)
```

### Integration Tests

```bash
make run  # Verify sprites render correctly
```

### Edge Cases

- [ ] Missing manifest file → use empty manifest
- [ ] Invalid sprite ID → return placeholder
- [ ] Corrupt image file → log error, return placeholder
- [ ] Missing font → fall back to debug font

## Success Criteria

### Asset Loading
- [ ] Sprites load from PNG files
- [ ] Fonts load from TTF files
- [ ] Sounds load from WAV/OGG files
- [ ] Manifest files parsed correctly

### Integration
- [ ] DrawCmd.Sprite renders actual images
- [ ] DrawCmd.Text uses loaded fonts
- [ ] No crashes on missing assets
- [ ] Asset IDs match AILANG definitions

### Performance
- [ ] Assets cached after first load
- [ ] Startup time < 2 seconds with test assets
- [ ] No per-frame file I/O
