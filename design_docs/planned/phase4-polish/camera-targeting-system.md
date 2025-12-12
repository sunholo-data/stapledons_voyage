# Camera & Targeting System

**Status**: Planned
**Target**: v0.2.0
**Priority**: P0 - High (blocks multiple features)
**Estimated**: 2 days
**Dependencies**: None

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Infrastructure |
| Civilization Simulation | + | +1 | Enables galaxy view navigation |
| Philosophical Depth | N/A | 0 | Infrastructure |
| Ship & Crew Life | + | +1 | Enables crew perspective shots |
| Legacy Impact | N/A | 0 | Infrastructure |
| Hard Sci-Fi Authenticity | + | +1 | Enables accurate astronomical views |
| **Net Score** | | **+3** | **Decision: Move forward** |

**Feature type:** Engine/Infrastructure
- This is core enabling tech used by multiple game systems

**Reference:** See [game-vision.md](../../../docs/game-vision.md)

## Problem Statement

The game needs a robust camera/targeting system that can:

1. **Point cameras at targets** - For 3D planet views, galaxy navigation, cinematic sequences
2. **Track moving objects** - Follow player, NPCs, ships, celestial objects
3. **Orient directional elements** - Lights toward planets, turrets toward targets, NPCs facing conversation partners

**Current State:**
- Each system implements its own ad-hoc targeting (Tetra3D LookAt, AILANG facing, etc.)
- Tetra3D's current LookAt implementation appears broken
- No unified system for camera control across game modes

**Use Cases That Need This:**

| Use Case | System | Description |
|----------|--------|-------------|
| Solar system cruise | Tetra3D | Camera flying through 3D planet scene |
| Galaxy map navigation | 2D Camera | Pan/zoom across starmap |
| Ship interior | Isometric | Follow player around bridge |
| Dialogue | UI | Portrait facing, camera focus |
| Combat | 2D/3D | Turrets tracking targets |
| Autopilot | Navigation | Ship pointing toward destination |

**Impact:**
- Tetra3D 3D planets not rendering in dome (blocks key visual feature)
- No camera follow for player in isometric view
- No smooth camera transitions between game modes

## Goals

**Primary Goal:** Create a unified camera/targeting system that works across 2D, isometric, and 3D contexts.

**Success Metrics:**
- LookAt works for Tetra3D cameras and lights
- Camera smoothly follows targets in all view modes
- Direction/facing calculations are mathematically correct
- AILANG can control targeting (state), Go handles math (rendering)

## Solution Design

### Overview

Two-layer architecture following AILANG-first principles:

1. **AILANG State Layer** - What is targeting what (target IDs, offsets)
2. **Go Render Layer** - Calculate actual angles/matrices, apply to rendering

### Core Concepts

```
Target: { type: Point | Entity | Direction, ... }
Camera: { position, target, mode: Follow | Fixed | Orbit }
Facing: Direction an entity should face
```

### AILANG Types (sim/targeting.ail)

```ailang
-- What can be targeted
type Target =
    | TargetPoint(float, float, float)        -- World coordinates
    | TargetEntity(string)                     -- Entity ID to track
    | TargetDirection(float, float, float)     -- Unit vector direction

-- Camera behavior modes
type CameraMode =
    | CameraFixed                              -- Static position
    | CameraFollow(float)                      -- Follow target with smoothing
    | CameraOrbit(float, float)                -- Orbit at radius, height

-- Camera state (owned by AILANG)
type CameraState = {
    position: (float, float, float),
    target: Target,
    mode: CameraMode,
    fov: float
}

-- Facing state for entities
type Facing = {
    target: Target,
    turnSpeed: float    -- Radians per second
}
```

### Go Implementation (engine/camera/)

```go
// LookAt calculates rotation to point from->to
func LookAt(from, to Vector3, up Vector3) Matrix4

// CalculateFacing returns angle to face target
func CalculateFacing(position, target Vector2) float64

// SmoothFollow interpolates camera position
func SmoothFollow(current, target Vector3, smoothing, dt float64) Vector3
```

### Phase 1: Fix Tetra3D LookAt (~4 hours)

The immediate blocker. Fix the current broken implementation.

**Root Cause Hypotheses:**
1. `NewMatrix4LookAt` returns view matrix, not rotation matrix
2. `SetLocalRotation` expects different matrix format
3. Tetra3D has built-in node LookAt we should use instead

**Tasks:**
- [ ] Create `cmd/demo-engine-lookat/main.go` diagnostic demo
- [ ] Test hypothesis 1: Extract rotation from view matrix
- [ ] Test hypothesis 2: Use Tetra3D's node.LookAt() if available
- [ ] Fix `SunLight.LookAt()` in `engine/tetra/lighting.go`
- [ ] Fix `Scene.LookAt()` if needed
- [ ] Verify dome renderer shows 3D planets

### Phase 2: AILANG Targeting Types (~4 hours)

Add targeting to AILANG so game logic can control what looks at what.

**Tasks:**
- [ ] Add `Target` type to `sim/types.ail`
- [ ] Add `CameraState` type
- [ ] Add `Facing` type for entities
- [ ] Add `stepCamera` function for camera updates
- [ ] Add `stepFacing` function for entity facing

### Phase 3: Go Camera System (~4 hours)

Generic camera utilities used by all rendering modes.

**Tasks:**
- [ ] Create `engine/camera/lookat.go` with math functions
- [ ] Create `engine/camera/smooth.go` for interpolation
- [ ] Integrate with Tetra3D wrapper
- [ ] Integrate with isometric camera
- [ ] Add camera debug visualization

### Phase 4: Integration (~4 hours)

Wire everything together.

**Tasks:**
- [ ] Dome renderer uses new camera system
- [ ] Bridge view camera follows player
- [ ] Add camera controls (debug mode)
- [ ] Screenshot tests for all camera modes

### Files to Modify/Create

**New files:**
- `sim/targeting.ail` - AILANG targeting types (~50 LOC)
- `engine/camera/lookat.go` - LookAt math (~100 LOC)
- `engine/camera/smooth.go` - Smooth follow (~50 LOC)
- `cmd/demo-engine-lookat/main.go` - Diagnostic demo (~200 LOC)

**Modified files:**
- `engine/tetra/lighting.go` - Fix SunLight.LookAt (~20 LOC)
- `engine/tetra/scene.go` - Use new camera math (~30 LOC)
- `engine/view/dome_renderer.go` - Use new system (~20 LOC)
- `sim_gen/` - Generated from new AILANG types

## Examples

### Example 1: 3D Sun Light Targeting Planet

**AILANG (state):**
```ailang
let sunTarget = TargetPoint(0.0, 0.0, -50.0)  -- Point toward planets
```

**Go (rendering):**
```go
// In dome_renderer.go
func (d *DomeRenderer) updateLighting(state *sim_gen.DomeState) {
    if state.SunTarget != nil {
        target := state.SunTarget.ToVector3()
        d.planetLayer.SetSunTarget(target.X, target.Y, target.Z)
    }
}

// In lighting.go (fixed)
func (s *SunLight) LookAt(x, y, z float64) {
    target := tetra3d.NewVector3(float32(x), float32(y), float32(z))
    s.light.LookAt(target, false)  // Use Tetra3D's built-in LookAt
}
```

### Example 2: Camera Following Player

**AILANG (state):**
```ailang
let cameraState = {
    position: (0.0, 0.0, 10.0),
    target: TargetEntity("player"),
    mode: CameraFollow(0.1),  -- 10% smoothing
    fov: 60.0
}
```

**Go (rendering):**
```go
func updateCamera(camera *Camera, state CameraState, entities map[string]Entity, dt float64) {
    targetPos := resolveTarget(state.Target, entities)

    switch mode := state.Mode.(type) {
    case CameraFollow:
        camera.Position = SmoothFollow(camera.Position, targetPos, mode.Smoothing, dt)
    case CameraOrbit:
        camera.Position = OrbitPosition(targetPos, mode.Radius, mode.Height, time)
    }

    camera.LookAt(targetPos)
}
```

### Example 3: NPC Facing During Dialogue

**AILANG:**
```ailang
pure func startDialogue(npc: NPC, player: Entity) -> NPC {
    { npc | facing: { target: TargetEntity(player.id), turnSpeed: 3.14 } }
}
```

## Success Criteria

- [ ] Tetra3D LookAt works (sun lights illuminate correct planet faces)
- [ ] demo-engine-lookat shows all test cases passing
- [ ] Bridge demo shows 3D textured planets
- [ ] Camera smoothly follows player in isometric mode (if implemented)
- [ ] AILANG can specify camera targets
- [ ] No regression in existing demos

## Testing Strategy

**Demo tests:**
- `demo-engine-lookat` - Comprehensive LookAt testing
- `demo-engine-solar` - Regression test
- `demo-game-bridge` - Integration test

**Visual tests:**
- Screenshot comparisons before/after fix
- Verify lighting direction on planets
- Verify camera tracking accuracy

## Non-Goals

**Not in this feature:**
- Cinematic camera paths (keyframe animation)
- Split-screen / multiple viewports
- VR camera support
- Post-processing effects tied to camera

## Timeline

**Day 1** (~8 hours):
- Phase 1: Fix Tetra3D LookAt (4 hours)
- Phase 2: AILANG types (4 hours)

**Day 2** (~8 hours):
- Phase 3: Go camera system (4 hours)
- Phase 4: Integration (4 hours)

**Total: ~16 hours across 2 days**

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Tetra3D LookAt unfixable | High | Workaround: manual rotation matrices |
| Breaking existing camera | High | Comprehensive regression testing |
| AILANG codegen issues | Medium | Can fall back to Go-only initially |

## References

- [tetra3d-planet-rendering.md](./next/tetra3d-planet-rendering.md) - Dependent feature
- [engine-capabilities.md](../reference/engine-capabilities.md) - Current engine features
- [Tetra3D Node API](https://pkg.go.dev/github.com/solarlune/tetra3d#Node) - Built-in LookAt
- [3D Math Primer](http://www.3dgep.com/understanding-the-view-matrix/) - View matrix explanation

## Future Work

- Cinematic camera system (keyframes, easing)
- Camera shake effects
- Picture-in-picture views
- First-person mode for crew members
- Observatory telescope mode (zoom to distant objects)

---

**Document created**: 2025-12-11
**Last updated**: 2025-12-11
