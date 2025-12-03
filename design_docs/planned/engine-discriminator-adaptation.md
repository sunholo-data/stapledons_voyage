# Engine Discriminator Struct Adaptation

## Status
- **Status:** Planned
- **Priority:** P0 (Blocking AILANG integration)
- **Estimated:** 1-2 days
- **Author:** Claude Code
- **Created:** 2025-12-03

## Overview

Adapt the Go engine layer to work with AILANG's discriminator struct pattern for ADTs (Algebraic Data Types). Currently, the engine expects interface-based type switching, but AILANG codegen produces discriminator structs for better performance.

## Game Vision Alignment

| Pillar | Score | Rationale |
|--------|-------|-----------|
| Time Dilation Consequence | N/A | Infrastructure |
| Civilization Simulation | N/A | Infrastructure |
| Philosophical Depth | N/A | Infrastructure |
| Ship & Crew Life | N/A | Infrastructure |
| Legacy Impact | N/A | Infrastructure |
| Hard Sci-Fi Authenticity | N/A | Infrastructure |

**This is enabling infrastructure** - required to use AILANG codegen instead of hand-written mock code.

## Problem Statement

### Current State (Mock sim_gen)

The engine uses Go interface-based ADTs:

```go
// DrawCmd is an interface - each variant is a separate type
type DrawCmd interface {
    isDrawCmd()
}

type DrawCmdRect struct {
    X, Y, W, H float64
    Color, Z   int
}
func (DrawCmdRect) isDrawCmd() {}

// Engine code uses type switch
switch c := cmd.(type) {
case DrawCmdRect:
    draw(c.X, c.Y, c.W, c.H, c.Color)
case DrawCmdSprite:
    drawSprite(c.ID, c.X, c.Y)
}
```

### Target State (AILANG Codegen)

AILANG produces discriminator structs:

```go
// Single struct with Kind discriminator
type DrawCmdKind int
const (
    DrawCmdKindRect DrawCmdKind = iota
    DrawCmdKindSprite
    DrawCmdKindText
    // ... all variants
)

type DrawCmd struct {
    Kind   DrawCmdKind
    Rect   *DrawCmdRect   // Non-nil when Kind == DrawCmdKindRect
    Sprite *DrawCmdSprite // Non-nil when Kind == DrawCmdKindSprite
    // ... one pointer per variant
}

type DrawCmdRect struct {
    Value0 float64 // x
    Value1 float64 // y
    Value2 float64 // w
    Value3 float64 // h
    Value4 int64   // color
    Value5 int64   // z
}

// Engine code uses Kind switch
switch cmd.Kind {
case DrawCmdKindRect:
    c := cmd.Rect
    draw(c.Value0, c.Value1, c.Value2, c.Value3, c.Value4)
}
```

### Key Differences

| Aspect | Interface-based | Discriminator Struct |
|--------|-----------------|---------------------|
| **Memory** | Separate type per variant | One struct, pointers to variants |
| **Dispatch** | Runtime type switch | Integer compare (faster) |
| **Field names** | Named (`X`, `Y`, `Color`) | Positional (`Value0`, `Value1`) |
| **Nil safety** | Type assertion can fail | Pointer dereference on Kind mismatch |
| **Go idiom** | More "Go-like" | More "functional/ML-like" |
| **Hot loops** | Interface dispatch overhead | Cache-friendly, predictable |

## Technical Approach

### Phase 1: Update Type References

Files to update:
- `engine/render/draw.go` - Main rendering logic
- `engine/render/isometric.go` - Isometric helpers (if exists)
- `engine/camera/transform.go` - Camera type usage
- `engine/camera/viewport.go` - Camera type usage

### Phase 2: Convert Type Switches to Kind Switches

**Before:**
```go
switch c := cmd.(type) {
case sim_gen.DrawCmdRect:
    col := biomeColors[c.Color%len(biomeColors)]
    ebitenutil.DrawRect(screen, c.X, c.Y, c.W, c.H, col)
}
```

**After:**
```go
switch cmd.Kind {
case sim_gen.DrawCmdKindRect:
    c := cmd.Rect
    col := biomeColors[int(c.Value4)%len(biomeColors)]
    ebitenutil.DrawRect(screen, c.Value0, c.Value1, c.Value2, c.Value3, col)
}
```

### Phase 3: Handle FrameOutput.Draw as interface{}

The generated `FrameOutput.Draw` is `interface{}` (list type). Need to cast:

```go
// Cast draw commands from interface{}
drawCmds, ok := out.Draw.([]*sim_gen.DrawCmd)
if !ok {
    // Handle error or empty list
    return
}
for _, cmd := range drawCmds {
    // process cmd
}
```

### Phase 4: Create Field Name Constants (Optional DX improvement)

To improve readability, create constants mapping positional fields:

```go
// DrawCmdRect field indices
const (
    RectX     = iota // Value0
    RectY            // Value1
    RectW            // Value2
    RectH            // Value3
    RectColor        // Value4
    RectZ            // Value5
)

// Usage
c := cmd.Rect
draw(c.Value0, c.Value1, c.Value2, c.Value3) // or use accessor functions
```

## Files to Modify

| File | Changes |
|------|---------|
| `engine/render/draw.go` | Main render loop, all type switches |
| `engine/camera/transform.go` | Camera struct field access |
| `engine/camera/viewport.go` | Camera struct field access |
| `sim_gen/compat.go` | (New) Optional compatibility layer |

## Testing Strategy

1. **Build test:** `go build ./...` passes
2. **Run test:** `make run-mock` works (with adapted engine)
3. **Visual test:** Screenshots match golden files
4. **Benchmark:** No performance regression

## Rollback Strategy

If issues arise:
1. `git checkout sim_gen/` to restore mock
2. Engine changes are backward-compatible with mock

## Success Criteria

- [ ] `make sim` compiles AILANG to Go
- [ ] `make game` builds with generated sim_gen
- [ ] Game runs and renders correctly
- [ ] All visual tests pass
- [ ] No performance regression

## AILANG Feedback

### Issues Found During Testing

1. **Positional field names (`Value0`)** - Makes code less readable
   - Workaround: Document field order in comments or use constants
   - Feature request: Named fields in codegen for ADT payloads

2. **`FrameOutput.Draw` as `interface{}`** - Should be typed slice
   - Feature request: Generate `[]*DrawCmd` instead of `interface{}` for list fields

## References

- [Consumer Contract v0.5](../../ailang_resources/consumer-contract-v0.5.md) - ADT specification
- [CLAUDE.md](../../CLAUDE.md) - Build commands, mock vs codegen
- [AILANG Prompt](run `ailang prompt`) - Language syntax reference

## Appendix: Full Type Mapping

### DrawCmd Variants

| Variant | AILANG | Go Fields |
|---------|--------|-----------|
| Sprite | `Sprite(id, x, y, z)` | `Value0:int64, Value1:float64, Value2:float64, Value3:int64` |
| Rect | `Rect(x, y, w, h, color, z)` | `Value0-3:float64, Value4-5:int64` |
| Text | `Text(text, x, y, fontSize, color, z)` | `Value0:string, Value1-2:float64, Value3-5:int64` |
| IsoTile | `IsoTile(tile, height, spriteId, layer, color)` | `Value0:Coord, Value1-4:int64` |
| IsoEntity | `IsoEntity(id, tile, offsetX, offsetY, height, spriteId, layer)` | Mixed |
| Ui | `Ui(id, kind, x, y, w, h, text, spriteId, z, color, value)` | Mixed |
| Line | `Line(x1, y1, x2, y2, color, width, z)` | `Value0-5:float64/int, Value6:int64` |
| TextWrapped | `TextWrapped(text, x, y, maxWidth, fontSize, color, z)` | Mixed |
| Circle | `Circle(x, y, radius, color, filled, z)` | `Value0-2:float64, Value3:int64, Value4:bool, Value5:int64` |
| RectScreen | `RectScreen(x, y, w, h, color, z)` | Same as Rect |
| GalaxyBg | `GalaxyBg(opacity, z, skyViewMode, viewLon, viewLat, fov)` | Mixed |
| Star | `Star(x, y, spriteId, scale, alpha, z)` | `Value0-1:float64, Value2:int64, Value3-4:float64, Value5:int64` |
