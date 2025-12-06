# Sprint: Relativistic Visual Effects Demo

**Status**: Implemented
**Sprint ID**: `sr-visual-demo`
**Duration**: 2 days (focused engine work)
**Goal**: Demonstrate SR optical effects (aberration, Doppler, beaming) at various velocities

## Scope Analysis

### Work Breakdown

| Component | Location | Effort | Notes |
|-----------|----------|--------|-------|
| SR Math Functions | `engine/relativity/` | 2h | Pure Go, well-defined formulas |
| SR Background Shader | `engine/shader/shaders/sr_*.kage` | 4h | Kage shader with uniform beta/gamma |
| Pipeline Integration | `engine/shader/effects.go` | 1h | Add to existing effect pipeline |
| Demo Scene | `engine/screenshot/sr_demo.go` | 2h | Extend existing demo infrastructure |
| Screenshot Generation | CLI flags | 1h | `--velocity 0.9` flag |
| Testing & Tuning | Manual | 2h | Verify visuals look correct |

**Total: ~12 hours (2 focused days)**

### AILANG Impact: Minimal

For the **demo**, we bypass AILANG entirely:
- Engine directly sets velocity/gamma uniforms
- No changes to `sim/protocol.ail` needed yet

**Future AILANG work** (not this sprint):
- Add `velocity: Vec3` and `gamma: float` to Camera type
- Compute gamma from ship velocity in step function

### Why Engine-First?

1. **Fast iteration** - shader tweaks don't need AILANG recompile
2. **Testable in isolation** - demo scene with preset velocities
3. **No AILANG blockers** - Vec3 type, float math all work in Go
4. **Proves the concept** - visual validation before integration

---

## Day 1: Core Implementation

### Morning: SR Math Package (~2h)

**Task 1.1**: Create `engine/relativity/transform.go`
- [ ] Vec3 type with basic operations (dot, normalize, scale, add, sub)
- [ ] `Gamma(beta float64) float64` - Lorentz factor
- [ ] `DopplerFactor(beta Vec3, direction Vec3) float64`
- [ ] `TransformDirection(n Vec3, beta Vec3, gamma float64) Vec3` - aberration
- [ ] Unit tests for known cases (gamma=1, gamma=2, etc.)

**Task 1.2**: Create `engine/relativity/color.go`
- [ ] `ShiftColorTemperature(baseTemp, dopplerFactor float64) float64`
- [ ] `TemperatureToRGB(temp float64) (r, g, b uint8)` - Planckian approximation
- [ ] `BeamBrightness(intensity, dopplerFactor float64) float64` - I' ~ D^3

### Afternoon: SR Shader (~4h)

**Task 1.3**: Create `engine/shader/shaders/sr_warp.kage`
```kage
// Uniforms:
// BetaX, BetaY, BetaZ: velocity components (units of c)
// Gamma: Lorentz factor
// ViewDir: camera look direction (for sky sphere mapping)
```
- [ ] Implement inverse aberration (screen pixel → galaxy direction)
- [ ] Sample background at transformed coordinate
- [ ] Apply Doppler color shift
- [ ] Apply beaming brightness

**Task 1.4**: Register shader in `engine/shader/manager.go`
- [ ] Add "sr_warp" to shader list
- [ ] Preload in manager.Preload()

---

## Day 2: Integration & Demo

### Morning: Pipeline Integration (~2h)

**Task 2.1**: Add SR effect to `engine/shader/effects.go`
- [ ] Add `sr *SREffect` field to Effects struct
- [ ] Create `NewSREffect(manager)` with velocity/gamma params
- [ ] `SetVelocity(beta Vec3)` and `SetGamma(gamma float64)`
- [ ] Apply SR warp before other post-processing

**Task 2.2**: Create demo infrastructure
- [ ] Add `--velocity` flag to CLI (0.0 to 0.99)
- [ ] Auto-compute gamma from velocity
- [ ] Add to screenshot config

### Afternoon: Demo Generation (~2h)

**Task 2.3**: Generate SR demo screenshots
```bash
# Generate at different velocities
./bin/game -screenshot 1 -demo-scene -velocity 0.0 -output out/sr/v0.0.png
./bin/game -screenshot 1 -demo-scene -velocity 0.5 -output out/sr/v0.5.png
./bin/game -screenshot 1 -demo-scene -velocity 0.9 -output out/sr/v0.9.png
./bin/game -screenshot 1 -demo-scene -velocity 0.99 -output out/sr/v0.99.png
```

**Task 2.4**: Visual validation
- [ ] v=0: Normal starfield (no transform)
- [ ] v=0.5: Slight forward clustering
- [ ] v=0.9: Clear tunnel effect, blue forward
- [ ] v=0.99: Dramatic searchlight, red/black rear

### Evening: Polish (~2h)

**Task 2.5**: Tune parameters
- [ ] Adjust beaming curve (D^3 may be too aggressive)
- [ ] Clamp color shifts to visible range
- [ ] Add bloom interaction for "star wind" effect

**Task 2.6**: Documentation
- [ ] Update design doc with actual screenshots
- [ ] Note any deviations from SR formulas

---

## Success Criteria

- [ ] `go build ./...` passes
- [ ] SR effect visible at v=0.9c (tunnel vision)
- [ ] Color shift visible (blue forward, red rear)
- [ ] No visual artifacts at high gamma
- [ ] Performance < 3ms GPU time
- [ ] Screenshots generated for v=0, 0.5, 0.9, 0.99

## Files to Create

```
engine/relativity/
├── transform.go      (~150 LOC) - SR math functions
├── transform_test.go (~100 LOC) - unit tests
└── color.go          (~80 LOC)  - Doppler color shifts

engine/shader/shaders/
└── sr_warp.kage      (~100 LOC) - SR background shader
```

## Files to Modify

```
engine/shader/manager.go  - register sr_warp shader
engine/shader/effects.go  - add SR effect type
engine/screenshot/*.go    - add velocity flag
cmd/game/main.go          - add --velocity CLI flag
```

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Kage shader math limitations | Medium | High | Fall back to CPU pre-transform |
| Color shift looks wrong | Medium | Medium | Use reference images from MIT game |
| Performance issues | Low | Medium | LOD for high gamma |

## AILANG Feedback (if needed)

**Not expected for this sprint** - all engine work.

If we later integrate with AILANG:
- Will need Vec3 type or 3-float record
- May need float math functions (sqrt, etc.)

---

## Handoff to Sprint Executor

```bash
# Create progress tracking file
.claude/skills/sprint-planner/scripts/create_sprint_json.sh sr-visual-demo design_docs/planned/sprint-relativistic-effects.md
```

**Ready for execution**: This sprint is self-contained engine work with clear deliverables.
