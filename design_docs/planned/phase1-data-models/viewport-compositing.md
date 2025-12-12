# Viewport Compositing System

**Status**: Planned
**Target**: v0.4.0
**Priority**: P0 (Foundation for Bridge Interior)
**Estimated**: 2-3 days
**Dependencies**: Isometric Depth & Parallax System
**Enables**: Bridge Interior, Observation Windows, Dome Views

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Infrastructure feature |
| Civilization Simulation | N/A | 0 | Infrastructure feature |
| Philosophical Depth | + | +1 | Dome view creates contemplative spaces |
| Ship & Crew Life | + | +1 | Windows/dome are key emotional spaces |
| Legacy Impact | N/A | 0 | Infrastructure feature |
| Hard Sci-Fi Authenticity | + | +1 | Shows SR/GR effects through viewports |
| **Net Score** | | **+3** | **Decision: Move forward** |

**Feature type:** Engine/Infrastructure

**Rationale:** The observation dome is described as the "crown jewel" for SR/GR visuals in the bubble ship design. Being able to embed a space view (with relativistic effects) inside an isometric interior is essential for the bridge experience. This supports "Ship is Home" by making the connection between interior life and the cosmic journey visible.

**Reference:** See [bubble-ship-design.md](../../input/bubble-ship-design.md) - "TOP LEVEL — Bridge: Strongest aberration, Starfield compressed into forward cone"

## Problem Statement

The bridge design requires embedding a 3D space view (planets, stars, SR/GR effects) inside a 2D isometric interior through a dome-shaped viewport. Currently there's no way to:
- Clip/mask rendered content to arbitrary shapes
- Composite different render types (3D space + 2D isometric)
- Apply effects selectively (SR aberration in dome, not on floor)

**Current State:**
- Space view renders full-screen with SR/GR shaders
- Isometric interior renders separately
- No masking or compositing system
- No way to combine these in one scene

**Impact:**
- Bridge interior cannot show observation dome with space view
- Windows throughout the ship cannot show exterior
- The feeling of being inside a transparent bubble is lost
- SR/GR effects are all-or-nothing, not spatially targeted

## Goals

**Primary Goal:** Enable rendering content through shaped viewports (domes, windows, portholes) that composite different render sources with appropriate visual effects.

**Success Metrics:**
- Dome-shaped viewport renders space view inside isometric bridge
- SR warp shader applies only within dome bounds
- Smooth edge blending between dome and interior
- Multiple viewports supported (dome + windows)
- 60 FPS with dome + full bridge interior

## Solution Design

### Overview

Introduce a **Viewport System** that:
1. Defines shaped regions (dome, rectangle, circle, polygon)
2. Renders content to those regions with masking
3. Composites viewport content into the scene layer
4. Applies shaders/effects per-viewport

```
┌─────────────────────────────────────────────────────┐
│                    Bridge Interior                    │
│              ┌─────────────────────┐                  │
│             ╱   OBSERVATION DOME    ╲                 │
│            │   ┌───────────────┐    │                 │
│            │   │ 🪐  Space View │    │  ← Viewport    │
│            │   │  with SR/GR   │    │     composited  │
│            │   │   effects     │    │     into scene  │
│            │   └───────────────┘    │                 │
│             ╲_______________________╱                  │
│                                                       │
│      [Console]    [Console]    [Console]              │
└───────────────────────────────────────────────────────┘
```

### Architecture

**Key Concepts:**

1. **Viewport**: A shaped region that renders specific content
2. **ViewportMask**: Shape definition (ellipse, polygon, rect)
3. **ViewportContent**: What renders inside (space view, camera feed, etc.)
4. **ViewportEffect**: Per-viewport shader/post-processing
5. **ViewportCompositor**: Combines viewports into final scene

```
                    ┌──────────────────┐
                    │ ViewportCompositor│
                    └────────┬─────────┘
                             │
           ┌─────────────────┼─────────────────┐
           │                 │                 │
    ┌──────┴──────┐   ┌──────┴──────┐   ┌──────┴──────┐
    │  Viewport 0 │   │  Viewport 1 │   │  Viewport 2 │
    │    (Dome)   │   │  (Porthole) │   │   (Window)  │
    └──────┬──────┘   └──────┬──────┘   └──────┬──────┘
           │                 │                 │
    ┌──────┴──────┐   ┌──────┴──────┐   ┌──────┴──────┐
    │ SpaceView   │   │ SpaceView   │   │ CameraFeed  │
    │ + SR Warp   │   │ (no warp)   │   │  (cargo)    │
    └─────────────┘   └─────────────┘   └─────────────┘
```

**Components:**

1. **ViewportManager** (`engine/render/viewport.go`): Manages viewport definitions and lifecycle
2. **ViewportMask** (`engine/render/viewport_mask.go`): Shape generation and mask textures
3. **ViewportRenderer** (`engine/render/viewport_render.go`): Content rendering with masking
4. **EdgeBlender** (`engine/shader/edge_blend.kage`): Smooth transition at viewport edges

### AILANG Types

```ailang
module sim/viewport

import sim/protocol (DrawCmd, Coord)

-- Shape definitions for viewports
type ViewportShape =
    | ShapeEllipse(centerX: float, centerY: float, radiusX: float, radiusY: float)
    | ShapeCircle(centerX: float, centerY: float, radius: float)
    | ShapeRect(x: float, y: float, width: float, height: float)
    | ShapePolygon(points: [(float, float)])
    | ShapeDome(centerX: float, centerY: float, width: float, height: float, archHeight: float)

-- What content fills the viewport
type ViewportContent =
    | ContentSpaceView(velocity: float, viewAngle: float, targetPlanet: Option[int])
    | ContentStarfield(density: float, scroll: bool)
    | ContentCamera(cameraId: int)  -- For security feeds, etc.
    | ContentSolid(color: int)      -- Solid color fill

-- Effects applied within viewport
type ViewportEffect =
    | EffectNone
    | EffectSRWarp(velocity: float)
    | EffectGRLensing(mass: float, distance: float)
    | EffectTint(color: int, intensity: float)
    | EffectBlur(radius: float)

-- Complete viewport definition
type Viewport = {
    id: string,
    shape: ViewportShape,
    content: ViewportContent,
    effects: [ViewportEffect],
    layer: int,           -- Z-order among viewports
    edgeBlend: float,     -- 0.0 = hard edge, 1.0 = soft blend
    opacity: float        -- 0.0-1.0
}

-- Draw command for viewport
type DrawCmdViewport = {
    viewport: Viewport,
    screenX: float,       -- Position in screen space
    screenY: float,
    z: int
}

-- Bridge observation dome definition
pure func bridgeObservationDome(state: DomeViewState) -> Viewport {
    Viewport({
        id: "bridge_dome",
        shape: ShapeDome(640.0, 200.0, 400.0, 250.0, 80.0),
        content: ContentSpaceView(state.velocity, state.viewAngle, state.targetPlanet),
        effects: [EffectSRWarp(state.velocity)],
        layer: 90,
        edgeBlend: 0.15,
        opacity: 1.0
    })
}

-- Small window viewport
pure func cabinWindow(x: float, y: float, facing: float) -> Viewport {
    Viewport({
        id: "cabin_window",
        shape: ShapeRect(x, y, 64.0, 48.0),
        content: ContentStarfield(0.8, true),
        effects: [],
        layer: 20,
        edgeBlend: 0.05,
        opacity: 1.0
    })
}
```

### Go Engine Implementation

```go
// engine/render/viewport.go

type ViewportShape interface {
    GenerateMask(w, h int) *ebiten.Image
    Contains(x, y float64) bool
    Bounds() (x, y, w, h float64)
}

type EllipseShape struct {
    CenterX, CenterY float64
    RadiusX, RadiusY float64
}

func (e *EllipseShape) GenerateMask(w, h int) *ebiten.Image {
    mask := ebiten.NewImage(w, h)
    // Draw filled ellipse to mask using vector graphics
    // White = visible, transparent = masked
    return mask
}

type DomeShape struct {
    CenterX, CenterY float64
    Width, Height    float64
    ArchHeight       float64  // How tall the curved top is
}

func (d *DomeShape) GenerateMask(w, h int) *ebiten.Image {
    // Dome = rectangle bottom + elliptical arch top
    mask := ebiten.NewImage(w, h)
    // Draw dome shape
    return mask
}
```

```go
// engine/render/viewport_render.go

type ViewportRenderer struct {
    contentBuffer *ebiten.Image  // Content rendered here
    maskBuffer    *ebiten.Image  // Mask applied
    blendShader   *ebiten.Shader // Edge blending
    spaceRenderer *SpaceRenderer // For ContentSpaceView
    srWarp        *SRWarp        // SR effects
}

func (r *ViewportRenderer) RenderViewport(viewport Viewport) *ebiten.Image {
    // 1. Generate or retrieve cached mask
    mask := viewport.shape.GenerateMask(viewport.bounds)

    // 2. Render content to buffer
    r.contentBuffer.Clear()
    switch viewport.content {
    case ContentSpaceView:
        r.spaceRenderer.Draw(r.contentBuffer, viewport.content)
    case ContentStarfield:
        r.drawStarfield(r.contentBuffer, viewport.content)
    }

    // 3. Apply viewport-specific effects
    for _, effect := range viewport.effects {
        r.applyEffect(r.contentBuffer, effect)
    }

    // 4. Apply mask with edge blending
    result := ebiten.NewImage(bounds.W, bounds.H)
    r.applyMaskWithBlend(result, r.contentBuffer, mask, viewport.edgeBlend)

    return result
}

func (r *ViewportRenderer) applyMaskWithBlend(
    dst, src, mask *ebiten.Image,
    blendAmount float64,
) {
    // Use edge blend shader for soft transitions
    opts := &ebiten.DrawRectShaderOptions{}
    opts.Uniforms["BlendAmount"] = blendAmount
    opts.Images[0] = src
    opts.Images[1] = mask
    dst.DrawRectShader(w, h, r.blendShader, opts)
}
```

```go
// engine/render/viewport_compositor.go

type ViewportCompositor struct {
    viewports []Viewport
    renderer  *ViewportRenderer
    cache     map[string]*ebiten.Image  // Cached rendered viewports
}

func (c *ViewportCompositor) CompositeToLayer(
    layer *ebiten.Image,
    viewports []Viewport,
) {
    // Sort by Z-order
    sort.Slice(viewports, func(i, j int) bool {
        return viewports[i].layer < viewports[j].layer
    })

    // Render and composite each
    for _, vp := range viewports {
        rendered := c.renderer.RenderViewport(vp)
        opts := &ebiten.DrawImageOptions{}
        opts.GeoM.Translate(vp.screenX, vp.screenY)
        opts.ColorScale.ScaleAlpha(float32(vp.opacity))
        layer.DrawImage(rendered, opts)
    }
}
```

### Edge Blend Shader

```kage
//kage:unit pixels
package main

var BlendAmount float

func Fragment(dst vec4, src vec2, color vec4) vec4 {
    content := imageSrc0At(src)
    mask := imageSrc1At(src).r

    // Soft edge based on mask gradient and blend amount
    edge := smoothstep(0.0, BlendAmount, mask)

    return vec4(content.rgb, content.a * edge)
}
```

### Dome Shape Details

The observation dome is not a simple ellipse - it's a dome shape (rectangular bottom + curved top):

```
        ╭─────────────────╮  ← Curved arch (elliptical)
       ╱                   ╲
      ╱                     ╲
     │                       │
     │    SPACE CONTENT      │  ← Rectangular body
     │                       │
     └───────────────────────┘  ← Flat bottom (frame edge)
```

```go
func (d *DomeShape) GenerateMask(w, h int) *ebiten.Image {
    mask := ebiten.NewImage(w, h)

    // Path for dome shape
    path := &vector.Path{}

    // Start at bottom-left
    path.MoveTo(0, h)
    // Left edge up to arch start
    path.LineTo(0, d.ArchHeight)
    // Arch curve (quadratic bezier or arc)
    path.QuadTo(w/2, 0, w, d.ArchHeight)
    // Right edge down
    path.LineTo(w, h)
    // Close
    path.Close()

    // Fill path with white
    vector.DrawFilledPath(mask, path, color.White, false, vector.FillRuleNonZero)

    return mask
}
```

### Implementation Plan

**Phase 1: Core Viewport Types** (~3 hours)
- [ ] Create `engine/render/viewport.go` with shape interfaces
- [ ] Implement EllipseShape, CircleShape, RectShape
- [ ] Implement DomeShape with arch curve
- [ ] Add Viewport types to AILANG protocol

**Phase 2: Mask Generation** (~3 hours)
- [ ] Implement `GenerateMask()` for each shape
- [ ] Add mask caching (shapes rarely change)
- [ ] Test mask generation with visual debug output
- [ ] Handle high-DPI scaling

**Phase 3: Content Rendering** (~4 hours)
- [ ] Create ViewportRenderer with content buffer
- [ ] Implement ContentSpaceView (reuse existing SpaceRenderer)
- [ ] Implement ContentStarfield (simple parallax stars)
- [ ] Wire up SR warp shader as viewport effect

**Phase 4: Edge Blending** (~2 hours)
- [ ] Create edge blend shader (edge_blend.kage)
- [ ] Implement soft edge masking
- [ ] Test with different blend amounts (0.05, 0.15, 0.3)
- [ ] Optimize for performance

**Phase 5: Integration** (~2 hours)
- [ ] Create ViewportCompositor
- [ ] Integrate with depth layer system
- [ ] Add DrawCmdViewport handling
- [ ] Test with bridge dome + interior

### Files to Modify/Create

**New files:**
- `engine/render/viewport.go` - Viewport types and shapes (~150 LOC)
- `engine/render/viewport_render.go` - Content rendering (~120 LOC)
- `engine/render/viewport_compositor.go` - Compositing logic (~80 LOC)
- `engine/shader/edge_blend.kage` - Edge blend shader (~30 LOC)
- `sim/viewport.ail` - AILANG viewport types (~60 LOC)

**Modified files:**
- `sim/protocol.ail` - Add DrawCmdViewport, ViewportContent (~40 LOC)
- `engine/render/draw.go` - Handle viewport commands (~30 LOC)
- `engine/render/depth_layers.go` - Viewport layer integration (~20 LOC)

## Examples

### Example 1: Bridge Observation Dome

**Goal:** Render space view through dome-shaped viewport on bridge.

```ailang
pure func renderBridgeScene(state: BridgeState) -> [DrawCmd] {
    -- Interior elements
    let floorCmds = renderBridgeFloor();
    let consoleCmds = renderConsoles(state.consoleStates);
    let crewCmds = renderBridgeCrew(state.crewPositions);

    -- Observation dome viewport
    let dome = bridgeObservationDome(state.domeView);
    let domeCmd = DrawCmdViewport({
        viewport: dome,
        screenX: 440.0,  -- Centered horizontally
        screenY: 50.0,   -- Near top
        z: 90
    });

    -- Compose: floor, dome, consoles, crew
    floorCmds ++ [domeCmd] ++ consoleCmds ++ crewCmds
}
```

**Visual Result:**
```
┌─────────────────────────────────────────────────┐
│                                                 │
│              ╭──────────────────╮               │
│             ╱  🪐               ╲              │
│            │    ✦  ✦    ✦  ✦    │  ← Space view │
│            │  ✦     Earth    ✦  │    with SR    │
│             ╲   approaching   ╱                │
│              ╰──────────────────╯               │
│                (soft blended edge)              │
│                                                 │
│    [HELM]      [COMMS]      [NAV]              │
│      👤          👤          👤                 │
│                                                 │
│                 [CAPTAIN]                       │
└─────────────────────────────────────────────────┘
```

### Example 2: Cabin Window

**Goal:** Small rectangular window in crew quarters showing stars.

```ailang
pure func renderCabinWithWindow(cabin: CabinState) -> [DrawCmd] {
    let interiorCmds = renderCabinInterior(cabin);

    -- Small window
    let window = Viewport({
        id: "cabin_window_" ++ intToString(cabin.id),
        shape: ShapeRect(cabin.windowX, cabin.windowY, 48.0, 32.0),
        content: ContentStarfield(0.6, true),
        effects: [],
        layer: 20,
        edgeBlend: 0.02,  -- Sharp edge for window frame
        opacity: 1.0
    });

    interiorCmds ++ [DrawCmdViewport({viewport: window, screenX: 100.0, screenY: 80.0, z: 20})]
}
```

### Example 3: Multiple Viewports

**Goal:** Bridge with dome + two side portholes.

```ailang
pure func renderBridgeFull(state: BridgeState) -> [DrawCmd] {
    -- Main dome
    let dome = bridgeObservationDome(state.domeView);

    -- Side portholes (no SR effect, just stars)
    let leftPorthole = Viewport({
        id: "porthole_left",
        shape: ShapeCircle(100.0, 200.0, 30.0),
        content: ContentStarfield(0.5, false),
        effects: [],
        layer: 15,
        edgeBlend: 0.1,
        opacity: 0.9
    });

    let rightPorthole = Viewport({
        id: "porthole_right",
        shape: ShapeCircle(1180.0, 200.0, 30.0),
        content: ContentStarfield(0.5, false),
        effects: [],
        layer: 15,
        edgeBlend: 0.1,
        opacity: 0.9
    });

    [domeCmd, leftPortholeCmd, rightPortholeCmd] ++ interiorCmds
}
```

## Success Criteria

- [ ] Dome-shaped viewport renders correctly (arch top, flat bottom)
- [ ] Space content visible through dome with SR warp applied
- [ ] Edge blending creates smooth transition (no hard pixel edges)
- [ ] Multiple viewports composite without artifacts
- [ ] Viewport content updates each frame (not static)
- [ ] Performance: 60 FPS with dome (256x192) + 16x12 bridge tiles
- [ ] Integrates with depth layer system (viewports in correct layer)

## Testing Strategy

**Unit tests:**
- Shape.GenerateMask() produces correct dimensions
- Shape.Contains() returns correct results
- Mask caching works (same shape returns cached mask)

**Integration tests:**
- Render dome with space content, capture screenshot
- Verify viewport respects layer ordering
- Test with moving camera (viewport stays fixed)

**Manual testing:**
- Visual: Dome looks like curved glass, not flat circle
- Visual: SR effects only appear within dome
- Visual: Edge blend is subtle but present
- Performance: Monitor FPS during gameplay

**Test Scenarios:**
```bash
# Dome viewport test
./bin/demo-game-bridge --dome-test

# Multiple viewports
./bin/demo-game-bridge --viewport-count 3

# Edge blend comparison
./bin/demo-viewport --blend 0.0 0.1 0.2 0.3
```

## Non-Goals

**Not in this feature:**
- **Dynamic viewport shapes** - Shapes are static, not animated
- **Viewport interaction** - Clicking in dome area is separate feature
- **Reflections** - No reflection/mirror viewports
- **Picture-in-picture** - Not a general PIP system
- **Video content** - ViewportContent is realtime render, not video

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Mask generation slow | Med | Cache masks, regenerate only on resize |
| Edge blend artifacts | Med | Test multiple blend amounts, use smooth gradient |
| SR shader conflicts | Med | Render to isolated buffer, composite result |
| High-DPI issues | Low | Scale mask generation to actual pixel size |

## References

- [isometric-depth-parallax.md](./isometric-depth-parallax.md) - Depth layer system (prerequisite)
- [bubble-ship-design.md](../../input/bubble-ship-design.md) - Dome visual requirements
- [02-bridge-interior.md](../phase2-core-views/02-bridge-interior.md) - Bridge design using viewports
- [engine-capabilities.md](../../reference/engine-capabilities.md) - SR warp shader details

## Future Work

- **Interactive Viewports** - Click in dome to access space navigation
- **Viewport Animations** - Open/close shutters, zoom effects
- **Reflective Surfaces** - Mirrors, polished metal showing environment
- **External View Mode** - Full-screen version of dome content

---

**Document created**: 2025-12-12
**Last updated**: 2025-12-12
