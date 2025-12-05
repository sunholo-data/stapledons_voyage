# Sprint: GR Visual Effects (M-GR-VIS) ✅ COMPLETE

## Summary

Implement General Relativity visual effects (gravitational lensing, redshift) near massive objects (black holes, neutron stars, white dwarfs). Builds on existing SR effects pipeline.

**Status:** ✅ COMPLETE (2025-12-05)
**Duration:** 3 days (focused engine work)
**Sprint ID:** `gr-visual-effects`
**Dependencies:** SR Effects (implemented), Shader System (implemented)
**Risk Level:** Medium (shader complexity)
**Type:** Mock-friendly (primarily Go/Ebiten, minimal AILANG)

## Current Status

### Prerequisites
- [x] SR Effects implemented (`engine/shader/sr.go`, `sr_warp.kage`)
- [x] Shader pipeline working (`engine/shader/effects.go`)
- [x] Screenshot capture for testing (`engine/screenshot/`)
- [x] Design doc complete (`design_docs/planned/gr-visual-mechanics.md`)

### Messages Check
```bash
ailang messages list --unread
```
- [ ] Messages checked

## Scope

**In scope:**
- GR context struct in Go (mock sim_gen)
- Gravitational lensing shader (Schwarzschild approximation)
- Gravitational redshift shader
- Integration with effects pipeline
- Demo mode controls (F5 for GR toggle)
- Screenshot tests

**Out of scope (future):**
- AILANG GRContext types (deferred to AILANG integration sprint)
- Kerr (rotating) black holes
- Accretion disk rendering
- HUD indicators

## Milestones

### Phase 1: GR Context & Data Structures (Day 1 AM) ✅ COMPLETE
**Goal:** Define Go types for GR context and shader uniforms

**Tasks:**
- [x] Create `engine/relativity/gr_context.go` with GRContext struct
- [x] Add MassiveObjectKind enum (BlackHole, NeutronStar, WhiteDwarf)
- [x] Add GRDangerLevel enum (None, Subtle, Strong, Extreme)
- [x] Implement `ComputeGRContext(shipPos, massiveObject)` function
- [x] Add `GRShaderUniforms` struct for shader communication
- [x] Write unit tests for GR physics calculations

**Acceptance:**
- [x] `go test ./engine/relativity/...` passes (10 tests)
- [x] GR context correctly computes Φ, time dilation, redshift factor

### Phase 2: GR Lensing Shader (Day 1 PM - Day 2 AM) ✅ COMPLETE
**Goal:** Implement gravitational lensing visual effect

**Tasks:**
- [x] Create `engine/shader/shaders/gr_lensing.kage`
- [x] Implement radial distortion based on Schwarzschild approximation
- [x] Add photon sphere ring effect at 1.5 r_s
- [x] Handle edge cases (object off-screen, very close approach)
- [x] Create `engine/shader/gr.go` wrapper (similar to sr.go)
- [x] Integrate with Effects pipeline

**Acceptance:**
- [x] Visible background distortion when GR enabled
- [x] Distortion increases with proximity (higher Φ)
- [x] Smooth falloff at effect boundaries

### Phase 3: GR Redshift Shader (Day 2 PM) ✅ COMPLETE
**Goal:** Implement gravitational redshift color effect

**Tasks:**
- [x] Create `engine/shader/shaders/gr_redshift.kage`
- [x] Implement radial redshift falloff from object center
- [x] Combine with SR Doppler (runs in sequence)
- [x] Add to effects pipeline after lensing

**Acceptance:**
- [x] Color shift toward red near massive object center
- [x] Smooth gradient falloff
- [x] Combines correctly with SR effects (GR → SR → Bloom → Pipeline)

### Phase 4: Demo Mode & Testing (Day 3) ✅ COMPLETE
**Goal:** Demo controls and visual test suite

**Tasks:**
- [x] Add F3 toggle for GR effects in demo mode
- [x] Add Shift+F3 to cycle GR intensity (Subtle/Strong/Extreme)
- [x] Update overlay text with GR info
- [x] Integrate into effects Apply() pipeline
- [ ] Generate golden file screenshots at each danger level (deferred)
- [ ] Performance profiling (deferred)

**Acceptance:**
- [x] F3 toggles GR lensing effect
- [x] F9 overlay shows GR status
- [x] All code compiles and passes tests

## Implementation Details

### GR Physics (from design doc)

**Key quantities:**
- Schwarzschild radius: `r_s = 2GM/c²`
- Dimensionless potential: `Φ = r_s / (2r)`
- Time dilation: `dτ/dt = sqrt(1 - r_s/r)`
- Redshift factor: `z = 1 / sqrt(1 - r_s/r)`

**Danger level thresholds:**
| Φ Range | Level | Effects |
|---------|-------|---------|
| < 1e-4 | None | GR disabled |
| 1e-4 to 1e-3 | Subtle | Light shading |
| 1e-3 to 1e-2 | Strong | Visible lensing |
| ≥ 1e-2 | Extreme | Heavy distortion |

### Shader Uniforms

```go
type GRShaderUniforms struct {
    Enabled         bool
    ObjectKind      int       // 0=BH, 1=NS, 2=WD
    Rs              float32   // Schwarzschild radius (screen units)
    Distance        float32   // ship→object distance
    Phi             float32   // dimensionless potential
    ScreenCenter    [2]float32 // screen position of object
    MaxEffectRadius float32   // effect bounds
    LensStrength    float32   // tunable parameter
}
```

### Lensing Approximation (Kage pseudo-code)

```kage
// For each pixel:
d := position.xy - ScreenCenter
b := length(d)  // impact parameter

if b > MaxEffectRadius {
    return texture(src, texCoord)
}

// Deflection angle (weak-field approximation)
alpha := LensStrength * Rs / max(b, 0.001)

// Radial stretching
dir := normalize(d)
warpedB := b + alpha * b
sampleCoord := ScreenCenter + dir * warpedB

return texture(src, screenToUV(sampleCoord))
```

## Files to Create

| File | Purpose | Est. LOC |
|------|---------|----------|
| `engine/relativity/gr_context.go` | GR context types and computation | ~150 |
| `engine/relativity/gr_context_test.go` | Unit tests | ~100 |
| `engine/shader/gr.go` | GR shader wrapper | ~120 |
| `engine/shader/shaders/gr_lensing.kage` | Lensing effect | ~80 |
| `engine/shader/shaders/gr_redshift.kage` | Redshift overlay | ~50 |

## Files to Modify

| File | Change |
|------|--------|
| `engine/shader/effects.go` | Add GRWarp to Effects struct |
| `engine/shader/manager.go` | Register GR shaders |
| `engine/screenshot/screenshot.go` | Add GR config options |
| `engine/screenshot/demo_scene.go` | Add massive object to demo |
| `cmd/game/main.go` | Add F5 GR toggle |

## Success Metrics

### Visual
- [x] Lensing visible at Φ > 0.01
- [x] Photon sphere ring effect at 1.5 r_s
- [x] Smooth redshift gradient
- [x] GR + SR combine correctly (effects pipeline order)

### Technical
- [x] All tests pass (10 GR context tests)
- [ ] Performance < 3ms GPU at 1080p (deferred - needs profiling)
- [x] No shader compilation errors

### Demo
- [x] F3 toggles effect
- [ ] Golden files generated for each danger level (deferred)

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Lensing looks wrong | Follow exact Schwarzschild formula; tune empirically |
| Performance issues | Limit effect radius; add quality settings |
| Edge artifacts | Clamp values; add smooth falloff at boundaries |

## Notes

- This is a **mock-friendly sprint** - no AILANG compiler needed
- Uses same pattern as SR effects (sr.go → sr_warp.kage)
- GR is post-process that runs BEFORE SR stage (GR → SR → Bloom → Pipeline)
- AILANG integration (GRContext in FrameOutput) deferred to future sprint

## Implementation Summary

**Files Created:**
- [engine/relativity/gr_context.go](engine/relativity/gr_context.go) - GR context types and computation (~200 LOC)
- [engine/relativity/gr_context_test.go](engine/relativity/gr_context_test.go) - 10 unit tests (~200 LOC)
- [engine/shader/gr.go](engine/shader/gr.go) - GRWarp wrapper (~180 LOC)
- [engine/shader/shaders/gr_lensing.kage](engine/shader/shaders/gr_lensing.kage) - Lensing shader (~100 LOC)
- [engine/shader/shaders/gr_redshift.kage](engine/shader/shaders/gr_redshift.kage) - Redshift shader (~90 LOC)

**Files Modified:**
- [engine/shader/effects.go](engine/shader/effects.go) - Added GRWarp integration
- [engine/shader/manager.go](engine/shader/manager.go) - Registered GR shaders

**Demo Controls:**
- F3: Toggle GR Warp
- Shift+F3: Cycle GR intensity (Subtle/Strong/Extreme)
- F9: Show effects overlay

---

**Created:** 2025-12-05
**Completed:** 2025-12-05
**Design Doc:** [gr-visual-mechanics.md](../design_docs/planned/gr-visual-mechanics.md)
