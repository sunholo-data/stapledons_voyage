# Sprint: Bubble Ship Dome Demo

**Design Doc:** [design_docs/planned/bubble-ship-dome.md](../design_docs/planned/bubble-ship-dome.md)
**Duration:** 1 day (focused demo)
**Goal:** Create `cmd/demo-dome` showing a functional hemisphere dome with starfield

## Scope

This sprint creates a **working demo** of the dome concept. Full SR/GR integration is deferred to the main implementation sprint.

### In Scope
- Hemisphere mesh generation in Tetra3D
- Cubemap starfield rendering (6 faces)
- Basic dome demo with player inside
- Stars visible through dome surface

### Out of Scope (deferred)
- AILANG Dome3D DrawCmd (use Go directly for demo)
- Multiple dome types

### Now In Scope (added Phase 6)
- SR aberration/Doppler on dome (V key cycles velocity)
- GR lensing on dome (G key cycles gravity phi)

## Tasks

### Phase 1: Hemisphere Mesh (~2 hours)
- [x] Create `engine/tetra/dome.go` with hemisphere mesh generation
- [x] Parameters: radius, rings, segments, arcAngle
- [x] UV mapping for texture projection
- [x] Test mesh renders correctly

### Phase 2: Equirectangular Starfield (~2 hours)
- [x] Extend `engine/render/space_view.go` with `RenderEquirectangular()` method
- [x] Full-sky equirectangular projection (simpler than cubemap)
- [x] SR velocity support for aberration/Doppler
- [x] Procedural fallback when no star catalog

### Phase 3: Dome Demo (~2 hours)
- [x] Create `cmd/demo-dome/main.go`
- [x] Create dome mesh centered at origin
- [x] Apply equirectangular texture to dome interior
- [x] Player camera inside dome looking around
- [x] Mouse look controls

### Phase 4: Polish (~1 hour)
- [x] Add HUD showing dome parameters
- [x] Add controls to adjust dome size (+/-)
- [x] Add velocity cycling (V key) for SR effects
- [x] Screenshot support (--screenshot flag)
- [x] Add direction cycling (D key) for side-facing domes
- [x] Rename to `cmd/demo-engine-dome` (pure Go engine demo)

### Phase 5: Starmap & LOD Integration (~2 hours)
- [x] Fix star catalog path (`assets/data/starmap/stars.json`)
- [x] Load 3,802 real stars from CNS5 catalog
- [x] Add LOD manager for nearby celestial objects
- [x] Add Tetra3D planets for Full3D LOD tier (Sol + nearby stars)
- [x] Add camera movement (WASD, Q/E)
- [x] Update HUD with star count and LOD stats
- [x] Toggle LOD objects (P key)

### Phase 6: SR/GR Effects & Orientation (~1 hour)
- [x] Add observation platform (floor) for spatial orientation
- [x] Add dome struts (meridians, rings, base) for dome boundary visualization
- [x] Add F key to toggle struts visibility
- [x] Add G key to cycle GR phi (0 → 0.1 → 0.2 → 0.3 → 0.4 → 0)
- [x] Add R key to reset camera position (looking at Sol)
- [x] Update HUD with GR phi and struts status
- [x] Platform and struts follow camera (always inside dome)

## Files Created

| File | Purpose | LOC |
|------|---------|-----|
| `engine/tetra/dome.go` | Hemisphere mesh generation with direction support | ~210 |
| `cmd/demo-engine-dome/main.go` | Demo with starmap, LOD, SR/GR, platform, struts | ~800 |

## Files to Modify

| File | Changes | LOC |
|------|---------|-----|
| `engine/render/space_view.go` | Add RenderCubemap() | ~80 |
| `Makefile` | Add demo-dome target | ~5 |

## Technical Notes

### Hemisphere Mesh Generation

```go
// Generate vertices for hemisphere
for ring := 0; ring <= rings; ring++ {
    phi := (float64(ring) / float64(rings)) * arcAngle  // 0 to π for hemisphere
    for seg := 0; seg <= segments; seg++ {
        theta := (float64(seg) / float64(segments)) * 2 * math.Pi
        x := radius * math.Sin(phi) * math.Cos(theta)
        y := radius * math.Cos(phi)  // Y is up
        z := radius * math.Sin(phi) * math.Sin(theta)
        // Add vertex with UV for cubemap sampling
    }
}
```

### Cubemap Face Directions

| Face | Direction | Camera Look |
|------|-----------|-------------|
| +X | Right | (1, 0, 0) |
| -X | Left | (-1, 0, 0) |
| +Y | Up | (0, 1, 0) |
| -Y | Down | (0, -1, 0) |
| +Z | Forward | (0, 0, 1) |
| -Z | Back | (0, 0, -1) |

## Success Criteria

- [x] Demo launches without errors
- [x] Hemisphere dome visible from inside
- [x] Stars render on dome surface
- [x] No visible seams (equirectangular mapping)
- [x] Player can look around with mouse
- [x] 60 FPS maintained

## AILANG Feedback

After sprint, evaluate:
- Is Dome3D DrawCmd needed, or can we use existing Window3D creatively?
- What parameters should the AILANG API expose?

**Conclusion:** A dedicated Dome3D DrawCmd is recommended for proper integration. The demo proves the concept works with Tetra3D hemisphere meshes and equirectangular textures.

---

**Created:** 2025-12-19
**Completed:** 2025-12-19
**Status:** COMPLETE
