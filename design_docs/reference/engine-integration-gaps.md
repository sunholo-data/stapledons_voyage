# Engine Integration Gaps

**Status**: Active (Updated 2025-12-04)
**Priority**: P1 - Required for playable game
**Purpose**: Define what engine (Go/Ebiten) features need implementation or completion

## Overview

This document identifies gaps in the Go engine that prevent a playable game, independent of AILANG features. These are pure Go tasks.

## Current Engine Status

| Component | Status | Location | Notes |
|-----------|--------|----------|-------|
| Game loop | **Working** | `cmd/game/main.go` | Ebiten Update/Draw |
| Render bridge | **Working** | `engine/render/` | All DrawCmd types rendered |
| Camera transform | **Working** | `engine/camera/` | WorldToScreen, viewport culling |
| Input capture | **Working** | `engine/input/` | Mouse, keyboard, click detection |
| Sprite loading | **Working** | `engine/assets/sprites.go` | Atlas, animations supported |
| Audio system | **Working** | `engine/assets/audio.go` | OGG/WAV, manifest, volume |
| Font loading | **Working** | `engine/assets/fonts.go` | TTF, size scaling |
| UI rendering | **Working** | `engine/render/draw.go` | Panel, Button, Label, Portrait, Slider, ProgressBar |
| Display config | **Working** | `engine/display/` | Resolution, F11 fullscreen |
| Screenshot | **Working** | `engine/screenshot/` | Test scenarios |
| Scenarios | **Working** | `engine/scenario/` | Eval framework |
| Rand effect | **Working** | Auto-init by AILANG | NPCs use random movement |
| Debug effect | **Working** | Auto-init by AILANG | tick logging works |

---

## Actual Remaining Gaps

Based on code review (2025-12-04), most engine features are complete.

### Completed (2025-12-04)

| Gap | Status | Implementation |
|-----|--------|----------------|
| Clock handler | ✅ Done | `engine/handlers/clock.go` - connected in main.go |
| AI handler stub | ✅ Done | `engine/handlers/ai.go` - multimodal support |
| Handler initialization | ✅ Done | `cmd/game/main.go` calls `sim_gen.Init()` |
| Runtime typed slice fix | ✅ Done | Fixed `ListLen`, `ListHead`, `ListTail`, `Length`, `Get`, `GetOpt` |

### Remaining

| Gap | Priority | Effort | Blocks |
|-----|----------|--------|--------|
| Save system | P2 | 2 days | Single-file save/load (no slots - Pillar 1) |
| AI handler real impl | P3 | 2 days | NPC dialogue with actual LLM |

---

## Gap 1: Clock Handler Integration (ACTUAL GAP)

**Priority:** P2
**Current state:** ClockHandler interface exists, not connected to game loop
**Blocks:** Frame-rate-independent movement, smooth animations

### Implementation

**File:** `engine/camera/camera.go`

```go
type Camera struct {
    X, Y     float64 // Current position (world coords)
    TargetX  float64 // Target position for smooth follow
    TargetY  float64
    Zoom     float64 // Current zoom level
    MinZoom  float64 // Zoom limits
    MaxZoom  float64
}

// Update smoothly moves camera toward target
func (c *Camera) Update(dt float64) {
    lerp := 1.0 - math.Pow(0.01, dt) // Smooth factor
    c.X += (c.TargetX - c.X) * lerp
    c.Y += (c.TargetY - c.Y) * lerp
}

// HandleZoom processes mouse wheel input
func (c *Camera) HandleZoom(wheelY float64) {
    c.Zoom *= 1.0 + wheelY*0.1
    c.Zoom = clamp(c.Zoom, c.MinZoom, c.MaxZoom)
}
```

### Test Plan

- [ ] Camera follows target smoothly (no jitter)
- [ ] Zoom in/out with mouse wheel
- [ ] Camera stops at world bounds
- [ ] Pan works with middle mouse drag

---

## Gap 2: Sprite Rendering

**Priority:** P1
**Current state:** Rect rendering works, sprite loading partial
**Blocks:** Visual game content

### Missing Features

| Feature | Status | Effort |
|---------|--------|--------|
| Sprite atlas loading | Partial | 1 day |
| Sprite DrawCmd rendering | Missing | 1 day |
| Animation frames | Missing | 2 days |
| Sprite layering (z-sort) | Partial | 0.5 day |

### Implementation

**File:** `engine/render/sprites.go`

```go
type SpriteAtlas struct {
    Image    *ebiten.Image
    Regions  map[int]image.Rectangle // spriteId → region
}

func (r *Renderer) DrawSprite(cmd sim_gen.DrawCmdSprite, cam camera.Transform) {
    region, ok := r.atlas.Regions[int(cmd.Id)]
    if !ok {
        return // Unknown sprite
    }

    sx, sy := cam.WorldToScreen(cmd.X, cmd.Y)

    // Cull if off-screen
    if sx < -64 || sx > float64(r.screenW)+64 {
        return
    }

    opts := &ebiten.DrawImageOptions{}
    opts.GeoM.Translate(sx, sy)
    r.screen.DrawImage(r.atlas.Image.SubImage(region).(*ebiten.Image), opts)
}
```

### Test Plan

- [ ] Load sprite atlas from PNG
- [ ] Render Sprite DrawCmd at correct position
- [ ] Z-ordering respects layer
- [ ] Animation cycles through frames

---

## Gap 3: Sound System

**Priority:** P2
**Current state:** Not implemented
**Blocks:** Audio feedback, atmosphere

### Missing Features

| Feature | Status | Effort |
|---------|--------|--------|
| Sound loading (OGG/WAV) | Missing | 1 day |
| SoundCmd playback | Missing | 1 day |
| Volume control | Missing | 0.5 day |
| Music streaming | Missing | 1 day |

### Implementation

**File:** `engine/audio/sound.go`

```go
type SoundManager struct {
    context *oto.Context
    sounds  map[int]*Sound // soundId → Sound
    music   *MusicPlayer
}

func (s *SoundManager) PlaySound(id int, volume float64) {
    if sound, ok := s.sounds[id]; ok {
        sound.Play(volume)
    }
}

// Called from game loop with FrameOutput.Sounds
func (s *SoundManager) ProcessSounds(soundIds []int64) {
    for _, id := range soundIds {
        s.PlaySound(int(id), 1.0)
    }
}
```

### Test Plan

- [ ] Load OGG file
- [ ] Play sound on SoundCmd
- [ ] Volume adjustment works
- [ ] Multiple simultaneous sounds
- [ ] Music loops correctly

---

## Gap 4: UI Rendering Layer

**Priority:** P2
**Current state:** UiKind types exist but no rendering
**Blocks:** Menus, HUD, dialogue

### Missing Features

| Feature | Status | Effort |
|---------|--------|--------|
| UI DrawCmd rendering | Missing | 2 days |
| Text rendering (wrapped) | Partial | 1 day |
| Button hover/click | Missing | 1 day |
| Panel backgrounds | Missing | 0.5 day |
| 9-slice scaling | Missing | 1 day |

### Implementation

**File:** `engine/render/ui.go`

```go
func (r *Renderer) DrawUI(cmd sim_gen.DrawCmdUi) {
    // UI is in screen space (not affected by camera)
    switch cmd.Kind.Kind {
    case sim_gen.UiKindKindUiPanel:
        r.drawPanel(cmd.X, cmd.Y, cmd.W, cmd.H, cmd.Color)
    case sim_gen.UiKindKindUiButton:
        r.drawButton(cmd.X, cmd.Y, cmd.W, cmd.H, cmd.Text, cmd.Color)
    case sim_gen.UiKindKindUiLabel:
        r.drawText(cmd.Text, cmd.X, cmd.Y, cmd.Color)
    case sim_gen.UiKindKindUiProgressBar:
        r.drawProgressBar(cmd.X, cmd.Y, cmd.W, cmd.H, cmd.Value, cmd.Color)
    }
}
```

### Test Plan

- [ ] Panel renders at screen position
- [ ] Button shows text, changes on hover
- [ ] Label renders text
- [ ] Progress bar fills correctly
- [ ] UI not affected by camera zoom/pan

---

## Gap 5: Save/Load System

**Priority:** P2
**Current state:** Not implemented
**Blocks:** Persistent game state

### Missing Features

| Feature | Status | Effort |
|---------|--------|--------|
| World serialization | Missing | 2 days |
| File save/load | Missing | 1 day |
| Auto-save | Missing | 0.5 day |
| Save file browser | Missing | 1 day |

### Design Decision

**Option A: Go serialization** - Marshal World struct to JSON
- Pro: Works regardless of AILANG
- Con: Must keep Go types in sync with AILANG

**Option B: AILANG FS effect** - AILANG handles serialization
- Pro: Game logic controls save format
- Con: Requires FS effect testing

**Recommended:** Option A for v1.0, migrate to Option B later

### Implementation

**File:** `engine/save/save.go`

```go
type SaveFile struct {
    Version   string          `json:"version"`
    Timestamp int64           `json:"timestamp"`
    World     json.RawMessage `json:"world"` // sim_gen.World serialized
}

func SaveGame(world interface{}, path string) error {
    worldJSON, err := json.Marshal(world)
    if err != nil {
        return err
    }
    save := SaveFile{
        Version:   "0.1.0",
        Timestamp: time.Now().Unix(),
        World:     worldJSON,
    }
    return os.WriteFile(path, save, 0644)
}
```

### Test Plan

- [ ] World serializes to JSON
- [ ] Deserialization produces equivalent World
- [ ] Save file written to disk
- [ ] Load restores game state
- [ ] Corrupt save handled gracefully

---

## Gap 6: Effect Handler Integration

**Priority:** P1
**Current state:** Handlers declared but not all connected
**Blocks:** Full AILANG feature usage

### Handler Status

| Handler | Interface | Implementation | Connected |
|---------|-----------|----------------|-----------|
| Rand | Defined | Working | Yes |
| Debug | Defined | Working | Yes |
| Clock | Defined | Missing | No |
| FS | Defined | Missing | No |
| Net | Defined | Missing | No |
| AI | Defined | Missing | No |

### Implementation Needed

**Clock handler** (`engine/handlers/clock.go`):
```go
type EbitenClockHandler struct {
    deltaTime  float64
    totalTime  float64
    frameCount int64
}

func (h *EbitenClockHandler) DeltaTime() float64  { return h.deltaTime }
func (h *EbitenClockHandler) TotalTime() float64  { return h.totalTime }
func (h *EbitenClockHandler) FrameCount() int64   { return h.frameCount }

// Called each frame from game loop
func (h *EbitenClockHandler) Update(dt float64) {
    h.deltaTime = dt
    h.totalTime += dt
    h.frameCount++
}
```

**AI handler stub** (`engine/handlers/ai.go`):
```go
type StubAIHandler struct{}

func (h StubAIHandler) Call(input string) (string, error) {
    // Default: echo with prefix for debugging
    return `{"action": "none", "reason": "AI stub"}`, nil
}

// Real implementation would call Anthropic API
type AnthropicAIHandler struct {
    client *anthropic.Client
}

func (h *AnthropicAIHandler) Call(input string) (string, error) {
    // Implementation in ai-effect-npcs.md
}
```

### Test Plan

- [ ] Clock handler updates each frame
- [ ] AILANG receives correct delta time
- [ ] AI stub returns valid JSON
- [ ] Handler registration in main.go works

---

## Implementation Priority

### Phase 1: Core Gameplay (P1)
1. Camera completion (1.5 days)
2. Sprite rendering (2 days)
3. Clock handler (0.5 day)

### Phase 2: Polish (P2)
4. Sound system (3 days)
5. UI rendering (4 days)
6. Save/load (3.5 days)

### Phase 3: Advanced (P3)
7. AI handler with real LLM (when needed)
8. FS/Net handlers (for leaderboards)

---

## Testing Strategy

### Unit Tests
Each engine component should have `*_test.go`:
```bash
go test ./engine/camera/...
go test ./engine/render/...
go test ./engine/audio/...
```

### Integration Tests
Use scenario framework:
```bash
make eval  # Runs all scenarios
```

### Manual Testing Checklist
- [ ] Camera follows player smoothly
- [ ] Sprites render at correct positions
- [ ] Sounds play on events
- [ ] UI panels respond to clicks
- [ ] Save/load preserves game state

---

## Related Documents

- [camera-viewport.md](v0_3_0/camera-viewport.md) - Camera design
- [audio-system.md](v0_2_0/audio-system.md) - Audio design
- [ailang-testing-matrix.md](ailang-testing-matrix.md) - AILANG feature testing
- [save-load-system.md](v0_5_0/save-load-system.md) - Save system design

---

**Document created**: 2025-12-04
**Last updated**: 2025-12-04
