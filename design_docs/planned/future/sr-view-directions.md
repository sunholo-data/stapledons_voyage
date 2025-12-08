# SR View Directions - Head-Turn Visual Effects

**Status**: Planned
**Target**: v0.2.0
**Priority**: P2 (Enhancement to existing SR system)
**Estimated**: 3-4 days
**Dependencies**: Current SR shader (implemented), Tetra3D scene (implemented)

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | 0 | 0 | Not directly related to time choices |
| Civilization Simulation | 0 | 0 | Not related to galaxy simulation |
| Philosophical Depth | + | +1 | Enhances sense of being "trapped" at relativistic speeds |
| Ship & Crew Life | + | +1 | Crew would naturally look around during travel |
| Legacy Impact | 0 | 0 | No direct impact on ending |
| Hard Sci-Fi Authenticity | ++ | +2 | **Critical** - SR effects must be accurate in all directions |
| **Net Score** | | **+4** | **Decision: Move forward** |

**Feature type:** Engine (enabling tech for authentic SR visuals)

**Reference:** See [game-vision.md](../../../docs/game-vision.md)

## Problem Statement

**The Current Bug:**
When player changes view direction (forward/left/right/behind), the scene content stays the same but SR shader parameters change. This is fundamentally wrong.

**What Happens Now:**
1. Scene renders with camera facing **forward** (toward planets)
2. Player presses S to "look behind"
3. SR shader applies **redshift** to the forward-rendered scene
4. Result: Empty/dark screen because there's nothing in the forward scene that corresponds to "behind"

**What Should Happen:**
1. Player presses S to "look behind"
2. Camera **rotates** to face opposite direction of travel
3. Scene re-renders showing what's **actually behind** (planets we've passed)
4. SR shader applies **redshift** to this rear-view scene

**Current State:**
- Forward view: Works correctly (blueshift, aberration visible)
- Side views: Shows forward scene with incorrect SR parameters → broken
- Behind view: Shows forward scene with redshift → appears empty/dark

**Impact:**
- Breaks immersion during arrival sequences
- Scientifically inaccurate (a hard sci-fi dealbreaker)
- Players can't watch planets recede as they pass them

## Physics Background

### Relativistic Aberration & Doppler Shift

When moving at relativistic speeds (significant fraction of c), light is affected by:

1. **Doppler Shift**: Wavelength/color changes based on relative motion
   - Looking forward (toward velocity): **Blueshift** (D > 1)
   - Looking perpendicular: **Neutral** (D ≈ 1)
   - Looking behind (opposite velocity): **Redshift** (D < 1)

2. **Aberration**: Light appears to come from different directions
   - Stars "bunch forward" in direction of motion
   - At 0.9c, the visible sky compresses into a forward cone

3. **Relativistic Beaming (D³)**: Intensity varies with direction
   - Forward: Brighter (light compressed)
   - Behind: Dimmer (light stretched)

### Key Insight: View Direction vs Velocity Direction

The **ViewAngle** parameter represents the angle between:
- The direction you're **looking**
- The direction you're **moving**

This is independent of the 3D camera orientation in the scene.

| View | Camera Faces | ViewAngle | Expected Effect |
|------|--------------|-----------|-----------------|
| Forward | +Z (direction of travel) | 0 | Blueshift, bright |
| Left | +X | π/2 | Neutral, moderate |
| Right | -X | π/2 | Neutral, moderate |
| Behind | -Z (opposite travel) | π | Redshift, dim |

## Goals

**Primary Goal:** Allow player to look in any direction during relativistic travel and see physically accurate SR effects on whatever is actually in that direction.

**Success Metrics:**
- Pass a planet at 0.7c, look forward → see blueshift
- Press S to look behind → see the planet receding with redshift
- All four directions (F/L/R/B) show correct scene content with correct SR effects

## Solution Design

### Overview

Decouple **camera direction** from **velocity direction**:

1. **Velocity** defines motion through space (what's approaching vs receding)
2. **Camera** rotates to show different views (F/L/R/B)
3. **SR Shader** takes both into account:
   - Samples from the camera's rendered scene
   - Applies Doppler/beaming based on angle between view direction and velocity

### Architecture

**Components:**

1. **ViewDirection Enum** (exists): Forward/Left/Right/Behind
2. **Camera Rotation** (new): Rotate Tetra3D camera based on ViewDirection
3. **SR Shader** (modify): ViewAngle already works, just needs correct input
4. **Scene Re-render** (implicit): Camera rotation means scene shows different content

### Current vs Proposed Data Flow

**Current (Broken):**
```
Input → ViewDirection → SR Shader ViewAngle (but camera never moves)
                      ↓
Camera always faces forward → Scene shows forward view
                           ↓
SR Shader applies wrong effects to wrong content
```

**Proposed (Correct):**
```
Input → ViewDirection → Camera Rotation → Scene shows actual view direction
                      ↓
                      → SR Shader ViewAngle → Correct effects for that view
```

### Implementation Plan

**Phase 1: Camera Rotation** (~4 hours)
- [ ] Add `SetCameraRotation(yaw float64)` to PlanetLayer
- [ ] Modify Tetra3D Scene to support camera yaw
- [ ] ViewDirection maps to camera yaw: F=0°, L=90°, R=-90°, B=180°
- [ ] Test: Pressing S should show what's behind (without SR)

**Phase 2: Velocity Vector Tracking** (~2 hours)
- [ ] Add `velocityDir` field to track ship's motion direction
- [ ] Velocity direction is independent of camera direction
- [ ] Update demo-arrival to track velocity direction properly

**Phase 3: SR Integration** (~4 hours)
- [ ] Compute ViewAngle = angle between camera direction and velocity direction
- [ ] SR shader already handles ViewAngle correctly
- [ ] Test all four directions at different velocities

**Phase 4: Starfield Background** (~2 hours)
- [ ] SpaceBackground needs to rotate with camera
- [ ] Galaxy image should rotate appropriately
- [ ] Stars should show correct aberration for view direction

### Files to Modify/Create

**Modified files:**
- `engine/view/planet_layer.go` - Add camera rotation (~20 LOC)
- `engine/tetra/scene.go` - Support camera yaw rotation (~15 LOC)
- `cmd/demo-arrival/main.go` - Integrate camera rotation with view direction (~30 LOC)
- `engine/view/background/space.go` - Rotate with camera (~20 LOC)

**No new files required** - this enhances existing architecture.

## Examples

### Example 1: Passing a Planet

**Before (Current Broken Behavior):**
```
Frame 1: Flying toward planet at 0.7c
         View: FORWARD, Camera: Forward, SR: Blueshift
         Result: See planet with blueshift ✓

Frame 2: Now passing the planet
         View: BEHIND, Camera: Still Forward!, SR: Redshift
         Result: See nothing (planet not in forward view) ✗
```

**After (Correct Behavior):**
```
Frame 1: Flying toward planet at 0.7c
         View: FORWARD, Camera: Forward, SR: Blueshift
         Result: See planet with blueshift ✓

Frame 2: Now passing the planet
         View: BEHIND, Camera: Rotated 180°, SR: Redshift
         Result: See planet receding with redshift ✓
```

### Example 2: Side View During Flyby

**Before:**
```
View: LEFT, Camera: Forward, ViewAngle: π/2
Result: See forward scene with wrong SR mapping → broken
```

**After:**
```
View: LEFT, Camera: Rotated 90°, ViewAngle: π/2
Result: See what's actually to our left with correct neutral/mixed SR
```

## Success Criteria

- [ ] Forward view: Blueshift, shows what's ahead
- [ ] Left view: Shows what's to port, neutral/mixed Doppler
- [ ] Right view: Shows what's to starboard, neutral/mixed Doppler
- [ ] Behind view: Redshift, shows what's behind (planets we've passed)
- [ ] Smooth camera transitions between views
- [ ] Starfield rotates consistently with 3D scene
- [ ] Physics remains accurate per SR formulas
- [ ] Demo works at velocities from 0.1c to 0.95c

## Testing Strategy

**Unit tests:**
- ViewAngle calculation: given camera yaw and velocity direction, verify angle

**Integration tests:**
- Screenshot comparison at each view direction
- Verify planet visibility when it should be in view

**Manual testing:**
- Fly past a planet, switch views, verify you can "track" it
- Check color shifts match expected physics

## Non-Goals

**Not in this feature:**
- Smooth animated camera rotation (just snap for now) - Add later
- 360° panoramic rendering - Too expensive
- Rear-view mirrors / multiple simultaneous views - Different feature
- VR head tracking - Future consideration

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Camera rotation breaks existing demos | Medium | Add rotation only when ViewDirection changes |
| Performance impact from scene re-render | Low | Scene already renders each frame; rotation is cheap |
| Starfield doesn't match 3D scene rotation | Medium | Sync SpaceBackground rotation with camera yaw |
| ViewAngle calculation edge cases | Low | Use atan2 for robust angle computation |

## Physics Validation Checklist

- [x] Doppler formula: D = γ(1 + β·cos(θ)) ✓
- [x] Aberration formula: cos(θ) = (cos(θ') + β) / (1 + β·cos(θ')) ✓
- [x] Beaming: I' = I × D³ ✓
- [ ] View direction independence from velocity direction (THIS FIX)
- [ ] Correct angle computation for all quadrants

## References

- [sr_warp.kage](../../../engine/shader/shaders/sr_warp.kage) - Current SR shader (works, just needs correct inputs)
- [demo-arrival/main.go](../../../cmd/demo-arrival/main.go) - Demo showing the bug
- [engine-capabilities.md](../reference/engine-capabilities.md) - Available SR/GR effects
- Wikipedia: [Relativistic Doppler Effect](https://en.wikipedia.org/wiki/Relativistic_Doppler_effect)
- Wikipedia: [Relativistic Aberration](https://en.wikipedia.org/wiki/Aberration_(astronomy)#Relativistic_aberration)

## Current Working State (for reference)

The forward view works correctly as of 2025-12-08:
- SR shader properly handles ViewAngle=0
- Doppler shift, aberration, and beaming all work
- Galaxy background integrates well

Screenshot: `out/sr-test-v2.png` shows correct forward blueshift at 0.7c

## Future Work

- Animated camera rotation with easing
- Smooth Doppler shift transitions during rotation
- VR head tracking integration
- Cockpit frame that rotates with view (if we add ship interior)

---

**Document created**: 2025-12-08
**Last updated**: 2025-12-08
