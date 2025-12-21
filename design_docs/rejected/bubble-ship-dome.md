# Bubble Ship Dome Rendering

---

**⚠️ REJECTED: 2025-12-20**

**Reason:** 3D dome rendering approach abandoned in favor of scene-based 2D/2.5D navigation.

**Why rejected:**
- 3D asset creation too complex for interior spaces
- Game focus is conversations and space visualization, not 3D spatial exploration
- AI-generated 2D images are faster to produce and more flexible
- Dome concept preserved but implemented differently: outward-facing deck scenes show space through windows

**Replacement:** See design decision "[2025-12-20] Interior Ship Experience: Scene-Based Navigation" in [docs/vision/design-decisions.md](../../docs/vision/design-decisions.md)

---

**Status**: Rejected
**Target**: v0.2.0
**Priority**: P1 - Core game concept
**Estimated**: 3-4 days
**Dependencies**: Window3D system (implemented), Tetra3D integration (implemented)

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Choices Are Final | N/A | 0 | Infrastructure feature |
| The Game Doesn't Judge | N/A | 0 | Visual system, no moral content |
| Time Has Emotional Weight | + | +1 | Watching stars shift/civilizations change through the dome reinforces temporal isolation |
| The Ship Is Home | ++ | +2 | The dome IS the ship experience - observation decks, crew gathering, intimate vs vast |
| Grounded Strangeness | + | +1 | Accurate physics in dome view grounds the strange cosmic experience |
| We Are Not Built For This | ++ | +2 | The vastness visible through transparent dome confronts human frailty |
| **Net Score** | | **+6** | **Decision: Move forward** |

**Feature type:** Gameplay/Engine hybrid
- Directly player-facing (they see the dome constantly)
- Enables core "bubble ship" concept from game vision

**Reference:** See [game-vision.md](../../docs/game-vision.md)

## Problem Statement

The ship in Stapledon's Voyage is a "bubble ship" - a transparent sphere offering panoramic views of space. Currently:

**Current State:**
- `Window3D` only supports flat rectangular planes
- Flat textures distort at wide fields of view (>90°)
- No support for curved/spherical projections
- Maximum texture resolution is 2048×2048 (fine for small windows, limiting for domes)

**Impact:**
- The core game fantasy of "floating in a bubble through space" is not achievable
- Observation decks feel like looking through small portholes, not panoramas
- Wide-angle views (180°+) look wrong due to projection distortion

## Goals

**Primary Goal:** Enable accurate, immersive dome/bubble rendering with correct spherical projection and real-time SR/GR effects.

**Success Metrics:**
- Dome renders 180°+ field of view without projection distortion
- SR aberration/Doppler correctly applied across curved surface
- 60 FPS maintained with full dome visible
- Crew can stand anywhere in dome room and see accurate star positions

## Physics Basis

### Spherical Projection (Real Physics)

A dome window requires **equirectangular** or **cubemap** projection to accurately represent what a viewer would see in all directions.

| Projection | Description | Use Case |
|------------|-------------|----------|
| Rectilinear | Standard perspective | Small windows (<90° FOV) |
| Equirectangular | Lat/long mapping on sphere | Dome interiors, planetariums |
| Cubemap | 6 faces of a cube | Real-time rendering, skyboxes |

**For our dome:** Use cubemap for real-time rendering, sample from it based on viewing angle.

### SR Effects on Dome

At relativistic velocities, stars appear to bunch forward (aberration). On a dome:
- Front of dome: Stars compressed, blue-shifted
- Sides of dome: Stars stretched, color neutral
- Rear of dome: Stars sparse, red-shifted

This is physically accurate and visually striking.

### GR Effects on Dome

Near massive objects, the entire dome view is lensed toward the mass. This requires:
- Per-fragment ray tracing through curved spacetime
- Or: Pre-distorted cubemap with lensing baked in

## Solution Design

### Overview

Implement a `Dome3D` DrawCmd and corresponding Go renderer that:
1. Renders a 6-face cubemap of the starfield
2. Projects it onto a hemisphere/sphere mesh
3. Applies SR/GR effects in the shader

### Architecture

**Components:**

1. **AILANG `Dome3D` DrawCmd** - Defines dome geometry, position, SR/GR parameters
2. **Go Dome Renderer** - Creates hemisphere mesh, manages cubemap textures
3. **Cubemap Star Renderer** - Extends SpaceView to render 6 faces
4. **Dome Shader** - Samples cubemap with SR/GR corrections

### AILANG Type Definition

```ailang
-- Dome3D: Render a spherical/hemispherical observation dome
-- x, y, z: dome center position in room coordinates
-- radius: dome radius in meters
-- arcAngle: coverage angle (π = hemisphere, 2π = full sphere)
-- openingAngle: cutoff angle for floor/ceiling opening
-- fov: base field of view (affects sampling density)
-- showStars, showPlanets: what to render through dome
-- velocity: ship velocity for SR effects (0-1, fraction of c)
-- grPhi: gravitational potential for GR effects
| Dome3D(
    id: string,
    x: float, y: float, z: float,
    radius: float,
    arcAngle: float,      -- π = hemisphere (typical), 2π = bubble
    openingAngle: float,  -- Floor cutoff (π/6 = 30° from bottom)
    showStars: bool, showPlanets: bool,
    velocity: float, grPhi: float,
    domeZ: int
)
```

### Implementation Plan

**Phase 1: Cubemap Infrastructure** (~4 hours)
- [ ] Add `Dome3D` to `sim/protocol.ail`
- [ ] Extend `SpaceView` to render 6-face cubemap
- [ ] Create `engine/tetra/dome.go` with hemisphere mesh generation
- [ ] Basic dome rendering without SR/GR

**Phase 2: Dome Shader** (~4 hours)
- [ ] Create `engine/shader/shaders/dome_sample.kage`
- [ ] Implement cubemap sampling from spherical coordinates
- [ ] Add SR aberration (star bunching) to shader
- [ ] Add SR Doppler (color shift) to shader

**Phase 3: GR Integration** (~3 hours)
- [ ] Add GR lensing to dome shader (ray deflection)
- [ ] Add GR redshift to dome shader
- [ ] Handle edge cases (photon sphere, event horizon)

**Phase 4: Demo & Polish** (~2 hours)
- [ ] Create `cmd/demo-dome/main.go`
- [ ] Performance optimization (LOD for distant stars)
- [ ] Edge blending at dome boundaries

### Files to Modify/Create

**New files:**
- `engine/tetra/dome.go` - Hemisphere mesh generation (~150 LOC)
- `engine/shader/shaders/dome_sample.kage` - Cubemap sampling shader (~100 LOC)
- `engine/render/draw_dome.go` - Dome rendering logic (~200 LOC)
- `cmd/demo-dome/main.go` - Demo/test harness (~150 LOC)

**Modified files:**
- `sim/protocol.ail` - Add Dome3D DrawCmd (~15 LOC)
- `engine/render/draw.go` - Add Dome3D case to switch (~10 LOC)
- `engine/render/space_view.go` - Add cubemap rendering (~50 LOC)

## Examples

### Example 1: Bridge Observation Dome

**AILANG:**
```ailang
-- Render a hemisphere dome on the bridge ceiling
pure func renderBridgeDome(nav: ShipNavigation) -> DrawCmd {
    Dome3D(
        "bridge_dome",
        0.0, 3.0, -2.0,    -- Center above bridge
        4.0,                -- 4 meter radius
        3.14159,            -- π = hemisphere
        0.52,               -- 30° floor cutoff
        true, true,         -- Show stars and planets
        nav.velocity, nav.grPhi,
        100                 -- Z-order
    )
}
```

**Visual result:** A 4-meter hemisphere showing half the sky, with SR/GR effects based on ship state.

### Example 2: Full Bubble Ship

**AILANG:**
```ailang
-- Render full transparent bubble (360° view)
pure func renderBubbleShip(nav: ShipNavigation) -> DrawCmd {
    Dome3D(
        "ship_bubble",
        0.0, 0.0, 0.0,      -- Ship center
        10.0,               -- 10 meter radius
        6.28318,            -- 2π = full sphere
        0.0,                -- No floor cutoff
        true, true,
        nav.velocity, nav.grPhi,
        100
    )
}
```

### Example 3: Observation Deck Window Dome

```ailang
-- Curved window in observation lounge (quarter sphere)
pure func renderObservationWindow(pos: Vec3, nav: ShipNavigation) -> DrawCmd {
    Dome3D(
        "obs_window",
        pos.x, pos.y, pos.z,
        2.0,                -- 2 meter radius
        1.57,               -- π/2 = quarter sphere
        0.0,
        true, false,        -- Stars only, no planets
        nav.velocity, nav.grPhi,
        100
    )
}
```

## Dome Mesh Geometry

The dome mesh is a tessellated hemisphere:

```
        *  *  *  *  *
      *   *   *   *   *
    *   *   *   *   *   *     <- Top ring (small triangles)
   *  *  *  *  *  *  *  *
  *  *  *  *  *  *  *  *  *   <- Middle rings
 *  *  *  *  *  *  *  *  *  *
*  *  *  *  *  *  *  *  *  *  * <- Equator (where it meets floor)
```

**Parameters:**
- `rings`: Number of latitude rings (16 for smooth curve)
- `segments`: Number of longitude segments (32 for smooth circle)
- `arcAngle`: How much of sphere to render (π = half, 2π = full)

## Success Criteria

- [ ] Dome renders 180° FOV without visible projection distortion
- [ ] Stars maintain correct relative positions across dome surface
- [ ] SR aberration visibly bunches stars toward front at v > 0.5c
- [ ] SR Doppler shifts colors correctly (blue forward, red backward)
- [ ] GR lensing distorts dome view when grPhi > 0.01
- [ ] 60 FPS maintained with 10,000+ visible stars
- [ ] Demo runs showing dome from inside ship interior
- [ ] All tests passing
- [ ] Documentation updated

## Testing Strategy

**Unit tests:**
- Hemisphere mesh vertex generation
- Cubemap face selection
- SR aberration angle calculation

**Visual tests:**
- `cmd/demo-dome` with known star positions
- Compare against reference images at various velocities
- Verify no seams at cubemap face boundaries

**Performance tests:**
- FPS with 10K, 50K, 100K stars visible
- Memory usage for cubemap textures

## Non-Goals

**Not in this feature:**
- Interior reflections on dome surface - Deferred (requires separate pass)
- Multiple domes in same room - Out of scope (edge case)
- Animated dome opening/closing - Later feature
- Weather/particle effects on dome - Different system

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Cubemap seams visible | Med | Use seamless cubemap sampling, overlap at edges |
| Performance with full 360° | High | LOD system - fewer stars at edges, more at center |
| SR/GR math complexity | Med | Pre-compute lookup tables for common angles |
| Tetra3D hemisphere limits | Low | Fall back to multi-panel flat approximation |

## References

- [Window3D implementation](../../engine/render/draw_interior.go) - Current flat window system
- [SpaceView](../../engine/render/space_view.go) - Star catalog rendering
- [SR effects design](implemented/v0_1_0/sr-effects.md) - Special relativity implementation
- [GR effects design](implemented/v0_1_0/gr-effects.md) - General relativity implementation
- [Equirectangular projection](https://en.wikipedia.org/wiki/Equirectangular_projection) - Math reference

## Future Work

- **Interior reflections** - Stars reflecting on floor, walls
- **Dome transitions** - Animate dome opening for docking
- **Multiple domes** - Different rooms with different dome configurations
- **Dome HUD overlay** - Navigation markers, target indicators on dome surface
- **Crew silhouettes** - Show crew members against starfield backdrop

---

**Document created**: 2025-12-19
**Last updated**: 2025-12-19
