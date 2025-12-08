# Sprint: Tetra3D Integration

## Overview

| Field | Value |
|-------|-------|
| Sprint ID | `tetra3d-v1` |
| Design Doc | [02-tetra3d-integration.md](../../design_docs/planned/next/02-tetra3d-integration.md) |
| Estimated Duration | 3 days |
| Type | Engine-only (Go) |
| Dependencies | View System (completed) |
| Enables | 3D planet rendering, solar system view |

## Goal

Integrate Tetra3D library to enable 3D sphere rendering with textures, creating the foundation for planet visuals in the arrival sequence.

## Sprint Structure

### Milestone 1: Basic Tetra3D Setup (Day 1)
**Goal**: Verify Tetra3D works with Ebitengine

**Tasks**:
1. Add Tetra3D dependency (`go get github.com/solarlune/tetra3d`)
2. Create `engine/tetra/scene.go` - Scene wrapper with camera
3. Create `cmd/demo-tetra/main.go` - Minimal demo
4. Render solid-color icosphere (no texture yet)
5. Verify 60fps maintained

**Acceptance Criteria**:
- [ ] `go build ./...` succeeds with Tetra3D
- [ ] Demo displays rotating sphere on screen
- [ ] Camera controls work (position, FOV)
- [ ] Screenshot captures sphere correctly

**LOC Estimate**: ~250

### Milestone 2: Textures & Lighting (Day 2)
**Goal**: Render textured planet with sun illumination

**Tasks**:
1. Create `engine/tetra/planet.go` - Planet mesh with texture
2. Create `engine/tetra/lighting.go` - Directional light (sun)
3. Download NASA Earth texture (4K or 2K)
4. Apply texture to icosphere
5. Position sun to create terminator line (day/night)
6. Add rotation animation

**Acceptance Criteria**:
- [ ] Earth texture loads correctly
- [ ] Terminator line visible (day/night boundary)
- [ ] Planet rotates smoothly
- [ ] Screenshot shows textured, lit planet

**LOC Estimate**: ~300

### Milestone 3: View Integration & Shaders (Day 3)
**Goal**: Composite 3D into view system with SR/GR effects

**Tasks**:
1. Create `engine/view/planet_view.go` or extend SpaceView
2. Composite Tetra3D render buffer into view layer
3. Test SR shader on 3D output (Doppler shift)
4. Test GR shader on 3D output (lensing)
5. Add `--demo-3d` flag to demo-view command
6. Performance profiling (target: <5ms for 3D render)

**Acceptance Criteria**:
- [ ] 3D planet composites over starfield background
- [ ] SR shader produces visible Doppler shift on planet
- [ ] GR shader produces visible lensing on planet
- [ ] 60fps maintained with planet + shaders
- [ ] Screenshots verify all visual states

**LOC Estimate**: ~350

## Visual Verification Checkpoints

Per sprint-executor skill requirements, screenshots are MANDATORY:

| Checkpoint | Frame | What to Verify |
|------------|-------|----------------|
| Basic sphere | 30 | Solid color icosphere renders |
| Textured planet | 30 | Earth texture applied correctly |
| Terminator line | 30 | Day/night boundary visible |
| Rotation | 60 | Planet has rotated from initial |
| With starfield | 30 | Planet composites over stars |
| SR effect | 30 | Blue shift visible at 0.3c |
| GR effect | 30 | Lensing distortion visible |

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Tetra3D API changed | Low | Pin version, check examples |
| Performance issues | Medium | Limit to 1 planet, LOD fallback |
| Texture memory | Low | Start with 2K, upgrade if needed |
| Shader compatibility | Low | Tetra3D renders to ebiten.Image |

## Demo Command Specification

```bash
# Basic 3D demo
./bin/demo-view --demo-3d

# With planet selection
./bin/demo-view --demo-3d --planet earth

# With velocity (SR effect)
./bin/demo-view --demo-3d --velocity 0.3

# Screenshot for testing
./bin/demo-view --demo-3d --screenshot 30 --output out/tetra-demo.png
```

## Technical Notes

### Texture Sources
- NASA Visible Earth: https://visibleearth.nasa.gov/
- Blue Marble: 4K Earth texture
- Start with bundled test texture, upgrade to NASA later

### Icosphere Subdivisions
- 3 subdivisions = 162 vertices (good for N64 aesthetic)
- 4 subdivisions = 642 vertices (smoother)
- Start with 3, adjust based on visual quality

### Coordinate System
- Tetra3D uses Y-up, right-handed
- Camera looks down -Z by default
- Planet at origin, camera positioned on +Z axis

## Success Metrics

| Metric | Target |
|--------|--------|
| Build time | <30s |
| Frame time (3D only) | <5ms |
| Frame time (full) | <16ms (60fps) |
| Texture load time | <500ms |
| Total LOC | ~900 |

## Post-Sprint

After completion:
1. Move design doc to `implemented/`
2. Update engine-capabilities.md with Tetra3D section
3. Proceed to 03-3d-sphere-planets.md for multiple planets
