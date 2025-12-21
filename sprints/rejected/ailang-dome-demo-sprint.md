# Sprint: AILANG Dome Demo - Player Movement with Solar System View

**Design Doc**: [design_docs/planned/ailang-dome-demo.md](../design_docs/planned/ailang-dome-demo.md)
**Type**: AILANG Demo (validates dual coordinate architecture)
**Estimated**: 7 hours (~1 day)
**Sprint Start**: 2025-12-19
**Status**: Planning

## Goal

Create AILANG demo with player walking inside dome viewing solar system, demonstrating dual coordinate system (player in meters ship-local, solar in AU galactic). All game logic in AILANG, rendering via existing DrawCmds.

## Success Criteria

- [ ] `ailang check sim/dome_demo.ail` passes
- [ ] `make sim` compiles without errors
- [ ] `go run ./cmd/demo-ailang-dome` runs smoothly
- [ ] Player walks around disc with WASD
- [ ] Solar system orbits visibly through dome
- [ ] Dual coordinates work (player ≠ solar positions)
- [ ] Frame rate > 30 FPS

## Pre-Sprint Checks

- [x] Design doc complete and approved
- [x] solar_demo.ail exists with solar system bodies
- [x] demo-engine-dome reference implementation exists
- [x] No blocking AILANG messages
- [ ] Review `ailang prompt` for current syntax

## Implementation Tasks

### Phase 1: Create AILANG Module (~3 hours)

- [ ] **Task 1.1**: Define core types in `sim/dome_demo.ail`
  - PlayerState (posX/Y/Z, yaw, pitch, walkSpeed)
  - DomeConfig (radius, floorY, visibility flags)
  - DomeState (player, dome, solarBodies, timeScale)
  - Run `ailang check` after types defined

- [ ] **Task 1.2**: Implement initialization functions
  - `initPlayer() -> PlayerState`
  - `initDomeConfig() -> DomeConfig`
  - `initDomeDemo() -> DomeState` (reuses `getAllSolarSystemBodies()`)
  - Test with `ailang run --entry initDomeDemo sim/dome_demo.ail`

- [ ] **Task 1.3**: Implement player movement logic
  - `updatePlayerPosition(player, input, domeRadius) -> PlayerState`
  - Handle WASD input (forward/back/strafe)
  - Clamp to dome boundary (circular constraint)
  - Run `ailang check` to verify

- [ ] **Task 1.4**: Implement DrawCmd generation
  - `drawFloorDisc() -> DrawCmd` (Circle for floor)
  - `drawFloorProps() -> [DrawCmd]` (markers at radii)
  - `solarPlanetToDrawCmd(p) -> DrawCmd` (map AU to meters)
  - `renderDomeDemo(state) -> [DrawCmd]` (combine all)

- [ ] **Task 1.5**: Implement step function
  - `stepDomeDemo(state, input) -> (DomeState, FrameOutput)`
  - Update player position
  - Update solar orbital phases
  - Generate DrawCmds and FrameOutput
  - Run `ailang check` for final validation

- [ ] **Task 1.6**: Add helper functions
  - `concatLists()` for combining DrawCmd lists
  - `buildCamera()` for Camera from player state
  - `buildDefaultRelativity/Lighting/LOD()` contexts
  - Ensure all exports have `export` keyword

### Phase 2: Generate Go Code (~1 hour)

- [ ] **Task 2.1**: Compile AILANG to Go
  - Run `make sim`
  - Check for codegen errors in output
  - If errors in sim_gen/, report via `ailang messages send user "..."`

- [ ] **Task 2.2**: Verify generated code quality
  - Run `.claude/skills/sprint-executor/scripts/check_codegen_quality.sh`
  - Check for excessive nesting (>20 chars indentation)
  - Check for too many closures (>10 consecutive)
  - Document any issues in sprint notes

- [ ] **Task 2.3**: Verify exports exist
  - Check `sim_gen/dome_demo.go` has `InitDomeDemo()`
  - Check `sim_gen/dome_demo.go` has `StepDomeDemo()`
  - Check types exported: `DomeState`, `PlayerState`, `DomeConfig`

### Phase 3: Create Go Host (~2 hours)

- [ ] **Task 3.1**: Create Go host file structure
  - Create `cmd/demo-ailang-dome/main.go`
  - Import sim_gen, engine/render, ebiten packages
  - Define Game struct with `state *sim_gen.DomeState`

- [ ] **Task 3.2**: Implement effect handlers initialization
  - Initialize Debug, Rand, Clock handlers
  - Call `sim_gen.Init(handlers)` BEFORE InitDomeDemo()
  - Verify no panics on startup

- [ ] **Task 3.3**: Implement ebiten.Game interface
  - `Update() error` - captures input, calls StepDomeDemo()
  - `Draw(screen) error` - calls StepDomeDemo(), renders output
  - `Layout(w, h) (int, int)` - returns screen dimensions

- [ ] **Task 3.4**: Implement input capture
  - Create `captureInput() sim_gen.FrameInput`
  - Map ebiten WASD keys to FrameInput.wPressed, etc.
  - Map mouse movement to yaw/pitch deltas
  - Add Escape key to quit

- [ ] **Task 3.5**: Wire up rendering
  - Call `render.RenderFrame(screen, output)` in Draw()
  - Ensure DrawCmds from AILANG render correctly
  - Test with empty state first

- [ ] **Task 3.6**: Add to Makefile
  - Add `demo-ailang-dome` target
  - Build to `bin/demo-ailang-dome`
  - Test `make demo-ailang-dome` builds successfully

### Phase 4: Test & Polish (~1 hour)

- [ ] **Task 4.1**: Manual testing - Player movement
  - Run `go run ./cmd/demo-ailang-dome`
  - Test W key - moves forward
  - Test S key - moves backward
  - Test A key - strafes left
  - Test D key - strafes right
  - Verify clamping to dome boundary

- [ ] **Task 4.2**: Manual testing - Solar system
  - Verify planets visible through dome
  - Verify planets orbit over time
  - Verify floor props render correctly
  - Verify floor disc renders at correct position

- [ ] **Task 4.3**: Manual testing - Dual coordinates
  - Walk around - verify floor stays fixed (not following player)
  - Watch solar system - verify positions independent of player
  - Verify no coordinate mixing (player meters ≠ solar AU)

- [ ] **Task 4.4**: Add HUD display (optional)
  - Show player position (X, Z) in meters
  - Show player yaw/pitch in degrees
  - Show frame count/FPS
  - Use `ebitenutil.DebugPrint()`

- [ ] **Task 4.5**: Update documentation
  - Add controls to file header comment
  - Document WASD for movement
  - Document ESC to quit
  - Add to demos.md reference

## Files Modified

### Created:
- `sim/dome_demo.ail` (~300 LOC)
- `cmd/demo-ailang-dome/main.go` (~150 LOC)
- `sprints/ailang-dome-demo-sprint.md` (this file)

### Modified:
- `Makefile` - Add demo-ailang-dome target (~5 LOC)
- `design_docs/reference/demos.md` - Add dome demo entry (~10 LOC)

### Generated:
- `sim_gen/dome_demo.go` (auto-generated by ailang compiler)

## Testing Checklist

After implementation, verify:

- [ ] **AILANG Compilation**: `ailang check sim/dome_demo.ail` ✓
- [ ] **Go Compilation**: `make sim && make demo-ailang-dome` ✓
- [ ] **Runtime**: `go run ./cmd/demo-ailang-dome` starts ✓
- [ ] **Movement**: WASD controls work smoothly ✓
- [ ] **Rendering**: Solar system visible and orbits ✓
- [ ] **Coordinates**: Floor fixed, player moves relative ✓
- [ ] **Performance**: Frame rate > 30 FPS ✓

## Known AILANG Constraints

| Constraint | Impact | Workaround |
|------------|--------|------------|
| List operations O(n) | Medium | 60+ planets may be slow, profile and optimize |
| No mutable state | Low | Use functional updates (already designed) |
| Recursion depth limits | Low | List operations shallow (<100 depth) |
| Module imports | None | Using `import sim/solar_demo` (should work in v0.5.0) |

## AILANG Feedback Checkpoints

### Before Starting:
```bash
# Check inbox for any updates
ailang messages list --unread

# Review current syntax
ailang prompt | grep -A10 "import\|export\|pure func"
```

### During Implementation:
- Report any type errors that are unclear
- Report any codegen issues in sim_gen/
- Document workarounds used

### After Completion:
```bash
# Send DX feedback
~/.claude/skills/ailang-feedback/scripts/send_feedback.sh dx \
  "Sprint DX: AILANG Dome Demo" \
  "Experience creating dual-coordinate demo in AILANG. Positives: [list]. Friction: [list]." \
  --from stapledons_voyage
```

## Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| DrawCmd API insufficient | Low | High | Validated with demo-engine-dome already |
| 60+ planets too slow | Medium | Medium | Profile first, can reduce planet count if needed |
| Module import fails | Low | Medium | v0.5.0 should have working imports, fallback to inline types |
| Input handling mismatch | Low | Low | Use existing FrameInput protocol |

## Post-Sprint

After completing this sprint:

- [ ] Move design doc from `planned/` to `implemented/v0_5_0/`
- [ ] Create demo video/GIF showing dual coordinates
- [ ] Document pattern for future AILANG features
- [ ] Consider: Extend with 3D rendering (Tetra3D DrawCmds)
- [ ] Consider: Add SR/GR effects when ship moves

## Notes

- **AILANG-only sprint** - No Go game logic, only rendering
- Validates architecture for observation deck gameplay
- Reuses existing solar_demo.ail (60+ bodies)
- All DrawCmds already exist in engine
- This is a **demo**, not production feature

---

**Sprint Created**: 2025-12-19
**Last Updated**: 2025-12-19
