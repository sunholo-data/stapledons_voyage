# Sprint: Isometric Depth & Parallax System

**Design Doc:** [design_docs/implemented/v0_3_0/isometric-depth-parallax.md](../design_docs/implemented/v0_3_0/isometric-depth-parallax.md)
**Duration:** 3-4 days
**Priority:** P0 (Foundation for Bridge)
**Dependencies:** Existing SpaceBackground (engine/view/background/space.go)

## Goal

Add depth perception to isometric views through parallax layers and transparency, enabling views from ship interior through transparent bubble structure to space.

## Pre-Sprint Checklist

- [x] Verify `ailang check sim/protocol.ail` passes
- [x] Review SpaceBackground implementation in `engine/view/background/space.go`
- [x] Check for unread AILANG messages: `ailang messages list --unread`
- [x] Ack any blocking issues before starting

## Day 1: Layer Infrastructure (~4 hours) ✅ COMPLETE

### Tasks
- [x] Create `engine/render/depth_layers.go` with DepthLayerManager
  - [x] Define DepthLayer enum (DeepBackground, MidBackground, Scene, Foreground)
  - [x] Implement layer buffer creation/management
  - [x] Implement Composite() for back-to-front blending
- [x] Create `engine/camera/parallax.go` with ParallaxCamera
  - [x] Extend CameraOffset interface from background package
  - [x] Implement ForLayer() returning layer-adjusted position
  - [x] Implement TransformForLayer() for rendering
- [x] Unit tests for parallax math (12 tests passing)

### Files Created
- `engine/depth/layer.go` (~54 LOC) - NEW (shared package to avoid import cycles)
- `engine/render/depth_layers.go` (~110 LOC) - NEW
- `engine/camera/parallax.go` (~103 LOC) - NEW
- `engine/camera/parallax_test.go` (~142 LOC) - NEW

### Verification
```bash
go build ./engine/render/...  # ✅ Passes
go test ./engine/camera/... -run Parallax  # ✅ 12 tests pass
```

## Day 2: Protocol Extension & Layer Rendering (~4 hours) ✅ COMPLETE

### Tasks
- [x] Add DepthLayer type to AILANG protocol
  ```ailang
  type DepthLayer =
      | LayerDeepBackground
      | LayerMidBackground
      | LayerScene
      | LayerForeground
  ```
- [x] Add helper functions: layerParallax, layerName, layerZBase
- [x] Extend RenderFrame() to sort commands by layer
- [x] Implement layer buffer rendering in draw.go
- [x] Wire SpaceBackground/GalaxyBg to Layer 0 (DeepBackground)

### Files Created/Modified
- `sim/depth.ail` (~52 LOC) - NEW (AILANG depth layer types)
- `engine/render/draw.go` (~150 LOC changes) - Added layered rendering path

### Verification
```bash
ailang check sim/depth.ail  # ✅ No errors
make sim                     # ✅ Compiles to sim_gen/depth.go
go build ./...               # ✅ Passes
go test ./engine/...         # ✅ All tests pass
```

## Day 3: Transparency Tiles (~4 hours) ✅ COMPLETE

### Tasks
- [x] Add TransparentTile type to AILANG
  ```ailang
  type TransparentTile = {
      baseId: int,
      alpha: float,
      seeThroughLayer: DepthLayer,
      tintRgba: int
  }
  ```
- [x] Add IsoTileAlpha DrawCmd variant to protocol
- [x] Implement alpha rendering for IsoTile in engine
- [x] Add helper functions: glassFloorTile, domeEdgeTile, frostedPanelTile

### Files Modified
- `sim/depth.ail` (~45 LOC added) - TransparentTile type and preset functions
- `sim/protocol.ail` - Added IsoTileAlpha DrawCmd variant
- `engine/render/draw_iso.go` (~60 LOC) - drawIsoTileAlpha with alpha/tint support
- `engine/render/draw.go` - Added IsoTileAlpha cases to both render paths

### Verification
```bash
ailang check sim/depth.ail    # ✅ Passes
ailang check sim/protocol.ail # ✅ Passes
make sim                       # ✅ Generates sim_gen with IsoTileAlpha
go build ./...                 # ✅ Passes
go test ./engine/...           # ✅ All tests pass
```

## Day 4: Spire Background & Integration (~4 hours) ✅ COMPLETE

### Tasks
- [x] Create spire silhouette placeholder (drawSpireBg function)
- [x] Add SpireBg DrawCmd to protocol
- [x] Route SpireBg to MidBackground layer (0.3x parallax)
- [x] Add rendering cases to both render paths

### Files Modified
- `sim/protocol.ail` - Added SpireBg DrawCmd variant
- `engine/render/draw_shapes.go` (~45 LOC) - drawSpireBg placeholder
- `engine/render/draw.go` - Layer routing and render cases for SpireBg

### Verification
```bash
make sim                       # ✅ Generates sim_gen with SpireBg
go build ./...                 # ✅ Passes
go test ./engine/...           # ✅ All tests pass
```

## Success Criteria ✅

- [x] Space background renders behind transparent tiles (GalaxyBg → DeepBackground)
- [x] Parallax system: background layers move at 0.1-0.3x camera speed
- [x] Multiple transparency levels composite correctly (IsoTileAlpha with alpha/tint)
- [x] Spire visible as background element (SpireBg → MidBackground at 0.3x)
- [x] Layer infrastructure: 4-layer system with separate buffers

## AILANG Feedback Checkpoint

- [x] No type inference issues with DepthLayer (ADT codegen works)
- [x] No codegen bugs with alpha/float fields (TransparentTile works)
- [x] No blocking issues encountered

## Handoff

This sprint enables:
- **Viewport Compositing** - Can now render space to layer, composite with masks
- **Multi-Level Ship** - Spire rendering pattern established
- **Bridge Interior** - Transparent floor + space background working

### New Engine APIs:
```go
renderer.EnableLayers(screenW, screenH)  // Enable layer system
renderer.ResizeLayers(screenW, screenH)  // Handle resize
```

### New AILANG Types:
```ailang
type DepthLayer = LayerDeepBackground | LayerMidBackground | LayerScene | LayerForeground
type TransparentTile = { baseId: int, alpha: float, seeThroughLayer: DepthLayer, tintRgba: int }
```

### New DrawCmd Variants:
```ailang
| IsoTileAlpha(tile: Coord, height: int, spriteId: int, layer: int, alpha: float, tintRgba: int)
| SpireBg(z: int)
```

---

**Sprint created:** 2025-12-12
**Sprint completed:** 2025-12-12
**Status:** ✅ COMPLETE
