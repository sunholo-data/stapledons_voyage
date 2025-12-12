# Isometric Depth & Parallax System

**Status**: Implemented
**Version**: v0.3.0
**Priority**: P0 (Foundation for Bridge Interior)
**Actual Effort**: 4 days
**Dependencies**: Isometric Engine (implemented)
**Enables**: Bridge Interior, Ship Exploration, Multi-Level Navigation

## Summary

Added depth perception to isometric views through parallax layers and transparency, enabling views from ship interior through transparent bubble structure to space.

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Ship & Crew Life | + | +1 | Makes ship feel like real 3D space |
| Hard Sci-Fi Authenticity | + | +1 | Transparency matches bubble physics |
| **Net Score** | | **+2** | Infrastructure feature |

## Implementation

### Architecture

**20 depth layers (0-19)** with configurable parallax factors:

```
Layer 0  (0.00x): Fixed at infinity (galaxy/space)
Layer 1-5 (0.05-0.25x): Distant ship structures
Layer 6-9 (0.30-0.60x): Mid-distance decks
Layer 10-15 (0.70-0.95x): Near elements
Layer 16-18 (1.0x): Scene layers
Layer 19 (1.0x): UI (screen-fixed)
```

**Key Layers:**
- L0: Galaxy background (physically realistic - stars at infinity)
- L6: Spire silhouette (0.3x parallax)
- L16: Main scene (tiles, entities)
- L19: UI overlay

### Files Created

| File | Purpose | LOC |
|------|---------|-----|
| [engine/depth/layer.go](../../../engine/depth/layer.go) | DepthLayer enum & parallax factors | ~54 |
| [engine/render/depth_layers.go](../../../engine/render/depth_layers.go) | Layer buffer management & compositing | ~110 |
| [engine/camera/parallax.go](../../../engine/camera/parallax.go) | Parallax camera transforms | ~103 |
| [engine/camera/parallax_test.go](../../../engine/camera/parallax_test.go) | Unit tests (12 tests) | ~142 |
| [sim/depth.ail](../../../sim/depth.ail) | AILANG depth types | ~52 |

### Files Modified

| File | Changes |
|------|---------|
| [sim/protocol.ail](../../../sim/protocol.ail) | Added IsoTileAlpha, SpireBg DrawCmd variants |
| [engine/render/draw.go](../../../engine/render/draw.go) | Layer-aware rendering, parallax dispatch |
| [engine/render/draw_iso.go](../../../engine/render/draw_iso.go) | drawIsoTileAlpha with alpha/tint support |
| [engine/render/draw_shapes.go](../../../engine/render/draw_shapes.go) | drawSpireBg, drawSpireBgParallax |
| [engine/render/draw_stars.go](../../../engine/render/draw_stars.go) | drawGalaxyBackgroundParallax |

### AILANG Types

```ailang
type DepthLayer =
    | LayerDeepBackground   -- Parallax 0.0x (fixed)
    | LayerMidBackground    -- Parallax 0.3x
    | LayerScene            -- Parallax 1.0x
    | LayerForeground       -- Screen-fixed

type TransparentTile = {
    baseId: int,
    alpha: float,
    seeThroughLayer: DepthLayer,
    tintRgba: int
}
```

### New DrawCmd Variants

```ailang
| IsoTileAlpha(tile: Coord, height: int, spriteId: int, layer: int, alpha: float, tintRgba: int)
| SpireBg(z: int)
| Marker(x: float, y: float, w: float, h: float, rgba: int, parallaxLayer: int, z: int)
```

**Marker:** Allows AILANG to place elements on any of the 20 parallax layers (0-19) via the `parallaxLayer` field.

### Engine APIs

```go
// Enable the layer system
renderer.EnableLayers(screenW, screenH)

// Handle resize
renderer.ResizeLayers(screenW, screenH)
```

## Demo

Run the parallax demo to see the system in action:

```bash
# Interactive mode (arrow keys to pan, +/- to zoom)
go run ./cmd/demo-parallax

# Take a screenshot
go run ./cmd/demo-parallax --screenshot 30 --output out/screenshots/parallax.png

# Screenshot with specific camera position
go run ./cmd/demo-parallax -camx 300 --screenshot 1 --output test.png
```

The demo shows:
- Layer 0: Galaxy background (0.1x parallax)
- Layer 1: Spire silhouette (0.3x parallax)
- Layer 2: Isometric tiles with transparency
- Layer 3: HUD overlay

## AILANG/Engine Boundary

This respects the AILANG/Engine divide:

| Component | Owner | Rationale |
|-----------|-------|-----------|
| DrawCmd types (IsoTileAlpha, SpireBg) | AILANG | WHAT to draw |
| Layer classification | Engine | HOW to render (visual only) |
| Parallax factors | Engine | Pure rendering detail |
| Compositing order | Engine | Visual polish |

AILANG doesn't need to know about parallax - it emits DrawCmds, the engine applies visual effects.

## Success Criteria - All Met

- [x] Space background renders behind transparent tiles
- [x] Parallax: background layers move at 0.1-0.3x camera speed
- [x] Multiple transparency levels composite correctly
- [x] Spire visible as background element at 0.3x parallax
- [x] Layer infrastructure: 4-layer system with separate buffers
- [x] 12 unit tests passing for parallax math

## References

- Sprint: [sprints/isometric-depth-parallax-sprint.md](../../../sprints/isometric-depth-parallax-sprint.md)
- Demo: [cmd/demo-parallax/main.go](../../../cmd/demo-parallax/main.go)
- Demo utility: [demo-screenshot-utility.md](demo-screenshot-utility.md)

---

**Document created**: 2025-12-12
**Implemented**: 2025-12-12
