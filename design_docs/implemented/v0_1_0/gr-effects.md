# General Relativity Visual Effects

**Version:** 0.1.0
**Status:** Implemented
**Priority:** P1 (Core Visual Identity)
**Complexity:** High
**Package:** `engine/shader`, `engine/relativity`

## Related Documents

- [SR Effects](sr-effects.md) - Special relativity effects (implemented)
- [Shader System](shader-system.md) - Kage shader pipeline (implemented)
- [CLI Dev Tools](cli-dev-tools.md) - `granimation` uses GR effects

## Overview

General Relativity visual effects simulate the distortion of spacetime near massive objects like black holes, neutron stars, and white dwarfs. This creates a visceral experience of "the universe bending" as players approach these objects.

## Implementation Summary

**Total:** ~1,000 LOC across Go wrapper, context management, and Kage shaders.

| Component | Location | LOC | Purpose |
|-----------|----------|-----|---------|
| GR Wrapper | `engine/shader/gr.go` | 212 | Effect manager with demo mode |
| GR Context | `engine/relativity/gr_context.go` | 227 | Physics computations |
| Transforms | `engine/relativity/transform.go` | 183 | Vec3, coordinate math |
| Color | `engine/relativity/color.go` | 150 | Redshift color mapping |
| Lensing Shader | `engine/shader/shaders/gr_lensing.kage` | 112 | Gravitational lensing |
| Redshift Shader | `engine/shader/shaders/gr_redshift.kage` | 112 | Color shift effect |

## Key Types

### GRContext

```go
type GRContext struct {
    Active            bool
    ObjectKind        MassiveObjectKind // BlackHole, NeutronStar, WhiteDwarf
    Distance          float64           // Ship distance from object
    Phi               float64           // Dimensionless potential: GM/(rc²)
    TimeDilation      float64           // dτ/dt = sqrt(1 - r_s/r)
    RedshiftFactor    float64           // z_gr ≈ 1/sqrt(1 - r_s/r)
    TidalSeverity     float64           // 0..1 heuristic
    DangerLevel       GRDangerLevel     // None, Subtle, Strong, Extreme
    CanHoverSafely    bool
    NearPhotonSphere  bool
    Rs                float64           // Schwarzschild radius
}
```

### GRDangerLevel

Based on dimensionless potential Φ = r_s/(2r):

| Level | Φ Range | Visual Effects |
|-------|---------|----------------|
| None | Φ < 1e-4 | Only SR active |
| Subtle | 1e-4 ≤ Φ < 1e-3 | Light shading |
| Strong | 1e-3 ≤ Φ < 1e-2 | Visible lensing |
| Extreme | Φ ≥ 0.01 | Heavy distortion |

### GRWarp Effect

```go
type GRWarp struct {
    manager    *Manager
    enabled    bool
    uniforms   relativity.GRShaderUniforms

    // Demo mode for visualization
    demoMode   bool
    demoPhi    float64
    demoCenter [2]float32
    demoRs     float32
}

// Key methods
func (g *GRWarp) SetEnabled(enabled bool)
func (g *GRWarp) Toggle() bool
func (g *GRWarp) SetUniforms(uniforms GRShaderUniforms)
func (g *GRWarp) SetDemoMode(centerX, centerY, rs, phi float32)
func (g *GRWarp) Apply(dst, src *ebiten.Image) bool
func (g *GRWarp) ApplyRedshift(dst, src *ebiten.Image) bool
func (g *GRWarp) CycleDemoIntensity() string  // Subtle → Strong → Extreme
```

## Physics Implementation

### Schwarzschild Radius

```go
// r_s = 2GM/c² ≈ 2.95 km per solar mass
rsKm := 2.95 * massSolar
rs := rsKm / 1000.0  // Convert to game units
```

### Time Dilation

```go
// dτ/dt = sqrt(1 - r_s/r)
tdArg := clamp(1.0-rs/r, 0.001, 1.0)
timeDilation := math.Sqrt(tdArg)
```

### Gravitational Redshift

```go
// z ≈ 1/sqrt(1 - r_s/r)
redshiftFactor := 1.0 / timeDilation
```

### Photon Sphere Detection

```go
// Photon sphere at r = 1.5 r_s (only for black holes)
nearPhotonSphere := obj.Kind == BlackHole && r >= 1.3*rs && r <= 2.0*rs
```

## Shader Uniforms

```go
type GRShaderUniforms struct {
    Enabled         bool
    ObjectKind      int       // 0=BH, 1=NS, 2=WD
    Rs              float32   // Schwarzschild radius (screen units)
    Distance        float32   // Ship→object distance
    Phi             float32   // Dimensionless potential
    TimeDilation    float32   // dτ/dt
    RedshiftFactor  float32   // Gravitational z
    ScreenCenter    [2]float32 // Screen position of object
    MaxEffectRadius float32   // Effect bounds
    LensStrength    float32   // Tunable lensing
}
```

## Demo Mode

The GR effect includes a demo mode for visualization without active simulation:

```go
// Enable demo mode
grWarp.SetDemoMode(0.5, 0.5, 0.05, 0.05)  // center, rs, phi

// Cycle intensity
level := grWarp.CycleDemoIntensity()  // "Subtle" → "Strong" → "Extreme"
```

**Demo via CLI:**
```bash
./bin/granimation  # Generates 60 frames of black hole approach
```

Output: `out/gr-animation/frame_XXX.png`

**Creating video:**
```bash
ffmpeg -framerate 30 -i out/gr-animation/frame_%03d.png \
  -c:v libx264 -pix_fmt yuv420p out/gr-journey.mp4
```

## Game Integration

**Keyboard shortcuts (in-game):**
- F7: Toggle GR effects
- F8: Cycle GR intensity (demo mode)

**Effects pipeline:**
```go
// In shader.Effects
func (e *Effects) Apply(src *ebiten.Image) *ebiten.Image {
    // ... SR effects first

    // GR lensing
    if e.grWarp.Apply(dst, src) {
        src = dst
    }

    // GR redshift (layered)
    if e.grWarp.ApplyRedshift(dst, src) {
        src = dst
    }

    // ... bloom, vignette, etc
}
```

## Success Criteria

### Implemented
- [x] GRContext physics computations
- [x] MassiveObject types (BlackHole, NeutronStar, WhiteDwarf)
- [x] Danger level classification
- [x] GRWarp shader wrapper
- [x] Demo mode with intensity cycling
- [x] Gravitational lensing shader
- [x] Gravitational redshift shader
- [x] Shader uniform passing
- [x] Animation generation tool (granimation)

### Pending (for full gameplay integration)
- [ ] AILANG GRContext integration
- [ ] Dynamic object tracking in starmap
- [ ] Tidal stress gameplay effects
- [ ] Photon sphere warning indicators

## Code References

| File | Line | Purpose |
|------|------|---------|
| `engine/shader/gr.go` | 1-212 | GRWarp effect wrapper |
| `engine/relativity/gr_context.go` | 136-184 | ComputeGRContext() |
| `engine/relativity/gr_context.go` | 121-133 | ClassifyDangerLevel() |
| `engine/shader/shaders/gr_lensing.kage` | 1-112 | Lensing shader |
| `engine/shader/shaders/gr_redshift.kage` | 1-112 | Redshift shader |
| `cmd/granimation/main.go` | 1-138 | Animation generator |

---

**Document created**: 2025-12-06
**Last updated**: 2025-12-06
