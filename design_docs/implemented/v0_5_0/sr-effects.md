# Special Relativity Visual Effects

**Version:** 0.5.0
**Status:** Implemented
**Priority:** P1 - Core to game identity
**Package:** `engine/shader`
**AILANG Impact:** Pending integration (CameraState not yet connected)

## Related Documents

- [Shader System](../v0_4_5/shader-system.md) - Base shader infrastructure (implemented)
- [GR Visual Mechanics](../../planned/gr-visual-mechanics.md) - GR effects layer (planned)
- [Black Holes](../../planned/black-holes.md) - Black hole mechanics (planned)

## Implementation Summary

The Special Relativity visual effects are fully implemented on the engine side. At high velocities (β approaching c), players experience:

1. **Aberration (Headlight Effect)** - Stars pile into a forward cone
2. **Doppler Shift** - Forward stars blueshift, rear stars redshift
3. **Relativistic Beaming** - Forward brightens (D³), rear dims to black

## Implemented Components

### Go Wrapper: `engine/shader/sr.go`

```go
type SRWarp struct {
    // Velocity components (as fraction of c)
    betaX, betaY, betaZ float64
    gamma               float64   // Lorentz factor
    fov                 float64   // Field of view in radians
    viewAngle           float64   // 0=front, π/2=side, π=back
}
```

**Key methods:**
- `SetVelocity(betaX, betaY, betaZ)` - Set 3D velocity
- `SetForwardVelocity(beta)` - Convenience for forward motion
- `GetGamma()` - Returns computed Lorentz factor
- `Apply(dst, src *ebiten.Image)` - Apply shader effect

### Kage Shader: `engine/shader/shaders/sr_warp.kage`

Implements exact SR physics:

```kage
// Aberration: cos(θ) = (cos(θ') + β) / (1 + β·cos(θ'))
cosThetaPrime := cos(thetaPrime)
cosTheta := (cosThetaPrime + betaMag) / (1.0 + betaMag * cosThetaPrime)

// Doppler: D = γ(1 + β·cos(θ))
D := Gamma * (1.0 + betaMag * cosTheta)

// Beaming: I' = I × D³
beamFactor := D * D * D
```

Color shifts:
- D > 1 (approaching): RGB shifts toward blue
- D < 1 (receding): RGB shifts toward red

### Integration: `engine/shader/effects.go`

SR warp is integrated into the main effects pipeline:
- Toggle: F4 in demo mode
- Cycle velocity: Shift+F4 (0.5c → 0.9c → 0.95c → 0.99c)
- Applied first in effect chain (before bloom, vignette, etc.)

### Screenshot Support: `engine/screenshot/`

```go
type Config struct {
    Velocity   float64 // Ship velocity as fraction of c (0.0-0.99)
    ViewAngle  float64 // View direction: 0=front, π/2=side, π=back
}
```

## Demo Mode Controls

| Key | Action |
|-----|--------|
| F4 | Toggle SR warp effect |
| Shift+F4 | Cycle velocity (0.5c, 0.9c, 0.95c, 0.99c) |
| F9 | Show effect overlay with SR status |

## Visual Progression

| Velocity | γ | Forward View | Rear View |
|----------|---|--------------|-----------|
| 0.5c | 1.15 | Slight clustering, hint of blue | Slight thinning, hint of red |
| 0.9c | 2.29 | Clear cone, blue-white stars | Sparse, red-shifted |
| 0.95c | 3.20 | Tight cone, bright forward | Very sparse, deep red |
| 0.99c | 7.09 | Intense point, "searchlight" | Near black |

## Pending Work

### AILANG Integration (not yet implemented)

The design docs specify CameraState should include:

```ailang
type CameraState = {
    pos: Vec3,
    vel: Vec3,         -- Velocity as fraction of c
    gamma: float,      -- Lorentz factor
    sr_enabled: bool   -- Cruise mode flag
}
```

Currently the engine uses hardcoded demo values. Full integration requires:
1. Add velocity/gamma to FrameOutput from AILANG
2. Engine reads from sim output each frame
3. Mode transitions (Docked → Cruise) to enable/disable

### HUD Elements (not yet implemented)

- Dual time display (Ship Time vs Galaxy Time)
- Velocity indicator (β, γ)
- External clock distortion visualization

### Testing (not yet implemented)

- Unit tests for SR math functions
- Golden file tests at various velocities
- Performance benchmarks

## Files

| File | Purpose | LOC |
|------|---------|-----|
| `engine/shader/sr.go` | Go wrapper for SR warp | ~160 |
| `engine/shader/shaders/sr_warp.kage` | Kage shader with exact physics | ~125 |
| `engine/shader/effects.go` | Integration with effects pipeline | ~280 |
| `engine/screenshot/demo_scene.go` | Demo scene for SR testing | ~330 |
| `engine/screenshot/screenshot.go` | Screenshot config with velocity | ~320 |

## Success Criteria (Achieved)

- [x] Aberration formula implemented exactly
- [x] Doppler shift with correct D calculation
- [x] D³ beaming (clamped for stability)
- [x] Blueshift/redshift color adjustment
- [x] View angle support (front/side/back)
- [x] Demo mode toggle and velocity cycling
- [x] Integration with shader effects pipeline
- [x] Screenshot capture with velocity parameter

## References

- [Relativistic Aberration](https://en.wikipedia.org/wiki/Relativistic_aberration)
- [Relativistic Beaming](https://en.wikipedia.org/wiki/Relativistic_beaming)
- [MIT Game Lab - "A Slower Speed of Light"](http://gamelab.mit.edu/games/a-slower-speed-of-light/)

---

**Implemented:** 2025-12-05
**Origin:** Consolidated from sr-rendering.md and relativistic-visual-effects.md
