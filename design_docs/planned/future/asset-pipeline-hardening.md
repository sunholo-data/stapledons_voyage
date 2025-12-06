# Asset Pipeline Hardening

**Version:** 0.5.0
**Status:** Planned
**Priority:** P1 (Engine Foundation)
**Complexity:** Medium
**Dependencies:** None (pure engine work)
**AILANG Impact:** None - AILANG just sends IDs, engine handles loading

## Problem Statement

**Current State:**
- Sprite loading works but sprites never render (SpriteID 0 triggers colored fallback)
- Existing tile sprites are 16x16 squares, not 64x32 isometric diamonds
- No audio system - `FrameOutput.Sounds` is ignored
- No font loading - uses `ebitenutil.DebugPrint` everywhere
- No animation support for sprite sequences

**Impact:**
- Game looks like a prototype with colored shapes
- No audio feedback for player actions
- Text rendering is limited to debug font

**What AILANG Will Send:**
```
DrawCmdIsoTile{SpriteID: 1, Tile: {3,5}, ...}  → Engine needs sprite 1 to exist
DrawCmdIsoEntity{SpriteID: 100, ...}           → Engine needs sprite 100 to exist
FrameOutput.Sounds: [1, 2]                     → Engine needs to play sounds 1 and 2
DrawCmdText{Text: "Hello"}                     → Engine needs proper font rendering
```

## Goals

1. **Sprites render correctly** - When AILANG says "sprite 5", sprite 5 renders
2. **Audio plays** - When AILANG says "play sound 2", sound 2 plays
3. **Fonts render** - Text uses proper TTF fonts, not debug text
4. **Animation ready** - Engine can advance sprite frames over time

## Current Implementation Gap

### Sprite Loading (Partial)
```go
// draw.go line 182 - SpriteID > 0 means sprite 0 never uses asset!
if r.assets != nil && c.SpriteID > 0 {
    sprite := r.assets.GetSprite(c.SpriteID)
```

### Manifest IDs vs Usage
```json
// Current manifest.json - IDs 0-3 for tiles
"0": {"file": "tile_water.png", ...}
"1": {"file": "tile_forest.png", ...}
```

```go
// But sim_gen/funcs.go uses Color field, not SpriteID
DrawCmdIsoTile{SpriteID: 0, Color: tile.Biome, ...}  // SpriteID=0 → fallback!
```

**Fix:** Either use SpriteID properly OR document that colored fallback is intentional.

---

## Design

### Phase 1: Fix Sprite ID Usage

**Option A: Use SpriteID for tiles (recommended)**
- Change sim_gen to emit `SpriteID: tile.Biome + 1` (1-4 instead of 0-3)
- Update manifest to use IDs 1-4 for tiles
- Keep SpriteID 0 as "use colored fallback" sentinel

**Option B: Keep colored fallback for tiles**
- Document that SpriteID 0 means "use Color field for colored diamond"
- Use SpriteID for entities only (100+)
- Pros: Colored fallback is deterministic for testing
- Cons: Never renders actual tile sprites

**Recommendation:** Option B for now - colored tiles are fine for testing, sprites for entities.

### Phase 2: Isometric Tile Sprites

Current sprites are 16x16 squares. For proper isometric rendering, we need:

**Tile Sprite Spec:**
- Size: 64x32 pixels (matching TileWidth/TileHeight constants)
- Shape: Diamond with transparent corners
- Format: PNG with alpha channel

```
    ████████
  ██████████████
████████████████████
  ██████████████
    ████████
```

**Entity Sprite Spec:**
- Size: 32x48 pixels (taller than tile for perspective)
- Anchor: Bottom-center (feet at tile center)
- Format: PNG with alpha channel

**Files to Create:**
```
assets/sprites/
├── iso_tiles/
│   ├── water.png      (64x32)
│   ├── forest.png     (64x32)
│   ├── desert.png     (64x32)
│   └── mountain.png   (64x32)
├── iso_entities/
│   ├── npc_red.png    (32x48)
│   ├── npc_green.png  (32x48)
│   ├── npc_blue.png   (32x48)
│   └── npc_yellow.png (32x48)
└── manifest.json
```

**Updated Manifest:**
```json
{
  "sprites": {
    "1": {"file": "iso_tiles/water.png", "width": 64, "height": 32, "type": "tile"},
    "2": {"file": "iso_tiles/forest.png", "width": 64, "height": 32, "type": "tile"},
    "3": {"file": "iso_tiles/desert.png", "width": 64, "height": 32, "type": "tile"},
    "4": {"file": "iso_tiles/mountain.png", "width": 64, "height": 32, "type": "tile"},
    "100": {"file": "iso_entities/npc_red.png", "width": 32, "height": 48, "type": "entity"},
    "101": {"file": "iso_entities/npc_green.png", "width": 32, "height": 48, "type": "entity"},
    "102": {"file": "iso_entities/npc_blue.png", "width": 32, "height": 48, "type": "entity"},
    "103": {"file": "iso_entities/npc_yellow.png", "width": 32, "height": 48, "type": "entity"}
  }
}
```

### Phase 3: Audio System

**New Files:**
- `engine/assets/audio.go` - Audio loading and playback

**Audio Manager:**
```go
type AudioManager struct {
    context    *audio.Context
    sounds     map[int]*audio.Player
    bgmPlayer  *audio.Player
    volume     float64
}

func (am *AudioManager) LoadManifest(soundPath string) error
func (am *AudioManager) PlaySound(id int)
func (am *AudioManager) PlayBGM(id int)
func (am *AudioManager) SetVolume(vol float64)
```

**Audio Manifest:**
```json
{
  "sounds": {
    "1": {"file": "click.wav", "volume": 1.0},
    "2": {"file": "build.wav", "volume": 0.8},
    "3": {"file": "error.wav", "volume": 0.7}
  },
  "bgm": {
    "1": {"file": "ambient.ogg", "loop": true, "volume": 0.5}
  }
}
```

**Integration:**
```go
// In game loop, after Step():
for _, soundID := range out.Sounds {
    audioMgr.PlaySound(soundID)
}
```

### Phase 4: Font System

**New Files:**
- `engine/assets/fonts.go` - Font loading

**Font Manager:**
```go
type FontManager struct {
    fonts map[string]*text.GoTextFace
}

func (fm *FontManager) LoadFont(name, path string, size float64) error
func (fm *FontManager) Get(name string) *text.GoTextFace
func (fm *FontManager) GetDefault() *text.GoTextFace
```

**Integration:**
Update `drawUiElement` to use loaded fonts instead of `DebugPrintAt`.

### Phase 5: Animation Support (Future)

**Extended Sprite Entry:**
```json
{
  "sprites": {
    "100": {
      "file": "npc_walk.png",
      "width": 32,
      "height": 48,
      "type": "entity",
      "frames": 4,
      "frameWidth": 32,
      "fps": 8
    }
  }
}
```

**Animation Runtime:**
Engine tracks elapsed time, calculates current frame, renders correct sub-rectangle.

---

## Implementation Plan

### Sprint 1: Sprite Pipeline Fix
- [ ] Update gensprites to create 64x32 isometric diamond PNGs
- [ ] Update manifest with proper IDs (1-4 for tiles, 100+ for entities)
- [ ] Fix draw.go to use SpriteID correctly (or document fallback behavior)
- [ ] Test that sprites render when SpriteID > 0
- [ ] Update golden files after visual change

### Sprint 2: Audio Foundation
- [ ] Create `engine/assets/audio.go`
- [ ] Add AudioManager to main Manager struct
- [ ] Load audio manifest at startup
- [ ] Implement PlaySound(id)
- [ ] Hook into game loop to play FrameOutput.Sounds
- [ ] Add test sound files

### Sprint 3: Font Rendering
- [ ] Create `engine/assets/fonts.go`
- [ ] Find/create pixel art TTF font
- [ ] Update UI rendering to use loaded fonts
- [ ] Add font size configuration

### Sprint 4: Polish & Animation (Optional)
- [ ] Add animation metadata to manifest
- [ ] Implement frame advancement in renderer
- [ ] Create animated NPC sprites

---

## Testing Strategy

### Sprite Tests
```bash
# Run game and verify sprites appear
make run-mock

# Take screenshot with sprites
make screenshot

# Compare golden files (will diff after sprite change)
make test-golden
```

### Audio Tests
```bash
# Manual: Run game, trigger action, hear sound
make run-mock
# Press B to build → should hear build.wav
```

### Integration Test
```bash
# Full test suite including visuals
make test-all
```

---

## Files to Modify

| File | Change |
|------|--------|
| `engine/assets/manager.go` | Add AudioManager, FontManager |
| `engine/assets/audio.go` | **New** - Audio loading and playback |
| `engine/assets/fonts.go` | **New** - Font loading |
| `engine/render/draw.go` | Use fonts in UI rendering |
| `cmd/game/main.go` | Play sounds from FrameOutput |
| `cmd/gensprites/main.go` | Generate 64x32 isometric tiles |
| `assets/sprites/manifest.json` | Updated IDs and paths |
| `assets/sounds/manifest.json` | **New** - Sound definitions |
| `assets/fonts/manifest.json` | **New** - Font definitions |

---

## Success Criteria

### Sprites
- [ ] Running `make run-mock` shows sprites (not just colored shapes) when SpriteID > 0
- [ ] Isometric tile sprites are 64x32 diamonds
- [ ] Entity sprites render with correct anchor point

### Audio
- [ ] `FrameOutput.Sounds: [1]` causes sound 1 to play
- [ ] No crashes if sound ID not found (log warning, skip)
- [ ] Volume control works

### Fonts
- [ ] UI text renders with loaded TTF font
- [ ] Fallback to debug font if TTF fails to load

### Integration
- [ ] All existing tests pass
- [ ] Golden files updated for new visuals
- [ ] No performance regression (< 1ms asset lookup)

---

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| ebiten/v2/audio | v2.8+ | Audio context and players |
| ebiten/v2/audio/wav | v2.8+ | WAV file decoding |
| ebiten/v2/audio/vorbis | v2.8+ | OGG file decoding |
| ebiten/v2/text/v2 | v2.8+ | TTF font rendering |

---

## Notes

- **Colored fallback is useful** - Keep SpriteID 0 as "use Color field" for testing
- **Audio is optional** - Game should work without sounds (log warning)
- **Fonts are optional** - Fall back to DebugPrint if TTF fails
- **Animation is future work** - Basic sprite rendering first

---

**Created:** 2025-12-01
**Last Updated:** 2025-12-01
