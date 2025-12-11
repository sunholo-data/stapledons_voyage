# Camera LookAt Fix

**Status**: Planned
**Target**: v0.2.0
**Priority**: P2 - Medium
**Estimated**: 4 hours
**Dependencies**: None

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Engine infrastructure |
| Civilization Simulation | N/A | 0 | Engine infrastructure |
| Philosophical Depth | N/A | 0 | Engine infrastructure |
| Ship & Crew Life | N/A | 0 | Engine infrastructure |
| Legacy Impact | N/A | 0 | Engine infrastructure |
| Hard Sci-Fi Authenticity | N/A | 0 | Engine infrastructure |
| **Net Score** | | **0** | **Decision: Move forward (enabling tech)** |

**Feature type:** Engine/Infrastructure
- This is enabling tech for demos and debugging
- N/A scores are acceptable - no negative impacts

## Problem Statement

The `Scene.LookAt()` function in `engine/tetra/scene.go` does not work correctly for cameras. When called, it causes rendered objects to disappear.

**Current State:**
- `LookAt()` uses `NewMatrix4LookAt()` to create a rotation matrix
- Matrix is applied via `camera.SetLocalRotation()`
- Objects disappear when this is called
- Current workaround: orbit objects around camera instead of camera around objects

**Root Cause Analysis:**

1. **LocalPosition vs WorldPosition**: The current implementation uses `camera.LocalPosition()` but `NewMatrix4LookAt` is documented to work with world positions
2. **Model vs Camera transforms**: `NewMatrix4LookAt` is designed for models pointing at targets, not cameras. Cameras may need the inverse transform.
3. **Tetra3D examples**: All camera examples use `NewMatrix4Rotate` with explicit tilt/rotate angles, not `NewMatrix4LookAt`

**Impact:**
- Affects all demos that need camera tracking (Saturn, planet views, arrival sequence)
- Forces awkward workarounds (orbiting objects instead of camera)
- Makes debugging 3D issues harder

## Goals

**Primary Goal:** Make `Scene.LookAt(x, y, z)` correctly orient the camera to look at a world position.

**Success Metrics:**
- Camera renders objects when LookAt is used
- Camera visibly points at the target position
- Works from any camera position to any target position

## Solution Design

### Overview

Investigate Tetra3D camera orientation handling and implement a working LookAt function. Three possible approaches:

### Approach A: Use WorldPosition + Matrix Inversion

The camera's view matrix is the inverse of its world transform. Try:
```go
func (s *Scene) LookAt(x, y, z float64) {
    camPos := s.camera.WorldPosition()  // Changed from LocalPosition
    from := tetra3d.Vector3{X: camPos.X, Y: camPos.Y, Z: camPos.Z}
    to := tetra3d.Vector3{X: float32(x), Y: float32(y), Z: float32(z)}
    up := tetra3d.WorldUp  // Use Tetra3D's constant

    lookMatrix := tetra3d.NewMatrix4LookAt(from, to, up)
    s.camera.SetLocalRotation(lookMatrix.Inverted())  // Try inverted
}
```

### Approach B: Compute Euler Angles

Convert target direction to rotation angles like Tetra3D examples do:
```go
func (s *Scene) LookAt(x, y, z float64) {
    camPos := s.camera.WorldPosition()
    dx := float32(x) - camPos.X
    dy := float32(y) - camPos.Y
    dz := float32(z) - camPos.Z

    // Calculate yaw (rotation around Y) and pitch (tilt)
    yaw := math32.Atan2(dx, -dz)  // Note: negative Z is forward
    dist := math32.Sqrt(dx*dx + dz*dz)
    pitch := math32.Atan2(dy, dist)

    rotate := tetra3d.NewMatrix4Rotate(0, 1, 0, yaw).Rotated(1, 0, 0, -pitch)
    s.camera.SetLocalRotation(rotate)
}
```

### Approach C: Node Hierarchy (Parent Camera to Target Node)

Use Tetra3D's scene graph - parent camera to a helper node that auto-tracks:
```go
// Create camera pivot/gimbal
pivot := tetra3d.NewNode("camera_pivot")
pivot.AddChildren(s.camera)
s.scene.Root.AddChildren(pivot)

// LookAt sets pivot position and rotation
func (s *Scene) LookAt(x, y, z float64) {
    // Move pivot to camera position, then use standard rotation
}
```

### Implementation Plan

**Phase 1: Investigation** (~1 hour)
- [ ] Test Approach A (WorldPosition + invert) in demo-saturn
- [ ] Test Approach B (Euler angles) if A fails
- [ ] Review Tetra3D camera source code for hints

**Phase 2: Implementation** (~2 hours)
- [ ] Implement working solution in `engine/tetra/scene.go`
- [ ] Add `LookAt` to `engine/view/planet_layer.go` if needed
- [ ] Update demo-saturn to use proper camera orbiting

**Phase 3: Testing** (~1 hour)
- [ ] Test from various camera positions
- [ ] Test with various target positions
- [ ] Verify with demo-saturn screenshot capture

### Files to Modify

**Modified files:**
- `engine/tetra/scene.go` - Fix LookAt function (~20 LOC)
- `cmd/demo-saturn/main.go` - Restore camera orbit instead of object orbit (~30 LOC)

## Examples

### Example 1: Camera Orbiting Saturn

**Before (current workaround):**
```go
// Saturn orbits around fixed camera at origin
g.orbitAngle += *orbitSpeed * dt
saturnX := math.Sin(g.orbitAngle) * *distance
saturnZ := -math.Cos(g.orbitAngle) * *distance
g.saturn.SetPosition(saturnX, saturnY, saturnZ)
g.planetLayer.SetCameraPosition(0, camY, 0)  // Camera at origin
```

**After (proper solution):**
```go
// Camera orbits Saturn at origin
g.orbitAngle += *orbitSpeed * dt
camX := math.Sin(g.orbitAngle) * *distance
camZ := math.Cos(g.orbitAngle) * *distance
g.planetLayer.SetCameraPosition(camX, camY, camZ)
g.planetLayer.LookAt(0, 0, 0)  // Camera looks at Saturn
g.saturn.SetPosition(0, 0, 0)  // Saturn at origin
```

## Success Criteria

- [ ] `LookAt(x, y, z)` makes camera face the target point
- [ ] Objects remain visible after LookAt is called
- [ ] demo-saturn works with camera orbiting Saturn (not reverse)
- [ ] Works with elevated camera positions (Y offset)

## Testing Strategy

**Unit tests:**
- Not practical for visual rendering - use demos

**Manual testing:**
- Run demo-saturn with camera orbit
- Capture screenshots at multiple orbit angles
- Verify Saturn stays centered in frame

**Regression testing:**
- Ensure demo-planet-view still works
- Ensure demo-sr-flyby still works

## Non-Goals

**Not in this feature:**
- Smooth camera interpolation/animation - separate feature
- Camera constraints (limits on angles) - separate feature
- Camera shake effects - separate feature

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| All approaches fail | High | Keep current workaround, file issue with Tetra3D |
| Performance impact from matrix operations | Low | Matrix operations are fast, not a bottleneck |
| Breaks existing demos | Medium | Test all demos before merging |

## References

- [Tetra3D matrix.go:799](https://github.com/solarlune/tetra3d/blob/main/matrix.go#L799) - `NewMatrix4LookAt` documentation
- [Tetra3D examples/common.go:250](https://github.com/solarlune/tetra3d/blob/main/examples/common.go#L250) - Camera rotation example
- `engine/tetra/scene.go` - Current broken implementation
- `cmd/demo-saturn/main.go` - Current workaround implementation

## Future Work

- Smooth camera transitions between LookAt targets
- Camera follow modes (track moving objects)
- First-person and third-person camera presets

---

**Document created**: 2025-12-08
**Last updated**: 2025-12-08
