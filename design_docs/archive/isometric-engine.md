# Isometric Engine Renderer

**Status**: Implemented
**Target**: v0.5.0 (Engine-side, independent of AILANG codegen)
**Priority**: P1 - Required for ship interior view
**Estimated**: 3-4 implementation sessions
**Dependencies**: None (engine-only, extends existing render system)

## Game Vision Alignment

**Feature type:** Infrastructure (Engine rendering)

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Rendering layer, no gameplay impact |
| Civilization Simulation | N/A | 0 | Rendering layer, no simulation impact |
| Philosophical Depth | N/A | 0 | No direct narrative impact |
| Ship & Crew Life | + | +1 | Enables ship interior visualization where crew lives |
| Legacy Impact | N/A | 0 | No direct impact on legacy systems |
| Hard Sci-Fi Authenticity | 0 | 0 | Neutral - isometric is a visual style choice |
| **Net Score** | | **+1** | **Decision: Move forward** |

**Rationale:** This is infrastructure enabling the ship interior view. The isometric perspective will make the ship feel like a real, navigable space where crew lives and works. No negative impacts on any pillar.

**Reference:** [ailang-and-engine-ui.md](ailang-and-engine-ui.md) - Full architecture spec

## Problem Statement

**Current State:**
- Engine renders simple sprites/rects at world coordinates with linear zoom
- No isometric projection for 2.5D ship interior view
- DrawCmd has only 3 variants (Sprite, Rect, Text) - no tile-aware types
- Hit testing returns screen coords, not tile positions
- No UI element system (buttons, panels)

**Impact:**
- Cannot render ship interior as isometric grid
- Cannot implement tile-based click interactions
- Cannot build proper HUD/UI overlay system

## Goals

**Primary Goal:** Implement isometric projection and tile-aware rendering so AILANG can describe the world in logical tile coordinates while the engine handles all pixel-space concerns.

**Success Metrics:**
- Tiles at `(3, 5, height=1)` render at correct isometric screen position
- Clicking on screen returns correct `(tileX, tileY)` coordinates
- Entities sort correctly by layer + depth (entities in front occlude those behind)
- UI elements render in screen space, independent of camera

## Solution Design

### Overview

Extend the existing render pipeline with:
1. **Isometric projection** - Convert tile coords ↔ screen coords
2. **New DrawCmd variants** - `IsoTile`, `IsoEntity`, `UiElement`
3. **Composite sorting** - Sort by `(layer, screenY)` not just Z
4. **UI renderer** - Render panels/buttons in normalized screen space
5. **Tile hit testing** - Convert clicks to tile coordinates

### Architecture

```
AILANG (sim_gen mock)                Engine (Go/Ebiten)
─────────────────────                ──────────────────
World state in tile coords    →      TileToScreen() projection
DrawIsoTile{tile:{3,5}, h:0}  →      Render at pixel (px, py)
DrawIsoEntity{...}            →      Sort by (layer, screenY)
DrawUi{...}                   →      Layout in screen space
                              ←      ScreenToTile() for clicks
                              ←      UI hit testing first
```

**Key principle:** AILANG never thinks in pixels; engine never owns game rules.

**Components:**

1. **Iso Projection** (`engine/render/iso.go`): Pure math for tile↔screen conversion
2. **Extended Protocol** (`sim_gen/protocol.go`): New DrawCmd variants and types
3. **Iso Renderer** (`engine/render/iso.go`): Render IsoTile/IsoEntity with projection
4. **UI Renderer** (`engine/render/ui.go`): Render UiElement in screen space
5. **Enhanced Input** (`engine/render/input.go`): Tile-based click detection

### Implementation Plan

**Phase 1: Protocol & Projection**
- [ ] Add coordinate types to `sim_gen/protocol.go`: `Coord`, `IsoTile`, `IsoEntity`, `UiElement`
- [ ] Add new DrawCmd variants: `DrawCmdIsoTile`, `DrawCmdIsoEntity`, `DrawCmdUi`
- [ ] Create `engine/render/iso.go` with `TileToScreen()` and `ScreenToTile()`
- [ ] Add projection constants: `TileWidth`, `TileHeight`, `HeightScale`
- [ ] Write unit tests for projection math (round-trip accuracy)

**Phase 2: Iso Renderer**
- [ ] Implement `renderIsoTile()` - project tile, draw sprite/rect at position
- [ ] Implement `renderIsoEntity()` - project entity with sub-tile offset
- [ ] Update `RenderFrame()` to handle new DrawCmd variants
- [ ] Implement composite sorting: `sort by (layer, screenY, id)`
- [ ] Add viewport culling for iso tiles (check if projected pos in view)

**Phase 3: UI System**
- [ ] Create `engine/render/ui.go` with UI layout logic
- [ ] Implement `renderUiElement()` for Panel, Button, Label, Portrait
- [ ] Add normalized coordinate system (0..1) mapped to screen
- [ ] UI elements render last (highest layer), not camera-transformed

**Phase 4: Input & Hit Testing**
- [ ] Add `ScreenToTile()` to convert mouse clicks to tile coords
- [ ] Implement UI hit testing (check UI elements before world)
- [ ] Extend `FrameInput` with `TileClicks` and `UiClicks` fields
- [ ] Add `UiMode` to `FrameInput` for mode-aware input handling

**Phase 5: Integration & Testing**
- [ ] Update mock `sim_gen` to emit `DrawIsoTile` commands
- [ ] Create simple test scene: 5x5 tile grid with one entity
- [ ] Verify click→tile→highlight cycle works
- [ ] Test entity depth sorting (entity at y=3 in front of y=2)

### Files to Modify/Create

**New files:**
- `engine/render/iso.go` - Isometric projection & tile rendering (~150 LOC)
- `engine/render/ui.go` - UI element rendering & layout (~100 LOC)
- `engine/render/iso_test.go` - Projection unit tests (~80 LOC)

**Modified files:**
- `sim_gen/protocol.go` - Add Coord, IsoTile, IsoEntity, UiElement, new DrawCmd variants (~80 LOC)
- `sim_gen/step.go` - Update mock to emit iso commands (~30 LOC)
- `engine/render/draw.go` - Dispatch new DrawCmd variants, composite sorting (~40 LOC)
- `engine/render/input.go` - Tile click detection, UI hit testing (~50 LOC)

## Examples

### Example 1: Tile Projection

**Projection formula:**
```go
// TileToScreen converts tile coordinates to screen pixels
func TileToScreen(tile Coord, height int, cam Camera) (float64, float64) {
    // Standard isometric projection
    screenX := float64(tile.X-tile.Y) * (TileWidth / 2)
    screenY := float64(tile.X+tile.Y) * (TileHeight / 2) - float64(height)*HeightScale

    // Apply camera offset and zoom
    screenX = (screenX - cam.CenterX) * cam.Zoom + ScreenWidth/2
    screenY = (screenY - cam.CenterY) * cam.Zoom + ScreenHeight/2

    return screenX, screenY
}
```

### Example 2: DrawCmd Emission (AILANG/Mock)

**Before (current):**
```go
// Render player at world position
cmds = append(cmds, DrawCmd{
    Tag: DrawCmdTagSprite,
    Sprite: &DrawCmdSprite{ID: 1, X: 48.0, Y: 80.0, Z: 100},
})
```

**After (with iso support):**
```go
// Render player at tile position - engine handles projection
cmds = append(cmds, DrawCmd{
    Tag: DrawCmdTagIsoEntity,
    IsoEntity: &DrawCmdIsoEntity{
        ID:       "player",
        Tile:     Coord{X: 3, Y: 5},
        OffsetX:  0.0,
        OffsetY:  0.0,
        Height:   0,
        SpriteID: 1,
        Layer:    100,
    },
})
```

### Example 3: Tile Click Detection

**Input flow:**
```
Mouse click at (412, 298) pixels
        ↓
UI hit test: no UI element at that position
        ↓
ScreenToTile(412, 298, camera) → Coord{X: 3, Y: 5}
        ↓
FrameInput.TileClicks = [{Tile: {3,5}, Click: Left}]
        ↓
AILANG receives tile-based click event
```

## Success Criteria

- [ ] `TileToScreen(Coord{3,5}, 0, cam)` returns correct pixel position
- [ ] `ScreenToTile(px, py, cam)` round-trips accurately (±0.5 tile tolerance)
- [ ] Grid of 10x10 tiles renders as diamond-shaped iso view
- [ ] Entity at `(3, 4)` renders in front of entity at `(3, 3)`
- [ ] Clicking on tile (5, 5) returns `TileClick{Tile: {5,5}}`
- [ ] UI panel at `(0.1, 0.1, 0.3, 0.2)` renders in top-left, unaffected by camera
- [ ] All existing tests pass (no regression)
- [ ] Manual test: arrow keys move camera, tiles project correctly at all zoom levels

## Testing Strategy

**Unit tests:**
- Projection round-trip: `ScreenToTile(TileToScreen(t)) ≈ t` for various tiles
- Sorting: verify `(layer=0, y=5)` sorts before `(layer=0, y=3)`
- UI layout: normalized `(0.5, 0.5)` maps to screen center

**Integration tests:**
- Mock emits 25 IsoTile commands, verify all render
- Mock emits overlapping entities, verify correct depth order
- Simulate click, verify correct tile returned in next FrameInput

**Manual testing:**
- Visual: tiles form correct diamond grid pattern
- Interactive: click on tiles, see selection highlight on correct tile
- Camera: pan/zoom, verify projection stays correct

## Non-Goals

**Not in this feature:**
- **Sprite assets** - Use colored rects as placeholders, sprites added later
- **Animation** - Static rendering only, animation system is separate feature
- **Planet surface rendering** - Ship interior first, planets reuse same system
- **Galaxy map** - Non-isometric, different renderer needed
- **Dialogue UI** - Complex UI layouts deferred to separate feature

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Projection math errors | High | Extensive unit tests, visual debugging |
| Sorting edge cases | Med | Use stable sort, test overlapping entities |
| UI coordinate confusion | Med | Clear naming: `normalized` vs `screen` vs `tile` |
| Performance with many tiles | Low | Viewport culling already planned |

## References

- [ailang-and-engine-ui.md](ailang-and-engine-ui.md) - Full architecture specification
- [DEVELOPMENT.md](../../../DEVELOPMENT.md) - Data flow and type reference
- Existing render code: `engine/render/draw.go`, `engine/camera/transform.go`

## Future Work

- **Sprite integration** - Load actual iso sprites from manifest
- **Animation system** - Frame-based sprite animation
- **Height levels** - Multi-floor ship with walkable roofs
- **Lighting/shadows** - Optional visual polish
- **Planet mode** - Reuse iso renderer for planetary surfaces

---

**Document created**: 2025-12-01
**Last updated**: 2025-12-01
