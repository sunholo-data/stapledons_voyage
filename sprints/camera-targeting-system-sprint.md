# Sprint: Camera & Targeting System

**Design Doc**: [camera-targeting-system.md](../design_docs/planned/camera-targeting-system.md)
**Duration**: 1 day (focused sprint)
**Priority**: P0 - Blocks 3D planet rendering in dome

## Goal

Fix the camera/targeting system so Tetra3D 3D planets render correctly in the dome viewport, and establish a foundation for game-wide targeting.

## Success Criteria

- [ ] `demo-engine-lookat` proves all LookAt modes work
- [ ] Bridge demo shows 3D textured planets through dome
- [ ] Saturn's rings render with proper alpha
- [ ] No regression in demo-engine-solar
- [ ] Screenshots captured for all test cases

## Phase 1: Diagnostic Demo (~2 hours)

Create `cmd/demo-engine-lookat/main.go` to isolate and test camera/targeting behaviors.

### Tasks

- [ ] **1.1** Create demo-engine-lookat binary structure
  - New file: `cmd/demo-engine-lookat/main.go`
  - Support multiple test modes via --mode flag

- [ ] **1.2** Test Mode: "sun-lookat"
  - Single planet at origin
  - Sun light using LookAt to point at planet
  - Verify planet is illuminated on correct side
  - Screenshot: `out/screenshots/lookat-sun.png`

- [ ] **1.3** Test Mode: "sun-position"
  - Same setup but sun using position-only (no LookAt)
  - Compare illumination with sun-lookat mode
  - Screenshot: `out/screenshots/lookat-sun-position.png`

- [ ] **1.4** Test Mode: "camera-track"
  - Planet orbiting around camera
  - Camera using LookAt to track planet
  - Verify camera follows correctly
  - Screenshot: `out/screenshots/lookat-camera-track.png`

- [ ] **1.5** Test Mode: "dome-replica"
  - Replicate exact dome renderer setup in isolation
  - Same planets, same positions, same lighting
  - This will prove if issue is in dome or in PlanetLayer
  - Screenshot: `out/screenshots/lookat-dome-replica.png`

### Files
- `cmd/demo-engine-lookat/main.go` (~300 LOC)

## Phase 2: Fix LookAt Implementation (~2 hours)

Based on diagnostics, fix the broken implementation.

### Tasks

- [ ] **2.1** Investigate Tetra3D's built-in LookAt
  - Check if `tetra3d.INode` has LookAt method
  - Test using built-in vs custom implementation

- [ ] **2.2** Fix SunLight.LookAt in `engine/tetra/lighting.go`
  - Current: Uses `NewMatrix4LookAt` + `SetLocalRotation`
  - Fix: Use Tetra3D's node.LookAt() if available
  - Or: Calculate rotation matrix correctly

- [ ] **2.3** Fix Scene.LookAt in `engine/tetra/scene.go` (if needed)
  - Same approach as SunLight fix

- [ ] **2.4** Verify demo-engine-lookat passes all modes
  - Re-run all test modes
  - Capture new screenshots showing fix works

### Files Modified
- `engine/tetra/lighting.go` (~30 LOC changed)
- `engine/tetra/scene.go` (~20 LOC changed)

## Phase 3: Dome Integration (~1 hour)

Apply fix to dome renderer and verify full integration.

### Tasks

- [ ] **3.1** Update dome_renderer if needed
  - Apply any necessary changes from fix
  - Ensure SetCameraFromState uses corrected code

- [ ] **3.2** Test bridge demo
  - Run `demo-game-bridge --screenshot 60`
  - Verify 3D planets visible
  - Screenshot: `out/screenshots/bridge-planets.png`

- [ ] **3.3** Test cruise animation
  - Run longer to see planets pass by
  - Screenshot at multiple frames: 60, 120, 180
  - Screenshots: `out/screenshots/bridge-cruise-*.png`

### Files Modified
- `engine/view/dome_renderer.go` (~10 LOC)

## Phase 4: Documentation & Cleanup (~30 min)

### Tasks

- [ ] **4.1** Update design doc status
  - Mark Phase 1 (Tetra3D fix) as complete
  - Note: Phase 2-4 (AILANG types, full system) deferred

- [ ] **4.2** Update tetra3d-planet-rendering sprint
  - Mark relevant tasks complete

- [ ] **4.3** Collect all screenshots
  - Ensure all test screenshots captured
  - Document what each shows

- [ ] **4.4** Clean up debug code
  - Remove any temporary logging
  - Comment out verbose debug output

## Test Plan

### Demo Tests (Automated Screenshots)

| Test | Command | Expected |
|------|---------|----------|
| Sun LookAt | `demo-engine-lookat --mode sun-lookat --screenshot 30` | Planet lit on facing side |
| Sun Position | `demo-engine-lookat --mode sun-position --screenshot 30` | Same as above |
| Camera Track | `demo-engine-lookat --mode camera-track --screenshot 30` | Planet centered |
| Dome Replica | `demo-engine-lookat --mode dome-replica --screenshot 30` | Planets visible |
| Bridge Full | `demo-game-bridge --screenshot 60` | 3D planets in dome |

### Regression Tests

| Test | Command | Expected |
|------|---------|----------|
| demo-engine-solar | `demo-engine-solar --mode full --screenshot 30` | Still works |
| demo-engine-tetra | `demo-engine-tetra --screenshot 30` | Still works |

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Tetra3D LookAt fundamentally broken | Fall back to position-only (no targeting) |
| Issue is not in LookAt but elsewhere | dome-replica test will isolate true cause |
| Fix breaks other demos | Run all demos before/after |

## AILANG Note

**UPDATE 2025-12-11**: Sprint pivoted to AILANG-first approach as per project architecture.

### AILANG Changes Completed
- [x] Added `SpaceBg`, `Planets3D`, `BubbleArc` DrawCmd variants to `sim/protocol.ail`
- [x] Updated `sim/bridge.ail` to emit 3D commands at correct layers
- [x] Added layer constants: `layerPlanets3D()=4`, `layerBubbleArc3D()=5`

### RESOLVED: AILANG Issue #30
**Status**: Fixed by AILANG team. Sprint completed!

**Completed Steps**:
1. ✅ AILANG team fixed issue #30 (std/list Option type)
2. ✅ `make sim` regenerated sim_gen with new DrawCmd types
3. ✅ Updated Go renderer to handle SpaceBg, Planets3D, BubbleArc
4. ✅ Simplified BridgeView - AILANG controls Z-ordering

## Completion Checklist

- [ ] All Phase 1 tasks complete
- [ ] All Phase 2 tasks complete
- [ ] All Phase 3 tasks complete
- [ ] All Phase 4 tasks complete
- [ ] All screenshots captured
- [ ] All tests passing
- [ ] Design doc updated

---

**Sprint created**: 2025-12-11
**Sprint started**: 2025-12-11
**Sprint completed**: 2025-12-11 (AILANG-first approach)
