# Relativistic Visual Effects

**Status**: Planned
**Target**: v0.5.0
**Priority**: P1 - Core to game identity
**Estimated**: 5-7 days
**Dependencies**: Shader pipeline (implemented), Star rendering (implemented)

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | + | +1 | **Core visual reinforcement** - player SEES relativity, not just reads about it |
| Civilization Simulation | 0 | 0 | Doesn't directly affect civs, but shows galaxy from moving frame |
| Philosophical Depth | + | +1 | Visceral demonstration of SR - "the universe bends around you" |
| Ship & Crew Life | 0 | 0 | Indirect - crew would experience this view |
| Legacy Impact | 0 | 0 | Visual effect, doesn't affect legacy calculation |
| Hard Sci-Fi Authenticity | + | +1 | **Maximum relevance** - real SR optics, not fake "warp streaks" |
| **Net Score** | | **+3** | **Decision: Move forward - core differentiator** |

**Feature type:** Engine/Rendering
- This is foundational visual infrastructure that makes the game's hard-SF premise *visible*
- "Show, don't tell" for special relativity

**Reference:** See [game-vision.md](../../docs/game-vision.md)

## Problem Statement

**Current State:**
- Ship can travel at relativistic speeds (up to 0.999c)
- Player sees a normal starfield regardless of velocity
- Time dilation is shown numerically but not *felt* visually
- Generic "space game" aesthetic instead of authentic SR experience

**Impact:**
- Misses opportunity to make relativity visceral, not abstract
- Fails to differentiate from other space games
- Breaks scientific authenticity pillar

## Goals

**Primary Goal:** Make relativistic travel *look* relativistic - stars bunch forward, blueshift ahead, redshift behind, rear sky goes dark.

**Success Metrics:**
- At v=0.9c (gamma ~2.3): noticeable forward clustering, color shift visible
- At v=0.99c (gamma ~7): dramatic "tunnel vision", strong blue/red shift
- At v=0.999c (gamma ~22): almost all stars in small forward cone, intense beaming
- Performance: <2ms GPU time for SR shader at 1080p
- Smooth transition as ship accelerates/decelerates

## Solution Design

### Overview

Implement Special Relativity optical effects using post-processing shaders:

1. **Aberration**: Transform star directions from galaxy frame to ship frame
2. **Doppler Shift**: Shift colors based on relative velocity angle
3. **Relativistic Beaming**: Adjust brightness (I' ~ D^3)

The key insight: **AILANG handles kinematics** (ship position, velocity), **Engine handles optics** (SR transforms, rendering).

### The Three SR Optical Effects

#### 1. Aberration (Direction Warping)

Stars that are "beside" or "behind" the ship in the galaxy frame appear shifted forward in the ship's frame. At high gamma, almost everything collapses into a forward cone.

**Visual effect:** "Tunnel vision" - stars pile into the direction of travel

**Math:** For a photon arriving from direction `n` in galaxy frame:
```
n'_parallel = (n_parallel - beta) / (1 - beta . n)
n'_perp = n_perp / (gamma * (1 - beta . n))
n' = normalize(n'_parallel + n'_perp)
```

#### 2. Doppler Shift (Color Change)

Light from ahead is blueshifted; light from behind is redshifted.

**Visual effect:**
- Forward stars: shift toward blue/white
- Rear stars: shift toward red, then infrared (invisible)

**Math:** Doppler factor:
```
D = gamma * (1 - beta . n)
frequency' = D * frequency
```

For colors, approximate by shifting color temperature: `T' = T * D`

#### 3. Relativistic Beaming (Brightness)

Forward directions get brighter; rear directions get dimmer.

**Visual effect:**
- Forward: "searchlight" effect, stars blow out to bright halo
- Rear: fades to near-black

**Math:** Intensity scales as:
```
I' = I * D^3
```

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         AILANG (sim/)                           │
│                                                                 │
│  Ship State:                                                    │
│  - position: Vec3          (galaxy frame)                       │
│  - velocity: Vec3          (beta vector, units of c)            │
│  - gamma: float            (Lorentz factor)                     │
│  - cruise_mode: bool       (enable SR rendering)                │
│                                                                 │
│  Star Catalog:                                                  │
│  - positions in galaxy frame                                    │
│  - base colors, magnitudes                                      │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │ FrameOutput includes:
                            │ - camera.velocity (Vec3)
                            │ - camera.gamma (float)
                            │ - camera.sr_enabled (bool)
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│                      Engine (Go/Ebiten)                         │
│                                                                 │
│  Starfield Renderer:                                            │
│  - CPU: transform star directions via aberration formula        │
│  - CPU: compute Doppler factor per star                         │
│  - CPU: adjust color & brightness                               │
│  - GPU: render points/sprites                                   │
│                                                                 │
│  Background Shader (skybox/nebulae):                            │
│  - GPU: for each pixel, invert aberration to sample galaxy-frame│
│  - GPU: apply Doppler shift to sampled color                    │
│  - GPU: apply beaming to brightness                             │
│                                                                 │
│  Post-Processing (existing pipeline):                           │
│  - Bloom (enhanced for beaming)                                 │
│  - Vignette (composable with SR darkening)                      │
└─────────────────────────────────────────────────────────────────┘
```

**Components:**

1. **RelativisticCamera** (new Go module)
   - Takes beta, gamma from sim
   - Provides SR transform functions for renderer

2. **SR Star Shader** (new Kage shader)
   - Per-star aberration + Doppler + beaming
   - Or CPU pre-transform + simple sprite draw

3. **SR Background Shader** (new Kage shader)
   - Warps skybox/nebula textures
   - Inverse aberration sampling

4. **Enhanced Bloom** (modify existing)
   - Beaming can blow out forward stars
   - Bloom makes this look like "star wind"

### Implementation Plan

**Phase 1: AILANG Ship State** (~4 hours)
- [ ] Add `velocity: Vec3` to Camera/Ship state
- [ ] Add `gamma: float` computed from |velocity|
- [ ] Add `sr_enabled: bool` for cruise mode
- [ ] Export in FrameOutput

**Phase 2: CPU Starfield SR Transform** (~8 hours)
- [ ] Create `engine/relativity/transform.go`
- [ ] Implement aberration direction transform
- [ ] Implement Doppler factor calculation
- [ ] Implement color temperature shift
- [ ] Implement brightness beaming
- [ ] Integrate with star rendering in `engine/render/draw.go`

**Phase 3: SR Background Shader** (~8 hours)
- [ ] Create `engine/shader/shaders/sr_background.kage`
- [ ] Implement inverse aberration (pixel dir -> galaxy dir)
- [ ] Sample existing galaxy background at transformed coord
- [ ] Apply Doppler color shift
- [ ] Apply beaming brightness
- [ ] Integrate with galaxy background rendering

**Phase 4: Visual Polish** (~8 hours)
- [ ] Tune parameters for gamma 2-20 range
- [ ] Add smooth transition as ship accelerates
- [ ] Integrate with existing bloom for "star wind" effect
- [ ] Test edge cases (v=0, v->c)
- [ ] Screenshot tests for different velocities

**Phase 5: UI Integration** (~4 hours)
- [ ] HUD shows "ship time" vs "galaxy time"
- [ ] Velocity indicator (beta, gamma)
- [ ] Optional "external view" toggle (see ship from galaxy frame)

### Files to Modify/Create

**New files:**
- `engine/relativity/transform.go` - SR math functions (~200 LOC)
- `engine/relativity/camera.go` - Relativistic camera wrapper (~100 LOC)
- `engine/shader/shaders/sr_background.kage` - Background warp shader (~80 LOC)
- `engine/shader/shaders/sr_star.kage` - Optional star shader (~60 LOC)

**Modified files:**
- `sim/protocol.ail` - Add velocity/gamma to Camera (~10 LOC)
- `sim/step.ail` - Compute gamma from velocity (~20 LOC)
- `engine/render/draw.go` - Use SR transforms for stars (~50 LOC)
- `engine/shader/effects.go` - Add SR effect to pipeline (~30 LOC)

## Examples

### Example 1: Star Direction Transform

**Input (galaxy frame):**
```go
starDir := Vec3{0, 0, -1}  // Star directly behind ship
beta := Vec3{0, 0, 0.9}    // Ship moving forward at 0.9c
gamma := 2.294             // Lorentz factor
```

**Output (ship frame):**
```go
// Star appears shifted forward due to aberration
observedDir := TransformDirection(starDir, beta, gamma)
// observedDir ≈ {0, 0, -0.36} (still behind, but less so)
// At higher gamma, would shift into forward hemisphere
```

### Example 2: Color Shift

**Input:**
```go
baseColor := RGB{255, 244, 214}  // Yellow star (5800K)
dopplerFactor := 1.5             // Approaching
```

**Output:**
```go
// Temperature shift: 5800K * 1.5 = 8700K
// Color shifts toward blue-white
shiftedColor := DopplerShiftColor(baseColor, dopplerFactor)
// shiftedColor ≈ RGB{200, 210, 255}  // Blue-white
```

### Example 3: Visual Progression

| Ship Velocity | Gamma | Forward View | Rear View |
|--------------|-------|--------------|-----------|
| 0.0c | 1.0 | Normal starfield | Normal starfield |
| 0.5c | 1.15 | Slight clustering, hint of blue | Slight thinning, hint of red |
| 0.9c | 2.29 | Clear cone, blue-white stars | Sparse, red-shifted |
| 0.99c | 7.09 | Tight cone, brilliant forward | Near black, few dim red |
| 0.999c | 22.4 | Intense point, "searchlight" | Black |

## SR Math Reference

### Core Formulas

**Lorentz factor:**
```
gamma = 1 / sqrt(1 - |beta|^2)
```

**Doppler factor for direction n:**
```
D = gamma * (1 - dot(beta, n))
```
- D > 1: blueshift (approaching)
- D < 1: redshift (receding)

**Aberration transform:**
```go
func TransformDirection(n, beta Vec3, gamma float64) Vec3 {
    betaMag := beta.Length()
    if betaMag < 1e-10 {
        return n
    }

    betaHat := beta.Normalize()
    nParallel := betaHat.Scale(n.Dot(betaHat))
    nPerp := n.Sub(nParallel)

    denom := 1.0 - beta.Dot(n)

    nPrimeParallel := nParallel.Sub(beta).Scale(1.0 / denom)
    nPrimePerp := nPerp.Scale(1.0 / (gamma * denom))

    return nPrimeParallel.Add(nPrimePerp).Normalize()
}
```

**Color temperature shift:**
```go
func ShiftColorTemp(baseTemp, dopplerFactor float64) float64 {
    return baseTemp * dopplerFactor
}

func TempToRGB(temp float64) RGB {
    // Planckian locus approximation
    // ...
}
```

**Brightness beaming:**
```go
func BeamBrightness(baseIntensity, dopplerFactor float64) float64 {
    // Clamp to avoid infinities
    d := clamp(dopplerFactor, 0.01, 100.0)
    return baseIntensity * d * d * d
}
```

## Success Criteria

- [ ] At v=0: stars render normally (SR disabled or D=1)
- [ ] At v=0.9c: visible forward clustering, noticeable color shift
- [ ] At v=0.99c: dramatic tunnel effect, strong beaming
- [ ] Smooth interpolation during acceleration/deceleration
- [ ] Performance: <2ms GPU for background shader
- [ ] Performance: <5ms CPU for 10k star transforms
- [ ] No visual artifacts at gamma > 10
- [ ] Works with existing bloom, vignette effects
- [ ] Screenshot tests pass for multiple velocities

## Testing Strategy

**Unit tests:**
- `transform_test.go`: Verify aberration formula against known cases
- Test gamma=1 (no transform), gamma=2, gamma=10
- Test Doppler factor for various angles

**Visual tests (golden files):**
- `out/sr/v0.0.png` - Stationary
- `out/sr/v0.5.png` - Half light speed
- `out/sr/v0.9.png` - 90% light speed
- `out/sr/v0.99.png` - 99% light speed

**Manual testing:**
- Verify "feels right" at various speeds
- Check for jarring transitions
- Verify bloom interaction looks like "star wind"

## Non-Goals

**Not in this feature:**
- **General Relativity effects** (gravitational lensing near black holes) - separate feature
- **Terrell rotation** (apparent rotation of extended objects) - too subtle
- **Exact spectral line shifts** - approximate with color temperature
- **Relativistic mass visualization** - not visually relevant

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Shader too expensive at high gamma | Med | Use LOD - fewer stars processed at high gamma |
| Math instabilities near v=c | Med | Clamp gamma to max 50, careful float handling |
| Looks "wrong" to physics-aware players | High | Follow SR formulas exactly, document approach |
| Disorienting for players | Med | Gradual transition, optional "comfort mode" |

## References

- [Special Relativity Visual Effects](https://en.wikipedia.org/wiki/Relativistic_beaming)
- [Relativistic Aberration](https://en.wikipedia.org/wiki/Relativistic_aberration)
- [MIT Game Lab - "A Slower Speed of Light"](http://gamelab.mit.edu/games/a-slower-speed-of-light/)
- [Real-time Relativistic Rendering (paper)](https://arxiv.org/abs/physics/0511081)
- Existing shader pipeline: `engine/shader/`
- Star rendering: `engine/render/draw.go:drawStar()`

## Future Work

- **GR Lensing**: Black hole gravitational lensing (separate effect)
- **Relativistic Doppler Sound**: Audio frequency shift
- **External Observer Mode**: See your ship from galaxy frame
- **Time Dilation Visualization**: Clocks running at different rates
- **Headlight Effect on Comms**: Messages appear frequency-shifted

---

**Document created**: 2025-12-05
**Last updated**: 2025-12-05
