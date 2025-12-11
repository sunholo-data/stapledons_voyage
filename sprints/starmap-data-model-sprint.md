# Sprint: Starmap Data Model

**Status:** Planned
**Duration:** 5-7 days
**Priority:** P0 (Foundation for Phase 2)
**Design Doc:** [starmap-data-model.md](../design_docs/planned/phase1-data-models/starmap-data-model.md)

## Goal

Create a star map data model combining real Gaia DR3 data with procedural generation at galactic scales. This enables galaxy map navigation and journey planning.

## Success Criteria

- [ ] Sol and nearest ~50 real stars match actual positions
- [ ] 10,000+ stars queryable within 1000 ly
- [ ] Procedural density matches spiral galaxy structure
- [ ] Distance queries return in <100ms
- [ ] Deterministic: same seed produces same galaxy

---

## Day 1: Core AILANG Types

### Task 1.1: Define Star Types
**File:** `sim/starmap.ail` (NEW)

```ailang
module sim/starmap

import std/prelude
import std/math (sqrt)

type StarID = StarID(int)
type SystemID = SystemID(int)

type Vec3 = { x: float, y: float, z: float }

type SpectralType = O | B | A | F | G | K | M

type Star = {
    id: StarID,
    name: string,
    pos: Vec3,
    spectral: SpectralType,
    luminosity: float,
    hasHZPlanet: bool
}
```

- [ ] Create `sim/starmap.ail` with Vec3, Star, SpectralType types
- [ ] Add distance calculation: `pure func distance(a: Vec3, b: Vec3) -> float`
- [ ] Run `ailang check sim/starmap.ail`

### Task 1.2: Define StarCatalog Type
**File:** `sim/starmap.ail`

```ailang
type StarCatalog = {
    stars: [Star],
    solIndex: int  -- Index of Sol in star list
}

export pure func emptyCatalog() -> StarCatalog {
    { stars: [], solIndex: 0 }
}

export pure func addStar(cat: StarCatalog, star: Star) -> StarCatalog {
    { cat | stars: star :: cat.stars }
}
```

- [ ] Add StarCatalog type with stars list
- [ ] Add emptyCatalog() and addStar() functions
- [ ] Run `ailang check`

### Task 1.3: Basic Queries
**File:** `sim/starmap.ail`

```ailang
-- Get stars within radius of a position
export pure func starsWithinRadius(cat: StarCatalog, center: Vec3, radius: float) -> [Star] {
    filter(\s. distance(s.pos, center) <= radius, cat.stars)
}

-- Find nearest star to position
export pure func nearestStar(cat: StarCatalog, pos: Vec3) -> Option(Star) {
    minByOpt(\s. distance(s.pos, pos), cat.stars)
}
```

- [ ] Add starsWithinRadius() function
- [ ] Add nearestStar() function
- [ ] Verify compilation

**Day 1 Test:**
```bash
ailang check sim/starmap.ail
make sim
go build ./...
```

---

## Day 2: Sol System & Local Stars

### Task 2.1: Sol System Definition
**File:** `sim/starmap.ail`

```ailang
-- Sol position (center of our coordinate system)
export pure func solPosition() -> Vec3 {
    { x: 0.0, y: 0.0, z: 0.0 }
}

-- Create Sol star entry
pure func makeSol() -> Star {
    {
        id: StarID(0),
        name: "Sol",
        pos: solPosition(),
        spectral: G,
        luminosity: 1.0,
        hasHZPlanet: true
    }
}
```

- [ ] Add solPosition() returning origin
- [ ] Add makeSol() for Sun entry
- [ ] Add Sol to default catalog

### Task 2.2: Nearby Real Stars (Hardcoded)
**File:** `sim/starmap.ail`

Add ~20 nearest real stars with accurate positions:

```ailang
-- Alpha Centauri (4.37 ly)
-- Barnard's Star (5.96 ly)
-- Wolf 359 (7.86 ly)
-- Lalande 21185 (8.29 ly)
-- Sirius (8.60 ly)
-- etc.
```

- [ ] Add localStars() returning [Star] with 20 nearest stars
- [ ] Verify positions match real data (Wikipedia/Gaia)
- [ ] Add initLocalCatalog() combining Sol + local stars

### Task 2.3: Spectral Type Distribution
**File:** `sim/starmap.ail`

```ailang
-- Spectral type distribution (realistic)
-- M: 76%, K: 12%, G: 7.5%, F: 3%, A: 0.6%, B: 0.13%, O: 0.00003%
pure func spectralFromRoll(roll: float) -> SpectralType {
    match roll < 0.76 {
        true => M,
        false => match roll < 0.88 {
            true => K,
            false => match roll < 0.955 {
                true => G,
                false => match roll < 0.985 {
                    true => F,
                    false => match roll < 0.991 {
                        true => A,
                        false => match roll < 0.99913 {
                            true => B,
                            false => O
                        }
                    }
                }
            }
        }
    }
}
```

- [ ] Add spectralFromRoll() for distribution
- [ ] Add luminosityForSpectral() helper
- [ ] Run tests

**Day 2 Test:**
```bash
ailang check sim/starmap.ail
# Should have 21+ stars in local catalog
```

---

## Day 3: Gaia Data Import (Go)

### Task 3.1: Gaia CSV Processing
**File:** `engine/stardata/gaia_import.go` (NEW)

```go
package stardata

type GaiaRecord struct {
    SourceID   int64
    RA         float64  // Right ascension (degrees)
    Dec        float64  // Declination (degrees)
    Parallax   float64  // Parallax (mas)
    Gmag       float64  // G-band magnitude
}

func LoadGaiaCatalog(path string) ([]GaiaRecord, error) {
    // Parse CSV, filter by parallax > 10 mas (within ~100 pc)
}
```

- [ ] Create `engine/stardata/` package
- [ ] Add GaiaRecord struct
- [ ] Add LoadGaiaCatalog() CSV parser

### Task 3.2: Coordinate Conversion
**File:** `engine/stardata/coords.go` (NEW)

```go
// Convert RA/Dec/Parallax to Galactocentric XYZ (light-years)
func ToGalactocentric(ra, dec, parallax float64) (x, y, z float64) {
    // parallax in mas → distance in parsecs
    distPC := 1000.0 / parallax
    distLY := distPC * 3.26156  // parsecs to light-years

    // Convert spherical to Cartesian (simplified, Sol-centered)
    raRad := ra * math.Pi / 180.0
    decRad := dec * math.Pi / 180.0

    x = distLY * math.Cos(decRad) * math.Cos(raRad)
    y = distLY * math.Cos(decRad) * math.Sin(raRad)
    z = distLY * math.Sin(decRad)
    return
}
```

- [ ] Add ToGalactocentric() conversion
- [ ] Test with known stars (Alpha Centauri should be ~4.37 ly)
- [ ] Handle edge cases (zero/negative parallax)

### Task 3.3: Bridge to AILANG
**File:** `engine/stardata/catalog.go` (NEW)

```go
// Convert Gaia records to sim_gen.Star slice
func ToAILANGStars(records []GaiaRecord) []*sim_gen.Star {
    // Filter, convert, return
}
```

- [ ] Create conversion to sim_gen.Star
- [ ] Add manifest/loading mechanism
- [ ] Test with small subset

**Day 3 Test:**
```bash
go test ./engine/stardata/...
```

---

## Day 4: Gaia Integration & Catalog Build

### Task 4.1: Download Gaia Subset
**Data:** `assets/data/starmap/gaia_nearby.csv`

- [ ] Download Gaia Catalogue of Nearby Stars subset
- [ ] Filter to ~10k-50k stars (G/K/M dwarfs with parallax > 10 mas)
- [ ] Save to assets/data/starmap/

### Task 4.2: Build Combined Catalog
**File:** `engine/stardata/builder.go` (NEW)

```go
func BuildCatalog(gaiaPath string) (*sim_gen.StarCatalog, error) {
    // 1. Load Gaia data
    // 2. Convert to AILANG stars
    // 3. Add Sol if not present
    // 4. Return catalog
}
```

- [ ] Create BuildCatalog() function
- [ ] Merge Gaia + manually added stars
- [ ] Verify Sol is at index 0

### Task 4.3: Catalog Initialization
**File:** `sim/starmap.ail` - Add extern declaration

```ailang
-- Extern: catalog loading happens in Go
extern func loadGaiaCatalog() -> StarCatalog
```

- [ ] Add extern func declaration for Go loading
- [ ] Wire up in game initialization
- [ ] Verify stars load correctly

**Day 4 Test:**
```bash
make run
# Verify stars appear in debug output
```

---

## Day 5: Procedural Generation

### Task 5.1: Galaxy Density Model
**File:** `sim/galaxy_model.ail` (NEW)

```ailang
module sim/galaxy_model

import std/prelude
import std/math (exp, sqrt)

-- Galactic parameters
pure func diskScaleLength() -> float { 10000.0 }  -- ly
pure func diskScaleHeight() -> float { 1000.0 }   -- ly

-- Stellar density at position (exponential disk)
export pure func stellarDensity(pos: Vec3) -> float {
    let r = sqrt(pos.x * pos.x + pos.y * pos.y);
    let z = pos.z;
    let hr = diskScaleLength();
    let hz = diskScaleHeight();
    exp(0.0 - r / hr) * exp(0.0 - abs(z) / hz)
}
```

- [ ] Create `sim/galaxy_model.ail`
- [ ] Add stellarDensity() exponential disk model
- [ ] Add spiral arm density boost (optional)

### Task 5.2: Deterministic Star Generation
**File:** `sim/galaxy_model.ail`

```ailang
-- Generate star at grid position (deterministic from seed)
export func generateStar(gridX: int, gridY: int, gridZ: int, baseSeed: int) -> Option(Star) ! {Rand} {
    let seed = hashGrid(gridX, gridY, gridZ, baseSeed);
    rand_seed(seed);

    let pos = gridToPosition(gridX, gridY, gridZ);
    let density = stellarDensity(pos);

    -- Roll for star existence based on density
    let roll = rand_float(0.0, 1.0);
    match roll < density {
        true => Some(makeProceduralStar(seed, pos)),
        false => None
    }
}
```

- [ ] Add grid-based generation
- [ ] Add hashGrid() for deterministic seeds
- [ ] Add makeProceduralStar() using spectralFromRoll

### Task 5.3: Boundary Blending
**File:** `sim/galaxy_model.ail`

```ailang
-- Blend real/procedural at boundary (~100 ly transition)
pure func blendWeight(distFromSol: float) -> float {
    let innerBound = 80.0;   -- Full real stars
    let outerBound = 120.0;  -- Full procedural
    match distFromSol < innerBound {
        true => 0.0,  -- Use real
        false => match distFromSol > outerBound {
            true => 1.0,  -- Use procedural
            false => (distFromSol - innerBound) / (outerBound - innerBound)
        }
    }
}
```

- [ ] Add blendWeight() function
- [ ] Implement blended query function
- [ ] Test transition zone

**Day 5 Test:**
```bash
ailang check sim/galaxy_model.ail
make sim && go build ./...
```

---

## Day 6: Spatial Indexing (Octree)

### Task 6.1: Go Octree Implementation
**File:** `engine/stardata/octree.go` (NEW)

```go
type Octree struct {
    bounds   AABB
    stars    []*sim_gen.Star  // Leaf data
    children [8]*Octree       // nil if leaf
}

func (o *Octree) Query(center Vec3, radius float64) []*sim_gen.Star {
    // Efficient spatial query
}
```

- [ ] Create Octree struct
- [ ] Add Build() constructor
- [ ] Add Query() for radius search

### Task 6.2: AILANG Extern Declaration
**File:** `sim/starmap.ail`

```ailang
-- Extern: Go octree for performance
extern func octreeQuery(center: Vec3, radius: float) -> [Star]
```

- [ ] Add extern func declaration
- [ ] Wire up Go implementation
- [ ] Benchmark query times

### Task 6.3: Query Functions
**File:** `sim/starmap.ail`

```ailang
-- Fast radius query using octree
export func queryStarsNear(center: Vec3, radius: float) -> [Star] ! {} {
    octreeQuery(center, radius)
}

-- Find N nearest stars
export func nearestNStars(center: Vec3, n: int) -> [Star] ! {} {
    -- Expand radius until we have N stars
    let initial = octreeQuery(center, 50.0);
    match length(initial) >= n {
        true => take(n, sortByDistance(initial, center)),
        false => nearestNStars(center, n)  -- Expand search
    }
}
```

- [ ] Add queryStarsNear() wrapper
- [ ] Add nearestNStars() function
- [ ] Test performance with large catalogs

**Day 6 Test:**
```bash
go test -bench=. ./engine/stardata/...
# Query 5000 ly radius should be <100ms
```

---

## Day 7: Integration & Visual Verification

### Task 7.1: Galaxy Map Rendering (Basic)
**File:** `sim/starmap.ail`

```ailang
-- Generate DrawCmds for visible stars
export pure func renderStarmap(catalog: StarCatalog, viewCenter: Vec3, viewRadius: float, screenW: float, screenH: float) -> [DrawCmd] {
    let visible = starsWithinRadius(catalog, viewCenter, viewRadius);
    map(\s. starToDrawCmd(s, viewCenter, screenW, screenH), visible)
}

pure func starToDrawCmd(star: Star, viewCenter: Vec3, w: float, h: float) -> DrawCmd {
    let rel = { x: star.pos.x - viewCenter.x, y: star.pos.y - viewCenter.y, z: star.pos.z - viewCenter.z };
    let screenX = w / 2.0 + rel.x * 0.1;  -- Scale factor
    let screenY = h / 2.0 + rel.y * 0.1;
    let brightness = match star.spectral {
        O => 255, B => 220, A => 200, F => 180, G => 160, K => 140, M => 100
    };
    Star(screenX, screenY, intToFloat(brightness) / 255.0, spectralColor(star.spectral))
}
```

- [ ] Add renderStarmap() function
- [ ] Add spectralColor() for star colors
- [ ] Generate DrawCmds for visible stars

### Task 7.2: Demo Command
**File:** `cmd/demo-starmap/main.go` (NEW)

```go
// Simple demo showing starmap centered on Sol
// Arrow keys to pan, scroll to zoom
```

- [ ] Create demo-starmap command
- [ ] Load catalog at startup
- [ ] Render stars with panning/zoom
- [ ] Display Sol neighborhood

### Task 7.3: Visual Verification

- [ ] Verify Sol at center
- [ ] Verify Alpha Centauri ~4.4 ly away
- [ ] Verify star density increases toward galactic center
- [ ] Verify spectral type colors (M=red, G=yellow, B=blue)

**Day 7 Test:**
```bash
go build -o bin/demo-starmap ./cmd/demo-starmap
bin/demo-starmap
# Visual inspection of star field
```

---

## AILANG Feedback Checkpoint

After sprint completion, report:

### Issues Encountered
- [ ] Document any AILANG bugs found
- [ ] Note features that would have helped
- [ ] Record workarounds used

### Performance Notes
- [ ] AILANG list operations at scale
- [ ] extern func integration experience
- [ ] Codegen performance

### Send Summary
```bash
ailang messages send user "Starmap sprint complete. Issues: ..." \
  --from "stapledons_voyage" \
  --title "Sprint Complete: Starmap Data Model"
```

---

## Files Created/Modified

### New AILANG Files
| File | LOC | Purpose |
|------|-----|---------|
| `sim/starmap.ail` | ~200 | Star types, catalog, queries |
| `sim/galaxy_model.ail` | ~150 | Density model, procedural gen |

### New Go Files
| File | LOC | Purpose |
|------|-----|---------|
| `engine/stardata/gaia_import.go` | ~150 | CSV parsing |
| `engine/stardata/coords.go` | ~80 | Coordinate conversion |
| `engine/stardata/catalog.go` | ~100 | Catalog building |
| `engine/stardata/octree.go` | ~200 | Spatial indexing |
| `cmd/demo-starmap/main.go` | ~150 | Visual demo |

### Data Files
| File | Size | Purpose |
|------|------|---------|
| `assets/data/starmap/gaia_nearby.csv` | ~5MB | Gaia subset |

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Gaia data too large | Med | Pre-filter to ~10k stars |
| Octree in pure AILANG too slow | High | Use extern func for Go impl |
| Coordinate conversion errors | Med | Test with known star positions |
| Procedural/real boundary visible | Low | Smooth blending over 40 ly |

---

## References

- [starmap-data-model.md](../design_docs/planned/phase1-data-models/starmap-data-model.md) - Design doc
- [Gaia DR3 Documentation](https://www.cosmos.esa.int/web/gaia/dr3)
- [engine-capabilities.md](../design_docs/reference/engine-capabilities.md) - Engine reference

---

**Document created**: 2025-12-11
