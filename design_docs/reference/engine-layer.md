# Engine Layer Design

**Version:** 0.1.0
**Status:** Planned
**Priority:** P0 (High)
**Complexity:** Medium
**Package:** `engine/render`

## Related Documents

- [Architecture Overview](architecture.md) - Three-layer design context
- [Evaluation System](eval-system.md) - Performance testing

## Overview

The engine layer bridges Ebiten (2D game library) with the AILANG-generated simulation. It handles:
- Input capture (keyboard, mouse → FrameInput)
- Rendering (FrameOutput → screen pixels)
- Asset management (sprite loading, caching)

## Components

### Input Capture (`input.go`)

```go
func CaptureInput() sim_gen.FrameInput
```

Converts Ebiten input state into the AILANG `FrameInput` type.

**Mouse State:**
- Position: `(x, y)` from `ebiten.CursorPosition()`
- Buttons: `[0, 1, 2]` for left, right, middle

**Keyboard Events:**
- Uses `inpututil.AppendPressedKeys` for held keys
- Uses `inpututil.AppendJustReleasedKeys` for release events
- Event kind: `"down"` or `"up"`

**Design Notes:**
- Captures full state each frame (not deltas)
- AILANG simulation decides what input means

### Renderer (`draw.go`)

```go
func RenderFrame(screen *ebiten.Image, out sim_gen.FrameOutput)
```

Processes `DrawCmd` list from simulation output.

**Draw Command Types:**

| Type | Parameters | v0.1.0 Implementation |
|------|------------|----------------------|
| `Rect` | x, y, w, h, color, z | Solid color rectangle |
| `Sprite` | x, y, sprite_id, z | Placeholder (white rect) |
| `Text` | text, x, y, z | Debug font rendering |

**Z-Ordering:**
- Commands sorted by z-index before rendering
- Lower z = rendered first (back)
- Higher z = rendered later (front)

**Biome Colors:**
```go
biomeColors = []color.RGBA{
    {0, 100, 200, 255},   // 0: Water (blue)
    {34, 139, 34, 255},   // 1: Forest (green)
    {210, 180, 140, 255}, // 2: Desert (tan)
    {139, 90, 43, 255},   // 3: Mountain (brown)
}
```

### Asset Manager (`assets.go`)

```go
type AssetManager struct {
    sprites map[int]*ebiten.Image
}
```

**Current Status:** Stub implementation (v0.1.0)

**Planned Features:**
- Sprite loading from `assets/` directory
- Sprite caching by ID
- Font loading for text rendering

## Game Loop (`cmd/game/main.go`)

```go
type Game struct {
    world sim_gen.World
    out   sim_gen.FrameOutput
}

func (g *Game) Update() error {
    input := render.CaptureInput()
    w2, out, err := sim_gen.Step(g.world, input)
    g.world = w2
    g.out = out
    return err
}

func (g *Game) Draw(screen *ebiten.Image) {
    render.RenderFrame(screen, g.out)
}

func (g *Game) Layout(w, h int) (int, int) {
    return 640, 480
}
```

**Fixed Resolution:** 640x480 (v0.1.0)

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/hajimehoshi/ebiten/v2` | v2.x | 2D game library |
| `stapledons_voyage/sim_gen` | generated | AILANG simulation |

## Future Enhancements

### v0.2.0 (Planned)
- [ ] Sprite loading and rendering
- [ ] Configurable resolution
- [ ] Sound playback from `FrameOutput.Sounds`

### v0.3.0 (Planned)
- [ ] Tilemap rendering optimization
- [ ] Camera/viewport support
- [ ] Fullscreen toggle

## File Listing

```
engine/render/
├── input.go    # CaptureInput()
├── draw.go     # RenderFrame(), biome colors
└── assets.go   # AssetManager (stub)
```

## AILANG Integration Notes

**Type Mapping:**

| AILANG Type | Go Type | Notes |
|-------------|---------|-------|
| `FrameInput` | `sim_gen.FrameInput` | Discriminator-based struct |
| `FrameOutput` | `sim_gen.FrameOutput` | Contains `DrawCmd` list |
| `DrawCmd` | `sim_gen.DrawCmd` | Tagged union with `.Kind` |

**Key Patterns:**
- Always switch on `DrawCmd.Kind` to dispatch rendering
- Check for nil pointers before accessing variant fields
- World state is opaque to engine - only read via FrameOutput

## Success Criteria

### Input System
- [ ] Mouse position captured each frame
- [ ] Mouse button state tracked (left, right, middle)
- [ ] Keyboard events captured (pressed keys)
- [ ] FrameInput struct populated correctly

### Rendering
- [ ] DrawCmd list processed from FrameOutput
- [ ] Rect commands render solid colors
- [ ] Z-ordering sorts back-to-front
- [ ] Biome colors display correctly

### Integration
- [ ] sim_gen.Step() called with captured input
- [ ] World state updated between frames
- [ ] No game logic in engine layer
- [ ] Asset manager stub compiles (ready for v0.2.0)
