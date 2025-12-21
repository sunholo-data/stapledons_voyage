# Sprint: Bubble Ship Dome System - Dual Coordinate Spaces

**Design Doc**: [design_docs/planned/bubble-ship-dome-system.md](../design_docs/planned/bubble-ship-dome-system.md)
**Type**: Engine Refactoring (Go only, no AILANG)
**Estimated**: 6 hours (less than 1 day)
**Sprint Start**: 2025-12-19
**Status**: In Progress

## Goal

Fix demo-engine-dome to separate ship-local coordinates (player movement via WASD) from galactic coordinates (ship velocity through space), enabling realistic observation deck rendering where walking doesn't move distant stars.

## Success Criteria

- [ ] Player movement (WASD) does not affect sky sphere star positions
- [ ] Ship velocity (V key) affects SR effects on sky sphere
- [ ] Ship heading (H key) rotates which stars are visible
- [ ] Interior geometry stays fixed in ship-local space
- [ ] 3D parallax visible for nearby objects (<5 ly)
- [ ] HUD shows both coordinate systems clearly

## Pre-Sprint Checks

- [x] Design doc complete and approved
- [x] demo-engine-dome currently broken (lines 696-707 move entire universe)
- [x] demo-engine-lod has working space travel system (reference)
- [x] demo-game-interior has working player movement (reference)
- [x] No AILANG work needed (engine-level only)

## Implementation Tasks

### Phase 1: Separate Variables (~2 hours)

- [ ] **Task 1.1**: Add `shipVelocity float64` field to Game struct
  - Default: 0.0 (stationary)
  - Range: 0.0 to 0.99c

- [ ] **Task 1.2**: Add `shipHeading` vector field to Game struct
  - Type: struct with X, Y, Z float64
  - Default: (0, 0, -1) pointing toward Sol (-Z direction)

- [ ] **Task 1.3**: Add `shipHeadingIndex int` for cycling through preset headings
  - Headings: North (-Z), South (+Z), East (+X), West (-X), Up (+Y), Down (-Y)

- [ ] **Task 1.4**: Add V key handler to cycle ship velocity
  - Press V: cycle through [0.0, 0.2, 0.4, 0.6, 0.8, 0.99]c
  - Update status message on change

- [ ] **Task 1.5**: Add H key handler to cycle ship heading
  - Press H: cycle through 6 cardinal directions
  - Update status message on change

- [ ] **Task 1.6**: Remove broken coordinate coupling (lines 698-707)
  - REMOVE: `g.platform.SetLocalPosition(g.camX, g.camY-2.0, g.camZ)`
  - REMOVE: struts following camera in loop
  - KEEP: `g.skySphere.SetPosition(g.camX, g.camY, g.camZ)` (geometry follows player)

### Phase 2: Fix Coordinate Systems (~3 hours)

- [ ] **Task 2.1**: Refactor `updateStarTexture()` to use ship params
  - **Before**: Uses camera look direction (`g.yaw`, `g.pitch`)
  - **After**: Uses `shipHeading` and `shipVelocity`
  - Update ViewParams to use ship direction, not player view
  - Player can look anywhere without affecting SR effects

- [ ] **Task 2.2**: Set platform to fixed ship-local position
  - Platform stays at (0, -2.0, 0) in ship coordinates
  - Player moves relative to platform, not vice versa

- [ ] **Task 2.3**: Set struts to fixed ship-local position
  - Struts stay at (0, 0, 0) in ship coordinates
  - Create helper: `rebuildDomeStructure()` for radius changes only

- [ ] **Task 2.4**: Fix LOD star positioning
  - Stars positioned in galactic coordinates (their actual positions)
  - 3D planets rendered relative to camera for parallax
  - Verify nearby stars (<5 ly) show correct parallax when walking

- [ ] **Task 2.5**: Update sky sphere texture regeneration
  - Only regenerate when ship velocity/heading changes (V or H keys)
  - NOT when player looks around (mouse) or walks (WASD)
  - Cache texture until ship parameters change

### Phase 3: Test & Polish (~1 hour)

- [ ] **Task 3.1**: Manual testing - Player movement isolation
  - Walk 10m forward (W key) → verify stars don't shift
  - Walk left/right (A/D keys) → verify platform stays fixed
  - Move up/down (Q/E keys) → verify struts stay fixed

- [ ] **Task 3.2**: Manual testing - Ship velocity effects
  - Press V to 0.8c → verify aberration (stars shift forward)
  - Press V to 0.99c → verify extreme Doppler (colors shift)
  - Press V to 0.0c → verify effects disappear

- [ ] **Task 3.3**: Manual testing - Ship heading effects
  - Press H to cycle through N/S/E/W/Up/Down
  - Verify different stars visible in each direction
  - Verify SR effects apply in travel direction, not view direction

- [ ] **Task 3.4**: Update HUD display
  - Add "PLAYER POSITION (Ship-Local):" section showing camX, camY, camZ
  - Add "SHIP STATUS (Galactic):" section showing velocity, heading
  - Clear visual separation between the two coordinate systems
  - Show current heading as cardinal direction name

- [ ] **Task 3.5**: Update file header controls documentation
  - Document V key: Cycle ship velocity (SR effects)
  - Document H key: Cycle ship heading (direction of travel)
  - Clarify WASD: Move player inside ship (local coordinates)
  - Clarify Mouse: Look around (doesn't affect ship)

## Files Modified

### `cmd/demo-engine-dome/main.go`
- **Lines modified**: ~60 LOC
- **Changes**:
  - Game struct: Add shipVelocity, shipHeading, shipHeadingIndex fields
  - Update(): Add V/H key handlers, remove lines 702-707 (broken coupling)
  - updateStarTexture(): Use shipHeading instead of camera look direction
  - Draw(): Update HUD to show dual coordinate systems
  - File header: Update controls documentation

## Testing Checklist

After implementation, verify:

- [ ] **Isolation Test**: Walk around → stars stay fixed ✓
- [ ] **SR Test**: Change velocity → aberration/Doppler effects ✓
- [ ] **Heading Test**: Change heading → different stars visible ✓
- [ ] **Parallax Test**: Walk around → nearby stars show parallax ✓
- [ ] **Mouse Test**: Look around → doesn't affect SR or star positions ✓
- [ ] **Reset Test**: Press R → returns to initial state ✓
- [ ] **HUD Test**: Both coordinate systems clearly displayed ✓

## Known Risks

| Risk | Mitigation |
|------|-----------|
| LOD stars still coupled to camera | Review lines 716-777, ensure stars use galactic positions |
| SR shader performance | Already working in demo-engine-lod, just parameter changes |
| Confusing UX | Clear HUD labels + separate key bindings (WASD vs V/H) |

## Post-Sprint

After completing this sprint:

- [ ] Move design doc from `planned/` to `implemented/v0_X_X/`
- [ ] Create demo video showing dual coordinate systems
- [ ] Document findings for future AILANG implementation
- [ ] Consider: Create AILANG version in new demo (separate sprint)

## Notes

- This is **engine-level only** (Go code)
- Future work: Port to AILANG for game integration
- The fix enables observation deck gameplay features
- Critical for hard sci-fi authenticity (walking ≠ universe shifting)

---

**Sprint Created**: 2025-12-19
**Last Updated**: 2025-12-19
