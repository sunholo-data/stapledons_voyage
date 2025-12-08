# Planetary Ring Systems

**Status**: Planned
**Target**: v0.2.0
**Priority**: P2 (Low - visual polish)
**Estimated**: 1-2 days
**Dependencies**: 3D Sphere Planet Rendering (03-3d-sphere-planets.md)

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Visual feature |
| Civilization Simulation | N/A | 0 | Visual feature |
| Philosophical Depth | N/A | 0 | Visual feature |
| Ship & Crew Life | N/A | 0 | Visual feature |
| Legacy Impact | N/A | 0 | Visual feature |
| Hard Sci-Fi Authenticity | + | +1 | All gas giants have rings in reality |
| **Net Score** | | **+1** | **Move forward** |

**Feature type:** Engine/Visual
- Infrastructure/rendering feature - N/A on most pillars is acceptable
- Positive score on Hard Sci-Fi Authenticity makes this worthwhile

## Problem Statement

Current planet rendering (when implemented) only includes Saturn's rings. In reality, **all four gas giants have ring systems**:

| Planet | Ring Visibility | Composition |
|--------|-----------------|-------------|
| **Saturn** | Spectacular, easily visible | Ice particles (mostly water ice) |
| **Jupiter** | Faint, dusty | Dust from moon impacts |
| **Uranus** | Dark, narrow | Dark carbonaceous particles |
| **Neptune** | Faint, clumpy arcs | Dust and ice |

**Current State:**
- Only Saturn is planned with rings in `03-3d-sphere-planets.md`
- Other gas giants would appear unrealistic without rings

**Impact:**
- Astronomical accuracy for hard sci-fi pillar
- Visual distinction between gas giants
- Educational value for players

## Scientific Reference

### Ring Prevalence

**In our solar system:**
- 100% of gas giants (4/4) have rings
- 0% of rocky planets (0/4) have rings
- Some dwarf planets have rings (Haumea)
- Some asteroids have rings (Chariklo, Chiron)

**Why gas giants have rings:**
1. **Strong gravity** - Can capture and retain debris
2. **Distance from Sun** - Ice doesn't sublimate
3. **Many moons** - Collisions produce debris
4. **Roche limit** - Material inside can't coalesce into moons

### Ring Properties by Planet

| Planet | Inner Radius | Outer Radius | Thickness | Opacity |
|--------|--------------|--------------|-----------|---------|
| Saturn | 1.11 R | 2.27 R | 10-100m | High |
| Jupiter | 1.29 R | 1.81 R | 30-300km | Very low |
| Uranus | 1.49 R | 1.95 R | km scale | Low |
| Neptune | 1.69 R | 2.54 R | km scale | Very low |

*R = planet radius*

### Visual Characteristics

| Planet | Color | Brightness | Distinctive Feature |
|--------|-------|------------|---------------------|
| Saturn | White/tan | Very bright | Wide, visible gaps (Cassini Division) |
| Jupiter | Rusty red | Very faint | Halo ring glows |
| Uranus | Gray/dark | Faint | Narrow, discrete rings |
| Neptune | Reddish | Very faint | Ring arcs (clumpy) |

## Solution Design

### Overview

Extend the `PlanetConfig` ring system to support all gas giants with accurate ring parameters.

### Ring Configuration

```go
// engine/tetra/planet.go

type RingConfig struct {
    InnerRadius  float64  // Multiple of planet radius
    OuterRadius  float64  // Multiple of planet radius
    Opacity      float64  // 0.0-1.0 (Saturn ~0.8, Jupiter ~0.05)
    Tilt         float64  // Degrees (matches planet axial tilt)
    TexturePath  string   // Ring texture with alpha
    HasGaps      bool     // Saturn's Cassini Division
    NumRings     int      // For Uranus-style discrete rings
}

var GasGiantRings = map[string]RingConfig{
    "saturn": {
        InnerRadius:  1.11,
        OuterRadius:  2.27,
        Opacity:      0.85,
        TexturePath:  "assets/planets/saturn_rings.png",
        HasGaps:      true,   // Cassini Division visible
        NumRings:     1,      // Continuous with gaps
    },
    "jupiter": {
        InnerRadius:  1.29,
        OuterRadius:  1.81,
        Opacity:      0.08,   // Very faint
        TexturePath:  "assets/planets/jupiter_rings.png",
        HasGaps:      false,
        NumRings:     1,
    },
    "uranus": {
        InnerRadius:  1.49,
        OuterRadius:  1.95,
        Opacity:      0.25,
        TexturePath:  "assets/planets/uranus_rings.png",
        HasGaps:      false,
        NumRings:     13,     // Discrete narrow rings
    },
    "neptune": {
        InnerRadius:  1.69,
        OuterRadius:  2.54,
        Opacity:      0.10,
        TexturePath:  "assets/planets/neptune_rings.png",
        HasGaps:      false,
        NumRings:     5,      // Named rings: Galle, Le Verrier, Lassell, Arago, Adams
    },
}
```

### Ring Rendering Approaches

| Ring Type | Planets | Approach |
|-----------|---------|----------|
| **Continuous** | Saturn, Jupiter | Single textured disk with alpha |
| **Discrete** | Uranus, Neptune | Multiple thin torus meshes or procedural |

### Implementation Plan

**Phase 1: Update Config** (~2 hours)
- [ ] Add `RingConfig` struct to `engine/tetra/planet.go`
- [ ] Update `PlanetConfig` to embed `RingConfig`
- [ ] Add ring configs for all four gas giants

**Phase 2: Ring Textures** (~4 hours)
- [ ] Create/source Saturn ring texture (already planned)
- [ ] Create subtle Jupiter ring texture (very faint)
- [ ] Create Uranus discrete ring texture
- [ ] Create Neptune ring arc texture

**Phase 3: Rendering** (~4 hours)
- [ ] Update `createRings()` to handle opacity
- [ ] Add discrete ring rendering for Uranus-style
- [ ] Test with all four planets

### Files to Modify/Create

**Modified files:**
- `engine/tetra/planet.go` - Add ring configs, update rendering

**New files:**
- `assets/planets/jupiter_rings.png` - Faint dusty ring texture
- `assets/planets/uranus_rings.png` - Discrete narrow rings
- `assets/planets/neptune_rings.png` - Ring arcs

## Visual Examples

### Saturn (High Visibility)
```
        ╭─────────────────────╮
     ───│     ╭───────╮       │───
    ────│    ╱         ╲      │────
   ─────│   ╱  SATURN   ╲     │─────
    ────│   ╲           ╱     │────
     ───│    ╲╭───────╮╱      │───
        ╰─────────────────────╯
```
Bright white/tan, Cassini Division gap visible

### Jupiter (Low Visibility)
```
           . . . . . . .
          .             .
         .   JUPITER    .
          .             .
           . . . . . . .
```
Faint rusty halo, only visible against dark background

### Uranus (Discrete Rings)
```
        │ │ │       │ │ │
        │ │ │ URANUS│ │ │
        │ │ │       │ │ │
```
Narrow, separated dark rings perpendicular to orbit

## Success Criteria

- [ ] Saturn renders with prominent rings (Cassini Division visible)
- [ ] Jupiter has subtle ring system (faint but present)
- [ ] Uranus has discrete narrow rings
- [ ] Neptune has faint ring arcs
- [ ] Rings tilt correctly with planet axial tilt
- [ ] Rings cast shadow on planet (stretch goal)
- [ ] Rings visible through SR Doppler shader

## Testing Strategy

**Visual tests:**
- `./bin/game --demo-planet saturn` - Rings clearly visible
- `./bin/game --demo-planet jupiter` - Rings barely visible (correct)
- `./bin/game --demo-planet uranus` - Discrete rings visible
- `./bin/game --demo-planet neptune` - Faint arcs visible

**Shader integration:**
- Rings should blue-shift when approaching at 0.3c

## Non-Goals

**Not in this feature:**
- Ring shadows on planet surface - Complex lighting, defer
- Ring particle simulation - Too computationally expensive
- Ring dynamics over time - Static is fine for game timescales

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Faint rings invisible | Med | Test on multiple monitors, add brightness option |
| Performance with multiple rings | Low | LOD: distant planets skip rings |
| Texture sourcing | Low | Procedural fallback for non-Saturn |

## References

- [03-3d-sphere-planets.md](../next/03-3d-sphere-planets.md) - Parent design doc
- [NASA Ring Systems](https://solarsystem.nasa.gov/planets/overview/) - Scientific reference
- [Cassini Saturn Ring Images](https://photojournal.jpl.nasa.gov/catalog/PIA08389) - Saturn ring textures

## Future Work

- Ring shadows on planet surfaces
- Shepherd moons visible near rings
- Exoplanet ring systems (J1407b has rings 200x Saturn's)
- Ring particle effects during close approach

---

**Document created**: 2025-12-08
**Last updated**: 2025-12-08
