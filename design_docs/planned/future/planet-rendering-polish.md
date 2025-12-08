# Planet Rendering Polish: Atmosphere Glow & LOD System

**Status**: Planned
**Target**: v0.3.0
**Priority**: P2 - Low (visual polish, not blocking gameplay)
**Estimated**: 2-3 days
**Dependencies**: Tetra3D Integration (implemented), 3D Sphere Planets (mostly implemented)

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Visual polish, no gameplay impact |
| Civilization Simulation | N/A | 0 | Visual polish |
| Philosophical Depth | N/A | 0 | Visual polish |
| Ship & Crew Life | N/A | 0 | Visual polish |
| Legacy Impact | N/A | 0 | Visual polish |
| Hard Sci-Fi Authenticity | + | +1 | Fresnel effect is real physics (Rayleigh scattering) |
| **Net Score** | | **+1** | **Decision: Move forward when time permits** |

**Feature type:** Engine/Infrastructure
- This is enabling tech for visual quality improvements
- N/A scores are acceptable for infrastructure features

**Reference:** See [game-vision.md](../../../docs/game-vision.md)

## Problem Statement

The current 3D planet rendering system works but lacks two polish features deferred from the initial implementation:

**Current State:**
- Planets render as solid spheres with textures
- No atmospheric glow at the limb (edge) of planets like Earth
- All planets render at full quality regardless of distance
- Performance degrades with multiple visible planets

**Impact:**
- Earth arrival lacks the iconic "blue marble" atmosphere glow
- Distant planets waste rendering resources
- Solar system views may stutter with many planets

## Physics Basis

### Fresnel Effect (Atmosphere Glow)

**Physics:** When viewing a planet with atmosphere, light scatters more at grazing angles (limb) than direct angles (center). This is caused by:

1. **Rayleigh Scattering**: Atmosphere scatters blue light more than red
2. **Path Length**: Light at limb travels through more atmosphere
3. **Fresnel Equations**: Reflection increases at grazing angles

**Visual Result**: Thin blue glow around Earth's edge, fading inward

**Equation (simplified Fresnel approximation):**
```
fresnel = pow(1.0 - dot(viewDir, normal), fresnelPower)
```

Where `fresnelPower` ≈ 3-5 for Earth-like atmospheres.

### LOD (Level of Detail)

**Physics Basis:** Angular size determines perceivable detail. A planet 100 units away appears smaller than one 10 units away - rendering high detail at 100 units is wasted.

**Principle:** Match polygon count to screen-space pixel coverage.

## Goals

**Primary Goal:** Add visual polish to planet rendering with minimal performance cost.

**Success Metrics:**
- Earth shows visible blue atmosphere glow at limb
- Planets at distance >50 units render as low-poly spheres
- Planets at distance >200 units render as 2D sprites
- FPS remains ≥20 with 8 visible planets

## Solution Design

### Overview

Two independent features that can be implemented separately:

1. **Fresnel Atmosphere Shader** - Post-process glow at planet edges
2. **Distance-Based LOD** - Swap planet detail levels based on camera distance

### Architecture

#### Feature 1: Fresnel Atmosphere Glow

**Approach A: Outer Sphere (Recommended)**
- Create slightly larger transparent sphere around planet
- Apply Fresnel shader to outer sphere
- Faster, works with existing Tetra3D setup

**Approach B: Post-Process Shader**
- Render planet to depth buffer
- Post-process to detect edges and add glow
- More complex, requires depth buffer access

**Components:**
1. `engine/tetra/atmosphere.go` - Atmosphere sphere mesh + material
2. `engine/shader/fresnel.kage` - Fresnel shader (if Ebitengine shader needed)
3. Extension to `Planet.AddAtmosphere()` method

#### Feature 2: LOD System

**LOD Levels:**

| Distance | Rendering Mode | Polys |
|----------|---------------|-------|
| <10 units | Full 3D + atmosphere | ~5000 |
| 10-50 units | Full 3D, no atmosphere | ~5000 |
| 50-200 units | Low-poly 3D | ~500 |
| >200 units | 2D sprite | ~2 |

**Components:**
1. `engine/tetra/lod.go` - LOD manager and distance calculations
2. Low-poly sphere meshes (1-2 subdivisions vs 3-4)
3. Pre-rendered planet sprites for far distance

### Implementation Plan

**Phase 1: Fresnel Atmosphere** (~4 hours)
- [ ] Create `atmosphere.go` with outer sphere generation
- [ ] Implement Fresnel material using Tetra3D transparency
- [ ] Add `Planet.AddAtmosphere(color, thickness)` method
- [ ] Test on Earth with blue atmosphere
- [ ] Screenshot verification

**Phase 2: LOD Infrastructure** (~4 hours)
- [ ] Create `lod.go` with LOD level enum
- [ ] Add distance-based mesh swapping to Planet
- [ ] Generate low-poly mesh variants
- [ ] Test LOD transitions

**Phase 3: Sprite Fallback** (~4 hours)
- [ ] Pre-render planet sprites at load time
- [ ] Implement sprite swap at far distances
- [ ] Test with solar system view (8 planets)
- [ ] Performance benchmarking

### Files to Modify/Create

**New files:**
- `engine/tetra/atmosphere.go` - Atmosphere glow (~80 LOC)
- `engine/tetra/lod.go` - LOD system (~120 LOC)
- `engine/shader/fresnel.kage` - Fresnel shader if needed (~30 LOC)

**Modified files:**
- `engine/tetra/planet.go` - Add atmosphere and LOD methods (~50 LOC)
- `engine/tetra/scene.go` - LOD distance calculations (~20 LOC)

## Examples

### Example 1: Adding Atmosphere to Earth

**Before:**
```go
earth := tetra.NewTexturedPlanet("earth", 1.0, earthTexture)
```

**After:**
```go
earth := tetra.NewTexturedPlanet("earth", 1.0, earthTexture)
earth.AddAtmosphere(color.RGBA{100, 150, 255, 128}, 0.02)  // Blue glow, 2% thickness
```

### Example 2: LOD Configuration

```go
// Configure LOD distances
planet.SetLODDistances(
    10.0,   // Full → Standard at 10 units
    50.0,   // Standard → Low at 50 units
    200.0,  // Low → Sprite at 200 units
)

// LOD updates automatically in scene.Update()
```

## Success Criteria

- [ ] Earth displays blue atmosphere glow at limb when close
- [ ] Atmosphere glow fades naturally toward center
- [ ] LOD transitions are not jarring (smooth or instant)
- [ ] FPS ≥20 with 8 planets visible
- [ ] FPS ≥30 with 1 planet visible (current ~21)
- [ ] Screenshot tests verify atmosphere appearance
- [ ] Demo command: `./bin/demo-atmosphere`

## Testing Strategy

**Unit tests:**
- LOD level selection for various distances
- Fresnel calculation accuracy

**Integration tests:**
- Screenshot comparison with/without atmosphere
- Performance benchmarks at various planet counts

**Manual testing:**
- Visual inspection of Earth atmosphere glow
- LOD transition smoothness during camera motion
- Solar system view with all planets

## Non-Goals

**Not in this feature:**
- Night-side city lights - Requires separate texture layer, deferred
- Cloud layers - Would need animated textures, deferred
- Atmospheric refraction - Complex shader, out of scope
- Per-planet atmosphere colors - Could add later, keep simple for MVP

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Tetra3D transparency issues | Med | Test early, have Ebitengine shader fallback |
| LOD popping visible | Low | Use distance hysteresis, gradual fade |
| Sprite rendering different from 3D | Low | Pre-render at consistent lighting |
| Performance worse than expected | Med | Profile early, simplify if needed |

## References

- [03-3d-sphere-planets.md](../../implemented/v0_2_0/03-3d-sphere-planets.md) - Parent feature (implemented)
- [02-tetra3d-integration.md](../../implemented/v0_2_0/02-tetra3d-integration.md) - Tetra3D setup (implemented)
- [Fresnel equations (Wikipedia)](https://en.wikipedia.org/wiki/Fresnel_equations)
- [Rayleigh scattering (Wikipedia)](https://en.wikipedia.org/wiki/Rayleigh_scattering)

## Future Work

Features that build on this but are out of scope:
- **Animated clouds** - Separate moving cloud texture layer
- **Aurora effects** - Polar light displays for planets with magnetic fields
- **Ring shadow casting** - Saturn's rings casting shadow on planet surface
- **Atmospheric entry effects** - Heat glow when entering atmosphere

---

**Document created**: 2025-12-08
**Last updated**: 2025-12-08
**Deferred from**: [03-3d-sphere-planets.md](../../implemented/v0_2_0/03-3d-sphere-planets.md) Sprint Progress
