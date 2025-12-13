# AILANG-Driven Planet Ring and Moon Data

## Status
- **Status**: Planned
- **Priority**: P2 (architecture compliance)
- **Estimated**: 1 day
- **Location**: `sim/celestial.ail`, `cmd/demo-game-saturn/main.go`

## Game Vision Alignment

| Pillar | Alignment | Notes |
|--------|-----------|-------|
| Time Dilation Consequence | N/A | Data architecture feature |
| Civilization Simulation | N/A | Rendering data, not simulation |
| Ship & Crew Life | N/A | Infrastructure feature |
| Hard Sci-Fi Authenticity | Supports | Accurate planet data enables realistic rendering |
| **Overall** | Enabler | Proper separation of data (AILANG) from rendering (Go) |

**This is an architecture compliance feature** - per CLAUDE.md, all game data should be defined in AILANG, with the Go engine as a "dumb renderer."

## Problem Statement

Currently, planet ring and moon data is **hardcoded in Go**:

```go
// cmd/demo-game-saturn/main.go - WRONG: game data in Go
ringBands := tetra.SaturnRingBands(saturnRadius)  // Hardcoded in engine
moons := g.createMoons(saturnRadius)               // Hardcoded moon definitions
```

This violates the AILANG-first architecture. The Go engine should only:
1. Read data from AILANG
2. Convert to rendering types
3. Display on screen

## Proposed Solution

### 1. New AILANG Types (`sim/celestial.ail`)

```ailang
-- Ring band definition for multi-band ring systems
export type RingBand = {
    innerRadius: float,  -- Planet radii multiplier (1.0 = planet surface)
    outerRadius: float,  -- Planet radii multiplier
    color: Color,        -- Band color (uses existing Color type)
    opacity: float,      -- 0.0-1.0 (transparency)
    density: float       -- 0.0-1.0 (affects dust clumping)
}

-- Moon definition
export type MoonDef = {
    name: string,
    radius: float,       -- Planet radii multiplier
    color: Color,        -- Surface color
    orbitRadius: float,  -- Distance from planet (planet radii)
    orbitSpeed: float,   -- Radians per second (visual, not realistic)
    orbitTilt: float     -- Inclination above ring plane (radians)
}
```

### 2. Extended CelestialPlanet Type

Add to existing `CelestialPlanet` record:

```ailang
export type CelestialPlanet = {
    -- ... existing fields ...
    axialTilt: float,        -- Degrees (affects ring plane orientation)
    ringBands: [RingBand],   -- Detailed ring bands (empty if no rings)
    moons: [MoonDef]         -- Moons orbiting this planet
}
```

### 3. Planet-Specific Data Functions

```ailang
-- Saturn's ring bands (scientifically accurate)
export pure func saturnRingBands() -> [RingBand] {
    [
        -- C Ring (inner, dim)
        { innerRadius: 1.24, outerRadius: 1.53,
          color: { r: 180, g: 160, b: 130, a: 255 },
          opacity: 0.3, density: 0.4 },
        -- B Ring (main, brightest)
        { innerRadius: 1.53, outerRadius: 1.95,
          color: { r: 220, g: 205, b: 170, a: 255 },
          opacity: 0.7, density: 0.9 },
        -- A Ring (outer)
        { innerRadius: 2.03, outerRadius: 2.27,
          color: { r: 210, g: 190, b: 150, a: 255 },
          opacity: 0.5, density: 0.7 }
    ]
}

-- Saturn's major moons
export pure func saturnMoons() -> [MoonDef] {
    [
        { name: "Titan", radius: 0.4,
          color: { r: 210, g: 160, b: 100, a: 255 },
          orbitRadius: 4.0, orbitSpeed: 0.15, orbitTilt: 0.1 },
        { name: "Enceladus", radius: 0.15,
          color: { r: 240, g: 245, b: 255, a: 255 },
          orbitRadius: 2.8, orbitSpeed: 0.4, orbitTilt: 0.0 },
        { name: "Mimas", radius: 0.12,
          color: { r: 180, g: 180, b: 190, a: 255 },
          orbitRadius: 2.5, orbitSpeed: 0.5, orbitTilt: -0.05 }
    ]
}

-- Uranus-style rings (narrow, dark)
export pure func uranusRingBands() -> [RingBand] {
    [
        { innerRadius: 1.6, outerRadius: 1.65,
          color: { r: 60, g: 60, b: 70, a: 255 },
          opacity: 0.4, density: 0.3 },
        { innerRadius: 1.9, outerRadius: 1.95,
          color: { r: 50, g: 50, b: 60, a: 255 },
          opacity: 0.3, density: 0.2 }
    ]
}
```

### 4. Go Engine Conversion

The Saturn demo reads from AILANG and converts to engine types:

```go
// Convert AILANG RingBand to tetra.RingBand
func convertRingBands(ailangBands []*sim_gen.RingBand, planetRadius float64) []tetra.RingBand {
    bands := make([]tetra.RingBand, len(ailangBands))
    for i, b := range ailangBands {
        bands[i] = tetra.RingBand{
            InnerRadius: b.InnerRadius * planetRadius,
            OuterRadius: b.OuterRadius * planetRadius,
            Color:       color.RGBA{uint8(b.Color.R), uint8(b.Color.G), uint8(b.Color.B), 255},
            Opacity:     b.Opacity,
            Density:     b.Density,
        }
    }
    return bands
}

// Convert AILANG MoonDef to engine Moon
func convertMoons(ailangMoons []*sim_gen.MoonDef, planetRadius float64) []*Moon {
    // ... similar conversion ...
}
```

## Files to Modify

| File | Changes |
|------|---------|
| `sim/celestial.ail` | Add RingBand, MoonDef types; extend CelestialPlanet; add data functions |
| `sim_gen/*.go` | Regenerate via `make sim` |
| `cmd/demo-game-saturn/main.go` | Replace hardcoded data with AILANG calls |
| `engine/tetra/ring.go` | Remove `SaturnRingBands()` preset (now in AILANG) |

## Migration Strategy

1. **Phase 1**: Add new types to AILANG without breaking existing code
2. **Phase 2**: Add data functions (saturnRingBands, saturnMoons, etc.)
3. **Phase 3**: Update Saturn demo to read from AILANG
4. **Phase 4**: Remove Go presets (SaturnRingBands in ring.go)

## Success Criteria

- [ ] RingBand and MoonDef types defined in AILANG
- [ ] CelestialPlanet includes ringBands and moons fields
- [ ] Saturn ring/moon data defined in AILANG functions
- [ ] Uranus ring data defined as alternative preset
- [ ] Saturn demo reads ring/moon data from AILANG
- [ ] Go engine has no planet-specific data (only rendering code)
- [ ] `make sim && go build ./...` succeeds
- [ ] Saturn demo produces same visual output

## AILANG Considerations

- Lists are the only collection type (no arrays)
- Records can be nested (Color inside RingBand)
- Functions should be pure (no effects needed for static data)
- Consider list recursion limits for large moon systems

## Future Extensions

This architecture enables:
- Procedural ring generation in AILANG
- Planet-specific data loaded from external sources
- User-customizable planet configurations
- Different ring styles for fictional planets

## References

- [engine-capabilities.md](../reference/engine-capabilities.md) - Ring rendering docs
- [CLAUDE.md](../../CLAUDE.md) - AILANG-first architecture rules
- `sim/celestial.ail` - Existing planet types
