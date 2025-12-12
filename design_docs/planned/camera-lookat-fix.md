# Camera LookAt Fix

## Status
- **Status**: Planned
- **Priority**: P1 (blocks camera controls for 3D scenes)
- **Estimated**: 1 day
- **Location**: `engine/tetra/scene.go`

## Game Vision Alignment

| Pillar | Alignment | Notes |
|--------|-----------|-------|
| Time Dilation Consequence | N/A | Infrastructure feature |
| Civilization Simulation | N/A | Infrastructure feature |
| Ship & Crew Life | N/A | Infrastructure feature |
| Hard Sci-Fi Authenticity | N/A | Pure rendering fix |
| **Overall** | ✅ Enabler | Unblocks 3D camera features |

**This is an engine infrastructure fix, not a gameplay feature.**

## Problem Statement

The `Scene.LookAt(x, y, z)` function in `engine/tetra/scene.go` is broken. When called, it causes the camera to look in the wrong direction (often making objects invisible).

### Current Broken Implementation

```go
func (s *Scene) LookAt(x, y, z float64) {
    camPos := s.camera.LocalPosition()
    from := tetra3d.Vector3{X: camPos.X, Y: camPos.Y, Z: camPos.Z}
    to := tetra3d.Vector3{X: float32(x), Y: float32(y), Z: float32(z)}
    up := tetra3d.Vector3{X: 0, Y: 1, Z: 0}

    lookMatrix := tetra3d.NewMatrix4LookAt(from, to, up)
    s.camera.SetLocalRotation(lookMatrix)  // <-- BUG HERE
}
```

### Root Cause

**View Matrix vs Rotation Matrix Confusion**

`tetra3d.NewMatrix4LookAt(from, to, up)` returns a **view matrix**, which transforms world coordinates into camera/view space. This is what you'd pass to a shader for rendering.

However, `SetLocalRotation` expects a **rotation matrix** that represents the camera's local orientation in world space. These are conceptually inverses:

- **View Matrix**: World → Camera Space (transforms objects)
- **Rotation Matrix**: Camera's orientation in World Space (transforms the camera)

Applying a view matrix as a rotation matrix inverts the camera's orientation.

### Evidence

| Test Case | Expected | Actual |
|-----------|----------|--------|
| Camera at (0,0,5), LookAt(0,0,0) | See object at origin | Object invisible (camera looks away) |
| Camera at (5,3,10), LookAt(0,0,0) | Camera faces origin | Camera faces wrong direction |

### Impact

This bug forces workarounds throughout the codebase:

1. **demo-game-saturn**: Rotates Saturn instead of orbiting the camera
2. **demo-game-orbital cruise mode**: Fixed camera path, can't track objects
3. **draw.go planet cache**: Comment says "Don't call LookAt"
4. **dome_renderer.go**: Comment says "Don't use SetSunTarget/LookAt"

## Proposed Fix

### Option A: Invert the Matrix (Recommended)

The rotation matrix is the inverse (transpose for orthonormal matrices) of the view matrix:

```go
func (s *Scene) LookAt(x, y, z float64) {
    camPos := s.camera.LocalPosition()
    from := tetra3d.Vector3{X: camPos.X, Y: camPos.Y, Z: camPos.Z}
    to := tetra3d.Vector3{X: float32(x), Y: float32(y), Z: float32(z)}
    up := tetra3d.Vector3{X: 0, Y: 1, Z: 0}

    viewMatrix := tetra3d.NewMatrix4LookAt(from, to, up)

    // Extract 3x3 rotation part and transpose (inverse for orthonormal)
    rotationMatrix := viewMatrix.Transposed()

    // Zero out translation components (rotation only)
    rotationMatrix.Set(0, 3, 0)
    rotationMatrix.Set(1, 3, 0)
    rotationMatrix.Set(2, 3, 0)

    s.camera.SetLocalRotation(rotationMatrix)
}
```

### Option B: Build Rotation from Basis Vectors

Construct the rotation matrix manually from the look direction:

```go
func (s *Scene) LookAt(x, y, z float64) {
    camPos := s.camera.LocalPosition()

    // Forward vector (camera looks along -Z in local space)
    forward := tetra3d.Vector3{
        X: float32(x) - camPos.X,
        Y: float32(y) - camPos.Y,
        Z: float32(z) - camPos.Z,
    }.Unit()

    // Right vector
    up := tetra3d.Vector3{X: 0, Y: 1, Z: 0}
    right := up.Cross(forward).Unit()

    // Recalculate up to ensure orthonormal
    up = forward.Cross(right).Unit()

    // Build rotation matrix from basis vectors
    rot := tetra3d.NewMatrix4()
    rot.Set(0, 0, right.X)
    rot.Set(1, 0, right.Y)
    rot.Set(2, 0, right.Z)
    rot.Set(0, 1, up.X)
    rot.Set(1, 1, up.Y)
    rot.Set(2, 1, up.Z)
    rot.Set(0, 2, -forward.X)  // Camera looks along -Z
    rot.Set(1, 2, -forward.Y)
    rot.Set(2, 2, -forward.Z)

    s.camera.SetLocalRotation(rot)
}
```

### Option C: Use Tetra3D's Built-in Node.LookAt

Check if Tetra3D's Node has a LookAt method we could use directly:

```go
func (s *Scene) LookAt(x, y, z float64) {
    target := tetra3d.Vector3{X: float32(x), Y: float32(y), Z: float32(z)}
    // If Tetra3D Node has LookAt:
    s.camera.LookAt(target)  // Hypothetical - need to verify API
}
```

## Investigation Required

Before implementing, verify:

1. **Tetra3D Matrix Convention**: Does `NewMatrix4LookAt` follow OpenGL or DirectX convention?
2. **SetLocalRotation Expectation**: What matrix format does it expect?
3. **Node.LookAt Existence**: Does Tetra3D's Node type have a built-in LookAt?
4. **Column-major vs Row-major**: How are matrices stored in Tetra3D?

### Research Tasks

```bash
# Check Tetra3D source for Matrix4 and LookAt:
# https://github.com/SolarLune/Tetra3d

# Check if Camera/Node has built-in LookAt:
grep -r "func.*LookAt" $GOPATH/pkg/mod/github.com/solarlune/tetra3d*

# Check matrix storage order:
grep -r "Set\|Get" tetra3d/matrix4.go
```

## Test Plan

### Demo for Verification

Enhance `cmd/demo-engine-lookat` to test the fix:

```bash
# Test cases to verify:
bin/demo-engine-lookat --mode camera-track     # Camera tracks orbiting planet
bin/demo-engine-lookat --mode orbit-camera     # Camera orbits around fixed object
bin/demo-engine-lookat --mode various-angles   # Camera at different positions
```

### Test Matrix

| Camera Position | LookAt Target | Expected Result |
|-----------------|---------------|-----------------|
| (0, 0, 5) | (0, 0, 0) | Object at origin centered in view |
| (5, 0, 0) | (0, 0, 0) | Object at origin centered in view |
| (0, 5, 0) | (0, 0, 0) | Object at origin centered (top-down) |
| (5, 5, 5) | (0, 0, 0) | Object at origin centered (diagonal) |
| (0, 3, 10) | (0, 0, -5) | Object at -5Z centered in view |

### Success Criteria

- [ ] Camera at any position can LookAt any target
- [ ] Objects at target position appear centered in view
- [ ] No visual artifacts or inverted views
- [ ] demo-game-saturn can use real orbiting camera
- [ ] demo-game-orbital cruise mode can track objects

## Files to Modify

| File | Changes |
|------|---------|
| `engine/tetra/scene.go` | Fix LookAt implementation |
| `cmd/demo-engine-lookat/main.go` | Add more test modes |
| `cmd/demo-game-saturn/main.go` | Replace workaround with real orbit |
| Various files | Remove "don't use LookAt" warnings |

## Cleanup After Fix

Once LookAt works, remove workarounds:

1. **demo-game-saturn**: Use real camera orbit instead of rotating Saturn
2. **draw.go**: Remove "Don't call LookAt" comment
3. **dome_renderer.go**: Enable SetSunTarget if needed
4. **scene.go**: Update function docstring, remove WARNING

## References

- [OpenGL LookAt Matrix Derivation](https://learnopengl.com/Getting-started/Camera)
- [View Matrix vs Model Matrix](https://www.opengl-tutorial.org/beginners-tutorials/tutorial-3-matrices/)
- [Tetra3D GitHub](https://github.com/SolarLune/Tetra3d)
- [Tetra3D Matrix4 Source](https://github.com/SolarLune/Tetra3d/blob/main/matrix.go)
