# Sprint: Viewport Compositing System

**Design Doc:** [design_docs/planned/phase1-data-models/viewport-compositing.md](../design_docs/planned/phase1-data-models/viewport-compositing.md)
**Duration:** 2-3 days
**Priority:** P0 (Foundation for Bridge)
**Dependencies:** Isometric Depth & Parallax System (must complete first)

## Goal

Enable rendering content through shaped viewports (domes, windows, portholes) that composite different render sources with appropriate visual effects.

## Pre-Sprint Checklist

- [x] Verify Depth & Parallax sprint is complete ✅ (2025-12-12)
- [x] Verify DepthLayerManager working ✅ (20 layers, parallax functioning)
- [x] Verify SpaceBackground renders to Layer 0 ✅ (GalaxyBg → L0)
- [x] Check for unread AILANG messages: `ailang messages list --unread` ✅ (none)

## Day 1: Viewport Shapes & Masks (~5 hours)

### Tasks
- [ ] Create `engine/render/viewport.go` with shape interfaces
  ```go
  type ViewportShape interface {
      GenerateMask(w, h int) *ebiten.Image
      Contains(x, y float64) bool
      Bounds() (x, y, w, h float64)
  }
  ```
- [ ] Implement EllipseShape with vector graphics mask
- [ ] Implement CircleShape (special case of ellipse)
- [ ] Implement RectShape (simplest case)
- [ ] Implement DomeShape (rect bottom + elliptical arch top)
- [ ] Add mask caching (regenerate only on size change)

### Files to Create
- `engine/render/viewport.go` (~150 LOC) - NEW

### Verification
```bash
go build ./engine/render/...
go test ./engine/render/... -run Viewport
# Visual test: render masks to PNG for inspection
```

## Day 2: Content Rendering & Edge Blending (~5 hours)

### Tasks
- [ ] Create `engine/render/viewport_render.go` with ViewportRenderer
  - [ ] Content buffer for rendering
  - [ ] ContentSpaceView: reuse SpaceBackground.Draw()
  - [ ] ContentStarfield: simple parallax stars
- [ ] Create edge blend shader (`engine/shader/edge_blend.kage`)
  ```kage
  var BlendAmount float
  func Fragment(...) vec4 {
      content := imageSrc0At(src)
      mask := imageSrc1At(src).r
      edge := smoothstep(0.0, BlendAmount, mask)
      return vec4(content.rgb, content.a * edge)
  }
  ```
- [ ] Implement applyMaskWithBlend()
- [ ] Test with different blend amounts (0.05, 0.15, 0.3)

### Files to Create
- `engine/render/viewport_render.go` (~120 LOC) - NEW
- `engine/shader/edge_blend.kage` (~30 LOC) - NEW

### Verification
```bash
go build ./engine/render/...
./bin/game --test-viewport-blend
# Compare: Hard edge (0.0) vs soft edge (0.15)
```

## Day 3: AILANG Types & Integration (~4 hours)

### Tasks
- [ ] Add AILANG viewport types
  ```ailang
  type ViewportShape =
      | ShapeEllipse(...)
      | ShapeCircle(...)
      | ShapeRect(...)
      | ShapeDome(...)

  type ViewportContent =
      | ContentSpaceView(velocity: float, viewAngle: float)
      | ContentStarfield(density: float)

  type Viewport = { id: string, shape: ViewportShape, ... }
  ```
- [ ] Add DrawCmdViewport to protocol
- [ ] Create ViewportCompositor for multi-viewport scenes
- [ ] Wire SR warp shader as viewport effect
- [ ] Test: Bridge dome with space view + SR effects

### Files to Create/Modify
- `sim/viewport.ail` (~60 LOC) - NEW
- `sim/protocol.ail` - Add DrawCmdViewport (~40 LOC)
- `engine/render/viewport_compositor.go` (~80 LOC) - NEW
- `engine/render/draw.go` - Handle viewport commands (~30 LOC)

### Verification
```bash
ailang check sim/viewport.ail
make sim
./bin/game --test-dome
# Visual: Dome-shaped viewport with space, SR warp inside dome only
```

## Success Criteria

- [ ] Dome-shaped viewport renders correctly (arch top, flat bottom)
- [ ] Space content visible through dome with SR warp applied
- [ ] Edge blending creates smooth transition (no hard pixel edges)
- [ ] Multiple viewports composite without artifacts
- [ ] Viewport content updates each frame (not static)
- [ ] Performance: 60 FPS with dome (256x192) + 16x12 bridge tiles

## AILANG Feedback Checkpoint

After sprint, report:
- [ ] Any issues with ViewportShape ADT codegen
- [ ] Any float precision issues in mask generation
- [ ] Shader loading issues (if any)

## Handoff

This sprint enables:
- **Bridge Interior** - Observation dome now possible
- **Ship Windows** - Same viewport system works for cabin windows
- **Multi-Level** - Can show space through transparent bubble at any deck

---

**Sprint created:** 2025-12-12
**Status:** Ready for execution (after Depth & Parallax sprint)
