# SR Rendering: Go Implementation

**Status:** Planned
**Priority:** P1
**Complexity:** High
**Depends On:** [SR Rendering](sr-rendering.md) (design), Galaxy Map (v0.5.2)
**Location:** `engine/relativity/`

## Overview

Go implementation of Special Relativity visual effects. This document covers the engine-side code that transforms galaxy-frame starfield data into ship-frame visuals.

**Key principle:** AILANG outputs physical state (β, γ). Go does all light-bending math.

## Package Structure

```
engine/
├── relativity/
│   ├── camera.go       # RelativisticCamera type and transforms
│   ├── aberration.go   # Direction transformation math
│   ├── doppler.go      # Colour shift and beaming
│   ├── starfield.go    # Batch star transformation
│   ├── shader.go       # Kage shader compilation/uniforms
│   └── camera_test.go  # Unit tests with known SR examples
└── render/
    └── draw.go         # Integration (modified)
```

## Types

### Core Types (engine/relativity/camera.go)

```go
package relativity

import "math"

// Vec3 is a 3D vector for positions and directions
type Vec3 struct {
    X, Y, Z float64
}

func (v Vec3) Dot(u Vec3) float64 {
    return v.X*u.X + v.Y*u.Y + v.Z*u.Z
}

func (v Vec3) Length() float64 {
    return math.Sqrt(v.Dot(v))
}

func (v Vec3) Normalize() Vec3 {
    l := v.Length()
    if l == 0 {
        return Vec3{}
    }
    return Vec3{v.X / l, v.Y / l, v.Z / l}
}

func (v Vec3) Sub(u Vec3) Vec3 {
    return Vec3{v.X - u.X, v.Y - u.Y, v.Z - u.Z}
}

func (v Vec3) Add(u Vec3) Vec3 {
    return Vec3{v.X + u.X, v.Y + u.Y, v.Z + u.Z}
}

func (v Vec3) Scale(s float64) Vec3 {
    return Vec3{v.X * s, v.Y * s, v.Z * s}
}

// CameraMode determines which rendering pipeline to use
type CameraMode int

const (
    CameraModeDocked  CameraMode = iota // Normal rendering, no SR
    CameraModeOrbital                    // Minimal SR (high orbital speeds)
    CameraModeCruise                     // Full SR effects
    CameraModeNearBH                     // SR + GR lensing
)

// RelativisticCamera holds ship state for SR transformations
type RelativisticCamera struct {
    // Ship state in galaxy frame
    Position Vec3    // Position (for parallax, not used in SR)
    Beta     Vec3    // Velocity as fraction of c (|β| < 1)
    Gamma    float64 // Lorentz factor: 1/sqrt(1-|β|²)

    // Rendering mode
    Mode CameraMode

    // Derived values (computed once per frame)
    BetaMag   float64 // |β|
    BetaHat   Vec3    // Unit velocity direction

    // Tuning parameters
    BeamingMin   float64 // Minimum brightness multiplier
    BeamingMax   float64 // Maximum brightness multiplier
    DopplerScale float64 // How aggressively to shift colours (0-1)
}

// NewRelativisticCamera creates a camera from AILANG output
func NewRelativisticCamera(beta Vec3, mode CameraMode) *RelativisticCamera {
    betaMag := beta.Length()

    // Clamp to prevent singularity at v=c
    if betaMag >= 1.0 {
        betaMag = 0.9999
        beta = beta.Normalize().Scale(betaMag)
    }

    gamma := 1.0 / math.Sqrt(1.0 - betaMag*betaMag)

    betaHat := Vec3{}
    if betaMag > 0 {
        betaHat = beta.Scale(1.0 / betaMag)
    }

    return &RelativisticCamera{
        Beta:         beta,
        Gamma:        gamma,
        Mode:         mode,
        BetaMag:      betaMag,
        BetaHat:      betaHat,
        BeamingMin:   0.01,
        BeamingMax:   100.0,
        DopplerScale: 1.0,
    }
}

// EffectStrength returns 0-1 based on how much SR effects should apply
// Smooth fade-in from γ=1 (no effect) to γ=5 (full effect)
func (c *RelativisticCamera) EffectStrength() float64 {
    if c.Mode == CameraModeDocked {
        return 0
    }
    if c.Gamma < 1.1 {
        return 0
    }
    if c.Gamma >= 5.0 {
        return 1.0
    }
    // Linear interpolation from γ=1.1 to γ=5
    return (c.Gamma - 1.1) / (5.0 - 1.1)
}
```

### Star Data Types

```go
// GalaxyStar is a star in galaxy rest frame (from AILANG)
type GalaxyStar struct {
    Dir       Vec3    // Unit direction from camera to star
    BaseColor RGB     // Rest-frame colour
    Intensity float64 // Rest-frame brightness
    ID        int     // For tracking/selection
}

// RGB is a colour in 0-1 range
type RGB struct {
    R, G, B float64
}

// ScreenStar is a transformed star ready for rendering
type ScreenStar struct {
    ScreenX   float64 // Screen X coordinate
    ScreenY   float64 // Screen Y coordinate
    Color     RGB     // Doppler-shifted colour
    Intensity float64 // Beaming-adjusted brightness
    Visible   bool    // False if culled (behind camera or too dim)
}
```

## Aberration (engine/relativity/aberration.go)

```go
package relativity

// Aberrate transforms a galaxy-frame direction to ship-frame direction.
// n is unit vector pointing toward light source in galaxy frame.
// Returns unit vector in ship frame.
func (c *RelativisticCamera) Aberrate(n Vec3) Vec3 {
    if c.BetaMag < 1e-10 {
        return n // No velocity, no aberration
    }

    // Split n into components parallel and perpendicular to β
    nDotBetaHat := n.Dot(c.BetaHat)
    nPar := c.BetaHat.Scale(nDotBetaHat)  // Parallel component
    nPerp := n.Sub(nPar)                   // Perpendicular component

    // Denominator: 1 - β·n
    denom := 1.0 - c.Beta.Dot(n)
    if math.Abs(denom) < 1e-10 {
        denom = 1e-10 // Avoid division by zero
    }

    // Transform components
    // n'_par = (n_par - β) / (1 - β·n)
    nPrimePar := nPar.Sub(c.Beta).Scale(1.0 / denom)

    // n'_perp = n_perp / (γ(1 - β·n))
    nPrimePerp := nPerp.Scale(1.0 / (c.Gamma * denom))

    // Combine and normalize
    nPrime := nPrimePar.Add(nPrimePerp)
    return nPrime.Normalize()
}

// InverseAberrate transforms ship-frame direction back to galaxy-frame.
// Used for background cubemap sampling: given screen direction, find galaxy direction.
func (c *RelativisticCamera) InverseAberrate(nPrime Vec3) Vec3 {
    if c.BetaMag < 1e-10 {
        return nPrime
    }

    // Inverse transformation (swap sign of β)
    // Split n' into components parallel and perpendicular to β
    nPrimeDotBetaHat := nPrime.Dot(c.BetaHat)
    nPrimePar := c.BetaHat.Scale(nPrimeDotBetaHat)
    nPrimePerp := nPrime.Sub(nPrimePar)

    // Denominator for inverse: 1 + β·n'
    denom := 1.0 + c.Beta.Dot(nPrime)
    if math.Abs(denom) < 1e-10 {
        denom = 1e-10
    }

    // Inverse transform (note: adding β instead of subtracting)
    nPar := nPrimePar.Add(c.Beta).Scale(1.0 / denom)
    nPerp := nPrimePerp.Scale(1.0 / (c.Gamma * denom))

    n := nPar.Add(nPerp)
    return n.Normalize()
}
```

## Doppler (engine/relativity/doppler.go)

```go
package relativity

import "math"

// DopplerFactor computes D = γ(1 - β·n) for direction n (galaxy frame)
// D < 1 means blue-shift (approaching), D > 1 means red-shift (receding)
func (c *RelativisticCamera) DopplerFactor(n Vec3) float64 {
    return c.Gamma * (1.0 - c.Beta.Dot(n))
}

// ShiftColor applies Doppler shift to a colour.
// D < 1: shift toward blue
// D > 1: shift toward red
func (c *RelativisticCamera) ShiftColor(col RGB, D float64) RGB {
    // Apply tuning scale
    scale := c.DopplerScale

    // Compute effective shift: D=1 means no shift
    // Blend between current colour and shifted colour
    if D < 1.0 {
        // Blue shift: interpolate toward blue
        blueCol := RGB{0.6, 0.7, 1.0}
        blend := (1.0 - D) * scale
        if blend > 1.0 {
            blend = 1.0
        }
        return lerpRGB(col, blueCol, blend)
    } else {
        // Red shift: interpolate toward red
        redCol := RGB{1.0, 0.5, 0.3}
        blend := (D - 1.0) * scale * 0.5 // Red shift less aggressive
        if blend > 1.0 {
            blend = 1.0
        }
        return lerpRGB(col, redCol, blend)
    }
}

// ShiftColorTemperature applies Doppler shift as temperature change.
// More physically accurate but requires colour temperature handling.
func (c *RelativisticCamera) ShiftColorTemperature(tempK float64, D float64) float64 {
    // T' = T * D
    // Note: D < 1 means higher temp (blue), D > 1 means lower temp (red)
    return tempK / D // Inverse because T ∝ 1/D for observed temp
}

// Beaming computes intensity multiplier from Doppler factor.
// I' = I * D^3 (clamped)
func (c *RelativisticCamera) Beaming(D float64) float64 {
    intensity := math.Pow(D, 3)

    // Clamp to prevent extreme values
    if intensity < c.BeamingMin {
        return c.BeamingMin
    }
    if intensity > c.BeamingMax {
        return c.BeamingMax
    }
    return intensity
}

// lerpRGB linearly interpolates between two colours
func lerpRGB(a, b RGB, t float64) RGB {
    return RGB{
        R: a.R + (b.R-a.R)*t,
        G: a.G + (b.G-a.G)*t,
        B: a.B + (b.B-a.B)*t,
    }
}
```

## Starfield Transformation (engine/relativity/starfield.go)

```go
package relativity

import "math"

// TransformStarfield transforms an entire starfield for one frame.
// This is the main entry point for CPU-based star rendering.
func (c *RelativisticCamera) TransformStarfield(
    stars []GalaxyStar,
    screenW, screenH int,
    fov float64,       // Horizontal FOV in degrees
    cameraForward Vec3, // Camera look direction in ship frame
) []ScreenStar {
    result := make([]ScreenStar, 0, len(stars))

    // Precompute projection values
    aspect := float64(screenW) / float64(screenH)
    halfFovRad := (fov / 2.0) * math.Pi / 180.0
    tanHalfFov := math.Tan(halfFovRad)

    effectStrength := c.EffectStrength()

    for _, star := range stars {
        // 1. Aberration: transform direction to ship frame
        var nPrime Vec3
        if effectStrength > 0 {
            nGalaxy := star.Dir
            nShip := c.Aberrate(nGalaxy)
            // Blend based on effect strength
            nPrime = lerpVec3(nGalaxy, nShip, effectStrength)
        } else {
            nPrime = star.Dir
        }

        // 2. Check if in front of camera
        forwardDot := nPrime.Dot(cameraForward)
        if forwardDot <= 0 {
            continue // Behind camera, skip
        }

        // 3. Project to screen coordinates
        // Standard perspective projection
        // Assume camera looks along +Z, up is +Y
        screenX, screenY, visible := projectToScreen(
            nPrime, cameraForward, tanHalfFov, aspect, screenW, screenH,
        )
        if !visible {
            continue
        }

        // 4. Doppler factor (computed in galaxy frame)
        D := 1.0
        if effectStrength > 0 {
            D = c.DopplerFactor(star.Dir)
            // Blend toward D=1 based on effect strength
            D = 1.0 + (D-1.0)*effectStrength
        }

        // 5. Color shift
        shiftedColor := star.BaseColor
        if effectStrength > 0 {
            shiftedColor = c.ShiftColor(star.BaseColor, D)
        }

        // 6. Beaming (brightness)
        intensity := star.Intensity
        if effectStrength > 0 {
            beamFactor := c.Beaming(D)
            // Blend toward beamFactor=1 based on effect strength
            beamFactor = 1.0 + (beamFactor-1.0)*effectStrength
            intensity *= beamFactor
        }

        // Skip very dim stars
        if intensity < 0.01 {
            continue
        }

        result = append(result, ScreenStar{
            ScreenX:   screenX,
            ScreenY:   screenY,
            Color:     shiftedColor,
            Intensity: intensity,
            Visible:   true,
        })
    }

    return result
}

// projectToScreen projects a direction vector to screen coordinates.
// Returns screen X, Y, and whether the point is visible.
func projectToScreen(
    dir Vec3,
    forward Vec3,
    tanHalfFov float64,
    aspect float64,
    screenW, screenH int,
) (float64, float64, bool) {
    // Simplified projection assuming forward is +Z
    // In practice, need full camera orientation

    // Compute angles from forward direction
    // X angle (horizontal)
    angleX := math.Atan2(dir.X, dir.Z)
    // Y angle (vertical)
    angleY := math.Atan2(dir.Y, dir.Z)

    // Convert to normalized device coordinates (-1 to +1)
    halfFovRad := math.Atan(tanHalfFov)
    ndcX := angleX / halfFovRad
    ndcY := angleY / (halfFovRad / aspect)

    // Check if in view
    if math.Abs(ndcX) > 1.0 || math.Abs(ndcY) > 1.0 {
        return 0, 0, false
    }

    // Convert to screen coordinates
    screenX := (ndcX*0.5 + 0.5) * float64(screenW)
    screenY := (0.5 - ndcY*0.5) * float64(screenH) // Flip Y

    return screenX, screenY, true
}

// lerpVec3 linearly interpolates between two vectors
func lerpVec3(a, b Vec3, t float64) Vec3 {
    return Vec3{
        X: a.X + (b.X-a.X)*t,
        Y: a.Y + (b.Y-a.Y)*t,
        Z: a.Z + (b.Z-a.Z)*t,
    }
}
```

## Kage Shader for Background (engine/relativity/shader.go)

```go
package relativity

import (
    "github.com/hajimehoshi/ebiten/v2"
)

// KageShaderSource is the Ebiten shader for relativistic background warp.
// This handles nebulae/galactic background as a cubemap/equirectangular texture.
const KageShaderSource = `
//kage:unit pixels
package main

// Uniforms set per frame
var BetaX float   // Ship velocity X
var BetaY float   // Ship velocity Y
var BetaZ float   // Ship velocity Z
var Gamma float   // Lorentz factor
var EffectStrength float // 0-1 blend
var ViewLon float // Current view longitude (radians)
var ViewLat float // Current view latitude (radians)
var FOV float     // Field of view (radians)

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    // Get screen position normalized to -1..+1
    screenW, screenH := imageDstSize()
    ndcX := (dstPos.x / screenW) * 2.0 - 1.0
    ndcY := (dstPos.y / screenH) * 2.0 - 1.0

    // Convert screen position to direction in ship frame
    // Assume looking along +Z
    halfFov := FOV / 2.0
    aspect := screenW / screenH

    dirX := ndcX * tan(halfFov)
    dirY := -ndcY * tan(halfFov) / aspect // Flip Y
    dirZ := 1.0

    // Normalize direction
    len := sqrt(dirX*dirX + dirY*dirY + dirZ*dirZ)
    nPrimeX := dirX / len
    nPrimeY := dirY / len
    nPrimeZ := dirZ / len

    // If no effect, sample directly
    if EffectStrength < 0.001 {
        // Convert direction to equirectangular UV
        lon := atan2(nPrimeX, nPrimeZ) + ViewLon
        lat := asin(nPrimeY) + ViewLat
        u := (lon / (2.0 * 3.14159265)) + 0.5
        v := (lat / 3.14159265) + 0.5
        return imageSrc0At(vec2(u * screenW, v * screenH))
    }

    // Inverse aberration: ship frame → galaxy frame
    // This is complex in shader; simplified version:
    beta := vec3(BetaX, BetaY, BetaZ)
    betaMag := length(beta)

    if betaMag < 0.0001 {
        // No velocity, sample directly
        lon := atan2(nPrimeX, nPrimeZ) + ViewLon
        lat := asin(nPrimeY) + ViewLat
        u := (lon / (2.0 * 3.14159265)) + 0.5
        v := (lat / 3.14159265) + 0.5
        return imageSrc0At(vec2(u * screenW, v * screenH))
    }

    betaHat := beta / betaMag
    nPrime := vec3(nPrimeX, nPrimeY, nPrimeZ)

    // Inverse transform: n = (n' + β) / (1 + β·n')
    // Split into parallel and perpendicular
    nPrimeDotBetaHat := dot(nPrime, betaHat)
    nPrimePar := betaHat * nPrimeDotBetaHat
    nPrimePerp := nPrime - nPrimePar

    denom := 1.0 + dot(beta, nPrime)
    nPar := (nPrimePar + beta) / denom
    nPerp := nPrimePerp / (Gamma * denom)

    n := nPar + nPerp
    n = normalize(n)

    // Blend with unshifted direction
    n = mix(nPrime, n, EffectStrength)

    // Sample texture at galaxy-frame direction
    lon := atan2(n.x, n.z) + ViewLon
    lat := asin(n.y) + ViewLat
    u := (lon / (2.0 * 3.14159265)) + 0.5
    v := (lat / 3.14159265) + 0.5

    baseColor := imageSrc0At(vec2(u * screenW, v * screenH))

    // Doppler shift (computed from galaxy-frame direction)
    D := Gamma * (1.0 - dot(beta, n))
    D = 1.0 + (D - 1.0) * EffectStrength

    // Color shift
    var shiftedColor vec4
    if D < 1.0 {
        // Blue shift
        blueCol := vec4(0.6, 0.7, 1.0, 1.0)
        blend := clamp((1.0 - D), 0.0, 1.0)
        shiftedColor = mix(baseColor, blueCol, blend)
    } else {
        // Red shift
        redCol := vec4(1.0, 0.5, 0.3, 1.0)
        blend := clamp((D - 1.0) * 0.5, 0.0, 1.0)
        shiftedColor = mix(baseColor, redCol, blend)
    }

    // Beaming
    beaming := clamp(pow(D, 3.0), 0.01, 10.0)
    beaming = 1.0 + (beaming - 1.0) * EffectStrength

    return shiftedColor * beaming
}
`

// RelativisticShader wraps the compiled Kage shader
type RelativisticShader struct {
    shader *ebiten.Shader
}

// NewRelativisticShader compiles the SR background shader
func NewRelativisticShader() (*RelativisticShader, error) {
    shader, err := ebiten.NewShader([]byte(KageShaderSource))
    if err != nil {
        return nil, err
    }
    return &RelativisticShader{shader: shader}, nil
}

// SetUniforms sets shader uniforms from camera state
func (s *RelativisticShader) SetUniforms(opts *ebiten.DrawRectShaderOptions, cam *RelativisticCamera, viewLon, viewLat, fov float64) {
    opts.Uniforms = map[string]interface{}{
        "BetaX":          float32(cam.Beta.X),
        "BetaY":          float32(cam.Beta.Y),
        "BetaZ":          float32(cam.Beta.Z),
        "Gamma":          float32(cam.Gamma),
        "EffectStrength": float32(cam.EffectStrength()),
        "ViewLon":        float32(viewLon),
        "ViewLat":        float32(viewLat),
        "FOV":            float32(fov),
    }
}

// Shader returns the underlying Ebiten shader for drawing
func (s *RelativisticShader) Shader() *ebiten.Shader {
    return s.shader
}
```

## Integration with Renderer (engine/render/draw.go modifications)

```go
// Add to Renderer struct:
type Renderer struct {
    assets       *assets.Manager
    anims        *AnimationManager
    lastTick     uint64
    galaxyBg     *ebiten.Image
    galaxyBgLoaded bool

    // NEW: Relativistic rendering
    srShader     *relativity.RelativisticShader
    srCamera     *relativity.RelativisticCamera
}

// Add new DrawCmd type in sim_gen (from AILANG):
// DrawCmdKindSRState - sets relativistic camera state for frame

// In RenderFrame, before processing draw commands:
func (r *Renderer) RenderFrame(screen *ebiten.Image, out sim_gen.FrameOutput) {
    // Check for SR state in frame output
    if out.SRState != nil {
        r.srCamera = relativity.NewRelativisticCamera(
            relativity.Vec3{
                X: out.SRState.BetaX,
                Y: out.SRState.BetaY,
                Z: out.SRState.BetaZ,
            },
            relativity.CameraMode(out.SRState.Mode),
        )
    }

    // ... rest of render loop
}

// Modify drawGalaxyBackground to use SR shader:
func (r *Renderer) drawGalaxyBackground(screen *ebiten.Image, ...) {
    // If SR camera active and in cruise mode, use shader
    if r.srCamera != nil && r.srCamera.Mode == relativity.CameraModeCruise {
        r.drawGalaxyBackgroundSR(screen, ...)
        return
    }

    // ... existing non-SR code
}

func (r *Renderer) drawGalaxyBackgroundSR(screen *ebiten.Image, ...) {
    if r.srShader == nil {
        var err error
        r.srShader, err = relativity.NewRelativisticShader()
        if err != nil {
            // Fallback to non-SR
            return
        }
    }

    opts := &ebiten.DrawRectShaderOptions{}
    opts.Images[0] = r.galaxyBg
    r.srShader.SetUniforms(opts, r.srCamera, viewLon, viewLat, fov)

    screen.DrawRectShader(screenW, screenH, r.srShader.Shader(), opts)
}

// Modify drawStar to transform through SR camera:
func (r *Renderer) drawStar(screen *ebiten.Image, c *sim_gen.DrawCmdStar) {
    // If SR camera active, transform star
    if r.srCamera != nil && r.srCamera.EffectStrength() > 0 {
        r.drawStarSR(screen, c)
        return
    }

    // ... existing non-SR code
}

func (r *Renderer) drawStarSR(screen *ebiten.Image, c *sim_gen.DrawCmdStar) {
    // Convert DrawCmdStar to GalaxyStar
    // Transform through SR camera
    // Draw transformed result
    // ...
}
```

## AILANG Protocol Extension

New fields in FrameOutput from AILANG:

```ailang
type SRState = {
    beta_x: float,
    beta_y: float,
    beta_z: float,
    mode: int          -- 0=Docked, 1=Orbital, 2=Cruise, 3=NearBH
}

type FrameOutput = {
    draw: [DrawCmd],
    debug: [string],
    camera: Camera,
    sr_state: Maybe(SRState)  -- None when not in cruise mode
}
```

## Test Cases (engine/relativity/camera_test.go)

```go
package relativity

import (
    "math"
    "testing"
)

func TestAberration_AtRest(t *testing.T) {
    // At rest, aberration should be identity
    cam := NewRelativisticCamera(Vec3{}, CameraModeCruise)

    n := Vec3{0, 0, 1}.Normalize()
    nPrime := cam.Aberrate(n)

    if !vecNear(n, nPrime, 1e-10) {
        t.Errorf("expected no aberration at rest, got %v", nPrime)
    }
}

func TestAberration_ForwardBeaming(t *testing.T) {
    // At high speed, light from side should shift forward
    beta := Vec3{0, 0, 0.9} // 0.9c toward +Z
    cam := NewRelativisticCamera(beta, CameraModeCruise)

    // Light arriving from +X (perpendicular)
    n := Vec3{1, 0, 0}.Normalize()
    nPrime := cam.Aberrate(n)

    // Should have shifted toward forward (+Z)
    if nPrime.Z <= 0 {
        t.Errorf("expected forward shift, got Z=%f", nPrime.Z)
    }
}

func TestDoppler_BlueShiftForward(t *testing.T) {
    // Moving toward a star should blue-shift it
    beta := Vec3{0, 0, 0.5} // 0.5c toward +Z
    cam := NewRelativisticCamera(beta, CameraModeCruise)

    // Star ahead of us (in direction we're moving)
    n := Vec3{0, 0, 1}.Normalize()
    D := cam.DopplerFactor(n)

    // D should be < 1 (blue-shift)
    if D >= 1.0 {
        t.Errorf("expected D < 1 for forward star, got %f", D)
    }
}

func TestDoppler_RedShiftBackward(t *testing.T) {
    // Moving away from a star should red-shift it
    beta := Vec3{0, 0, 0.5}
    cam := NewRelativisticCamera(beta, CameraModeCruise)

    // Star behind us
    n := Vec3{0, 0, -1}.Normalize()
    D := cam.DopplerFactor(n)

    // D should be > 1 (red-shift)
    if D <= 1.0 {
        t.Errorf("expected D > 1 for backward star, got %f", D)
    }
}

func TestInverseAberration_Roundtrip(t *testing.T) {
    beta := Vec3{0.3, 0.4, 0.5}
    cam := NewRelativisticCamera(beta, CameraModeCruise)

    n := Vec3{0.5, 0.3, 0.8}.Normalize()

    // Forward then inverse should recover original
    nPrime := cam.Aberrate(n)
    nRecovered := cam.InverseAberrate(nPrime)

    if !vecNear(n, nRecovered, 1e-6) {
        t.Errorf("roundtrip failed: %v → %v → %v", n, nPrime, nRecovered)
    }
}

func vecNear(a, b Vec3, tol float64) bool {
    return math.Abs(a.X-b.X) < tol &&
           math.Abs(a.Y-b.Y) < tol &&
           math.Abs(a.Z-b.Z) < tol
}
```

## Success Criteria

### Unit Tests
- [ ] Aberration at rest = identity
- [ ] Aberration shifts sideways light forward at high γ
- [ ] Doppler D < 1 for approaching light
- [ ] Doppler D > 1 for receding light
- [ ] Inverse aberration roundtrips correctly
- [ ] Beaming D³ clamped to configured bounds

### Visual Tests
- [ ] At γ=1.1: Slight visible effect
- [ ] At γ=5: Strong bunching, clear colour shift
- [ ] At γ=15: Dramatic tunnel effect, bright forward cone
- [ ] Smooth transition when entering/leaving cruise mode
- [ ] Background warp matches star warp

### Performance
- [ ] 10,000 stars transformed in < 1ms (CPU)
- [ ] Shader runs at 60fps for 1080p background
- [ ] No visible stuttering during γ changes

## References

- Penrose, R. "The Apparent Shape of a Relativistically Moving Sphere"
- [Real Time Relativity](http://www.anu.edu.au/physics/Searle/) (ANU)
- [A Slower Speed of Light](http://gamelab.mit.edu/games/a-slower-speed-of-light/) (MIT Game Lab)
- Ebiten Kage shader documentation
