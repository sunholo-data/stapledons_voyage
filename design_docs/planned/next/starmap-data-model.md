# Starmap Data Model

**Status**: Planned
**Target**: v0.3.0
**Priority**: P1 - High
**Estimated**: 1 week
**Dependencies**: World Gen Settings

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | + | +1 | Distance determines time cost - core to journey planning |
| Civilization Simulation | + | +1 | Where civs exist, how far apart, affects all interactions |
| Philosophical Depth | 0 | 0 | Infrastructure - enables but doesn't directly add philosophy |
| Ship & Crew Life | 0 | 0 | Infrastructure |
| Legacy Impact | + | +1 | Galaxy structure shapes what you can affect in 100 years |
| Hard Sci-Fi Authenticity | + | +1 | Real Gaia data + procedural grounded in astrophysics |
| **Net Score** | | **+4** | **Decision: Move forward** |

**Feature type:** Infrastructure (enables core gameplay)

## Problem Statement

The game needs a 3D star map that:
- Feels scientifically authentic (real stars where possible)
- Supports gameplay scale (thousands of light-years reachable)
- Enables efficient queries (nearest neighbors, distance calculations)
- Integrates with civilization placement via Anthropic Luck factor

**Current State:**
- No star map exists yet
- Need to balance realism vs performance vs gameplay density

**Impact:**
- Foundation for all navigation, journey planning, and civ simulation
- Determines what the galaxy "feels like" to explore

## Goals

**Primary Goal:** Create a star map data model that combines real Gaia DR3 data locally with procedural generation at galactic scales.

**Success Metrics:**
- Real stars within ~300 parsecs (~1000 ly) from Gaia catalog
- Procedural stars beyond that follow realistic density distributions
- Query "stars within N light-years" returns in <100ms for N<5000
- Civ placement respects Anthropic Luck parameter

## Solution Design

### Overview

Three-layer star map:

1. **Local Bubble** (0-1000 ly): Real Gaia DR3 stars with accurate positions
2. **Mid Galaxy** (1000-10,000 ly): Procedurally generated following galactic structure
3. **Deep Galaxy** (10,000+ ly): Sparse procedural, mostly for background/extreme journeys

### Architecture

**Components:**

1. **StarCatalog**: Core data structure holding all stars
   ```ailang
   type Star = {
       id: int,
       pos: Vec3,           -- Galactocentric coords in light-years
       spectralType: SpectralType,
       luminosity: float,
       hasHZPlanet: bool,   -- Habitable zone planet exists
       planets: [Planet]
   }

   type SpectralType = O | B | A | F | G | K | M

   type Vec3 = { x: float, y: float, z: float }
   ```

2. **SpatialIndex**: Octree for fast spatial queries
   ```ailang
   type Octree =
       | Leaf([Star])
       | Node(Octree, Octree, Octree, Octree, Octree, Octree, Octree, Octree)
   ```

3. **GalaxyModel**: Density functions for procedural generation
   - Exponential disk falloff
   - Spiral arm enhancement
   - Central bulge
   - Halo (sparse)

### Data Sources

**Gaia DR3 Integration:**
- Use "Gaia Catalogue of Nearby Stars" (~300k stars within 100 pc)
- Filter to ~10k-50k most relevant (G/K/M dwarfs with HZ potential)
- Convert to Galactocentric Cartesian coordinates

**Procedural Beyond Local:**
- Stellar density: ρ(r,z) = ρ₀ × exp(-r/h_r) × exp(-|z|/h_z)
- h_r ≈ 10,000 ly (radial scale length)
- h_z ≈ 1,000 ly (vertical scale height)
- Spiral arms: density boost along logarithmic spiral pattern

### Star Properties Generation

For procedural stars:

```ailang
pure func generateStar(seed: int, pos: Vec3) -> Star {
    let spectral = rollSpectralType(seed)  -- M most common, O rarest
    let lum = luminosityForType(spectral)
    let hasHZ = rollHZPlanet(seed, spectral)  -- η⊕ varies by type
    { id: seed, pos: pos, spectralType: spectral,
      luminosity: lum, hasHZPlanet: hasHZ, planets: [] }
}
```

Spectral type distribution (realistic):
- M: 76%
- K: 12%
- G: 7.5%
- F: 3%
- A: 0.6%
- B: 0.13%
- O: 0.00003%

### Implementation Plan

**Phase 1: Core Data Structures** (~2 days)
- [ ] Define Star, Planet, Vec3 types in AILANG
- [ ] Implement distance calculations
- [ ] Create basic in-memory star list

**Phase 2: Gaia Import Pipeline** (~2 days)
- [ ] Download/process Gaia CSV subset
- [ ] Convert coordinates (RA/Dec/parallax → Galactocentric XYZ)
- [ ] Import into Go, expose to AILANG

**Phase 3: Procedural Generation** (~2 days)
- [ ] Implement galaxy density model
- [ ] Deterministic star generation from seed + position
- [ ] Blend real + procedural at boundary

**Phase 4: Spatial Indexing** (~1 day)
- [ ] Octree implementation for fast queries
- [ ] "Stars within radius" query
- [ ] "Nearest N stars" query

### Files to Modify/Create

**New files:**
- `sim/starmap.ail` - Star types and query functions (~200 LOC)
- `sim/galaxy_model.ail` - Density functions, procedural gen (~150 LOC)
- `engine/stardata/gaia_import.go` - Gaia CSV processing (~300 LOC)
- `engine/stardata/catalog.go` - Go-side star catalog (~200 LOC)

**Data files:**
- `assets/data/gaia_nearby.csv` - Processed Gaia subset (~5MB)

## Examples

### Example 1: Query Stars Near Sol

```ailang
let nearSol = starsWithinRadius(catalog, solPosition, 50.0)
-- Returns ~2000 stars within 50 ly of Sol
```

### Example 2: Find Nearest G-type with HZ Planet

```ailang
let candidates = filter(\s. s.spectralType == G && s.hasHZPlanet,
                        starsWithinRadius(catalog, currentPos, 500.0))
let nearest = minBy(\s. distance(currentPos, s.pos), candidates)
```

## Success Criteria

- [ ] Sol and nearest ~50 real stars match actual positions
- [ ] 10,000+ stars queryable within 1000 ly
- [ ] Procedural density visually matches spiral galaxy structure
- [ ] Distance queries return in <100ms
- [ ] Deterministic: same seed produces same galaxy

## Testing Strategy

**Unit tests:**
- Distance calculations correct
- Spectral type distribution matches expected percentages
- Coordinate conversions accurate

**Integration tests:**
- Gaia import produces valid star catalog
- Procedural + real blend smoothly at boundary
- Spatial queries return correct results

**Visual verification:**
- Render top-down galaxy view, verify spiral structure
- Render local bubble, verify Sol neighborhood recognizable

## Non-Goals

**Not in this feature:**
- Planet surface details - deferred to planet generation system
- Civilization placement - handled by world-gen-settings
- Star rendering/visuals - handled by engine layer
- Proper motion over time - optional flourish, not core

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Gaia data too large | Med | Pre-filter to relevant subset (~50k stars) |
| AILANG recursion limits on octree | High | Implement octree in Go, expose query API |
| Procedural/real boundary visible | Low | Smooth blending function over 100 ly transition |

## References

- [startmaps.md](startmaps.md) - Original design discussion
- [Gaia DR3 Documentation](https://www.cosmos.esa.int/web/gaia/dr3)
- [Gaia Catalogue of Nearby Stars](https://www.cosmos.esa.int/web/gaia/gcns)

## Future Work

- Proper motion for nearby stars over deep time (visual only)
- Binary/multiple star systems
- Stellar evolution (giants, white dwarfs, etc.)
- Nebulae and other non-stellar objects

---

**Document created**: 2025-12-02
**Last updated**: 2025-12-02
