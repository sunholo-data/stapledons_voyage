# Special Relativity Rendering System

**Status:** Planned
**Priority:** P1
**Complexity:** High
**Depends On:** Galaxy Map (v0.5.2), Journey System (v0.6.0)
**AILANG Workarounds:** None required (engine-side implementation)

## Related Documents

- [Journey System](v0_6_0/journey-system.md) - SR effects active during journey
- [Galaxy Map](v0_5_2/galaxy-map.md) - Starfield rendering (pre-SR transform)
- [Black Holes](black-holes.md) - GR lensing effects (separate system)
- [docs/vision/design-decisions.md](../../docs/vision/design-decisions.md) - "SR Visual Effects: Hard SF Made Visible"

## Overview

Implement physically accurate Special Relativity visual effects for relativistic travel. The goal is to make time dilation *visible* and *visceral*, not just numerical.

**Core principle:** The sim says "how fast am I going"; the renderer bends the light.

## The Three SR Effects

At high γ (10-20), players experience:

### 1. Aberration (Headlight Effect)

Stars pile into a forward cone as speed increases. The universe "rushes toward" your velocity vector.

- **Low γ:** Stars distributed normally across sky
- **High γ:** Almost everything visible in narrow cone ahead; behind is blackness

**Physical cause:** Light that arrives from the side/back in the galaxy frame appears squeezed into the forward direction in the ship frame.

### 2. Doppler Shift (Colour Change)

Light from ahead blue-shifts; light from behind red-shifts.

- **Forward cone:** Stars become blue/white
- **Sideways:** Slight shift
- **Rear:** Stars shift to red/infrared and effectively disappear

Nebulae and background textures get blue tint ahead, red smear behind.

### 3. Relativistic Beaming (Brightness)

Intensity scales as I' ∝ D³ where D is the Doppler factor.

- **Forward:** Bright halo, potential "star wind" bloom
- **Rear:** Dims to near-black

At high γ, creates the "searchlight" / "warp tunnel" feeling grounded in physics.

## AILANG / Engine Boundary

### AILANG Outputs (sim layer)

```ailang
type CameraState = {
    pos: Vec3,         -- Ship position in galaxy frame
    vel: Vec3,         -- Velocity as fraction of c (β vector)
    gamma: float,      -- Lorentz factor (or derived from |vel|)
    mode: CameraMode   -- Docked | Orbital | Cruise | NearBH
}

type CameraMode =
    | Docked           -- Normal rendering (no SR effects)
    | Orbital          -- Normal rendering (minor SR at high orbital speed?)
    | Cruise           -- Full SR effects
    | NearBH           -- SR + GR lensing

-- Starfield in galaxy rest frame (before SR transform)
type StarData = {
    dir: Vec3,         -- Unit vector from camera to star
    base_color: RGB,   -- Rest-frame colour
    base_intensity: float
}
```

AILANG provides physical state. It does NOT know about pixels or screen space.

### Engine Outputs (Go/Ebiten layer)

All SR math lives in the engine:

1. **Aberration transform** - Transform star directions from galaxy frame to ship frame
2. **Doppler calculation** - Compute D for each direction, shift colours
3. **Beaming** - Multiply intensities by D³ (clamped)
4. **Projection** - Convert ship-frame directions to screen positions

## Core Math (Engine Implementation)

### Variables

```go
β   := ship velocity vector (units of c, |β| < 1)
γ   := 1 / sqrt(1 - |β|²)
n   := unit vector in galaxy frame pointing toward light source
n'  := direction in ship frame (what camera sees)
D   := Doppler factor
```

### Aberration (Direction Transformation)

Split incoming direction into components parallel and perpendicular to velocity:

```go
β_hat := normalize(β)                    // unit velocity direction
n_par := dot(n, β_hat) * β_hat           // parallel component
n_perp := n - n_par                      // perpendicular component

// Transform to ship frame
denom := 1 - dot(β, n)
n'_par := (n_par - β) / denom
n'_perp := n_perp / (γ * denom)

n' := normalize(n'_par + n'_perp)
```

### Doppler Factor

```go
D := γ * (1 - dot(β, n))
```

Properties:
- Ahead of motion (β·n > 0): D < 1, blue-shift
- Behind motion (β·n < 0): D > 1, red-shift
- The exact sign depends on convention; verify against test cases

### Colour Shift

Approximate approach:

```go
// Map D to colour temperature shift
T' := T * D

// Or simpler: interpolate toward blue (D < 1) or red (D > 1)
if D < 1 {
    color = lerp(base_color, blue_tint, 1 - D)
} else {
    color = lerp(base_color, red_tint, (D - 1) / max_red_shift)
}
```

### Beaming (Intensity)

```go
brightness := base_intensity * clamp(pow(D, 3), min_bright, max_bright)
```

Clamp prevents infinities and keeps visuals readable.

## Implementation Strategy

### Option A: Per-Star CPU Transform (Discrete Stars)

For starfield with discrete star sprites:

```go
func (r *Renderer) TransformStars(stars []StarData, β Vec3, γ float64) []ScreenStar {
    result := make([]ScreenStar, 0, len(stars))

    for _, star := range stars {
        // Aberration
        n_prime := aberrate(star.Dir, β, γ)

        // Skip if behind camera
        if dot(n_prime, camera_forward) < 0 {
            continue
        }

        // Doppler
        D := γ * (1 - dot(β, star.Dir))

        // Colour shift
        color := shiftColor(star.BaseColor, D)

        // Beaming
        intensity := star.BaseIntensity * clamp(pow(D, 3), 0.01, 100)

        // Project to screen
        screen_pos := project(n_prime)

        result = append(result, ScreenStar{
            Pos:       screen_pos,
            Color:     color,
            Intensity: intensity,
        })
    }

    return result
}
```

Works for tens of thousands of stars. Good for our HIP catalog subset.

### Option B: Shader-Based Cubemap Warp (Background)

For nebulae / galactic background:

```kage
// Fragment shader (Ebiten Kage)
package main

var Beta Vec3    // ship velocity (uniform)
var Gamma float  // Lorentz factor (uniform)

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
    // Get direction from camera in ship frame
    n_prime := screenToDirection(position.xy)

    // Invert aberration to get galaxy-frame direction
    n := inverseAberrate(n_prime, Beta, Gamma)

    // Sample cubemap at galaxy-frame direction
    base_color := sampleCubemap(n)

    // Compute Doppler factor
    D := Gamma * (1 - dot(Beta, n))

    // Shift colour
    shifted := shiftColor(base_color, D)

    // Apply beaming
    brightness := clamp(pow(D, 3), 0.1, 10.0)

    return shifted * brightness
}
```

### Option C: Hybrid (Recommended)

- **Stars:** CPU-level transform (accurate, fewer objects)
- **Background:** Shader warp for nebulae/galactic arm textures
- **HUD/Effects:** Purely aesthetic overlays (lens flare, motion blur)

## Camera Modes

| Mode | SR Effects | Notes |
|------|-----------|-------|
| Docked | Off | Ship interior, station |
| Orbital | Minimal | Maybe slight effect at high orbital speed? |
| Cruise | Full | Main relativistic travel mode |
| NearBH | SR + GR | Add gravitational lensing overlay |

Transition: Effects fade in/out smoothly as γ increases/decreases.

Threshold suggestion: Effects begin subtle at γ ≈ 2, become pronounced at γ ≥ 5.

## HUD Elements

Beyond visual effects, SR should be reflected in UI:

### Dual Time Display

```
┌─────────────────────────────────┐
│  Ship Time:   37y 142d          │
│  Galaxy Time: 8,423y            │
│  γ = 15.3                       │
└─────────────────────────────────┘
```

### External Clock Distortion

Remote beacons / planet rotations / civilization activity shown running "fast" from ship's perspective.

## Visual Exaggeration

Real SR at γ=20 may be too extreme. Options:

1. **Pure physics:** D³ beaming, exact aberration
2. **Clamped physics:** Apply formulas but clamp to comfortable ranges
3. **Aesthetic mapping:** Map physical β → visual "wow factor" via custom curves

**Recommendation:** Start with clamped physics, tune based on playtesting.

## Open Questions

See [docs/vision/open-questions.md](../../docs/vision/open-questions.md):

- How much exaggeration?
- How should GR lensing near BH interact with SR?
- Mode transitions (docked vs cruise)?

## Success Criteria

### Visual

- [ ] At γ ≥ 10, stars visibly bunch into forward cone
- [ ] Forward stars are bluer/brighter than rest-frame
- [ ] Rear-view becomes sparse and dim
- [ ] Background textures (nebulae) warp correctly
- [ ] Smooth transition as γ increases/decreases

### Technical

- [ ] AILANG outputs CameraState with β, γ
- [ ] Engine implements aberration transform
- [ ] Engine implements Doppler colour shift
- [ ] Engine implements D³ beaming (clamped)
- [ ] Shader for background cubemap warp
- [ ] HUD shows dual time display

### Integration

- [ ] Effects activate in Cruise mode, deactivate in Docked
- [ ] Smooth fade-in/out at mode transitions
- [ ] Performance acceptable with full starfield

## References

- Penrose, R. "The Apparent Shape of a Relativistically Moving Sphere"
- Real Time Relativity (ANU project)
- A Slower Speed of Light (MIT Game Lab)
