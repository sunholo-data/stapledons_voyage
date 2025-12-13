# Sprint: AILANG Solar System Demo

**Design Doc:** [design_docs/planned/ailang-solar-system-demo.md](../design_docs/planned/ailang-solar-system-demo.md)
**Priority:** P1 (AILANG integration validation)
**Duration:** 3 days
**Start Date:** TBD

## Goal

Create a unified solar system demo where AILANG controls all celestial data (planets, moons, rings, lighting, relativity) while the Go engine only handles rendering. This validates the AILANG-first architecture and proves protocol types work end-to-end.

## Prerequisites

- [x] LightingContext type in protocol.ail
- [x] RelativityContext type in protocol.ail
- [x] TexturedPlanet DrawCmd working
- [x] LOD system working (demo-lod)
- [x] Ring rendering working (demo-game-saturn)
- [x] GR Faint level added for subtle testing

## Day 1: AILANG Types & Basic State

### Morning: Core Types
- [ ] Create `sim/solar_demo.ail` module
- [ ] Define Vector3 type for 3D positions
- [ ] Define Moon type (name, radius, color, orbit params)
- [ ] Define RingBand type (inner/outer radius, color, opacity)
- [ ] Define Planet type (position, radius, texture, moons, rings)
- [ ] Run `ailang check sim/solar_demo.ail`

### Afternoon: State & Init
- [ ] Define SolarDemoState type (tick, camera, velocity, planets, lighting params)
- [ ] Implement `init_solar_demo()` function
- [ ] Create helper functions for each planet:
  - [ ] `create_mercury()` - small gray rock
  - [ ] `create_venus()` - yellow-white
  - [ ] `create_earth()` - blue with 1 moon
  - [ ] `create_mars()` - red with 2 moons
  - [ ] `create_jupiter()` - banded with 4 moons
  - [ ] `create_saturn()` - ringed with 6 moons
  - [ ] `create_uranus()` - tilted with thin rings
  - [ ] `create_neptune()` - blue with 1 moon
- [ ] Run `make sim` to generate Go code
- [ ] Fix any compilation errors

### Day 1 Checkpoint
```bash
ailang check sim/solar_demo.ail
make sim && go build ./...
```

**Files created:**
- `sim/solar_demo.ail`

## Day 2: Step Function & DrawCmd Generation

### Morning: Input Handling
- [ ] Import FrameInput, FrameOutput, DrawCmd from protocol
- [ ] Implement camera movement functions:
  - [ ] `handle_camera_input(state, input) -> (x, y, z)`
  - [ ] `handle_velocity_input(velocity, input) -> velocity`
- [ ] Implement `update_planet_rotations(planets) -> planets`

### Afternoon: Draw Commands
- [ ] Implement `generate_draw_commands(planets) -> [DrawCmd]`:
  - [ ] SpaceBg for background
  - [ ] Star for sun sprite
  - [ ] TexturedPlanet for each planet
- [ ] Implement `build_lighting_context(sunEnergy, ambientLevel) -> LightingContext`
- [ ] Implement `build_relativity_context(velocity, grEnabled, ...) -> RelativityContext`
- [ ] Implement main `step_solar_demo(state, input) -> (state, FrameOutput)`
- [ ] Run `make sim` and verify compilation

### Day 2 Checkpoint
```bash
ailang check sim/solar_demo.ail
make sim && go build ./...
```

**Files modified:**
- `sim/solar_demo.ail` (step function)
- `sim_gen/*.go` (regenerated)

## Day 3: Go Demo Entry Point & Integration

### Morning: Basic Demo
- [ ] Create `cmd/demo-ailang-solar/main.go`
- [ ] Implement Ebiten Game struct with:
  - [ ] sim_gen state field
  - [ ] Tetra3D scene
  - [ ] LOD manager
  - [ ] Shader manager (SR/GR)
- [ ] Implement `Update()`:
  - [ ] Capture input → FrameInput
  - [ ] Call `sim_gen.StepSolarDemo()`
  - [ ] Apply lighting context to Tetra3D
  - [ ] Apply relativity context to shaders
- [ ] Implement `Draw()`:
  - [ ] Render DrawCmd list
  - [ ] Handle TexturedPlanet with LOD
- [ ] Build and test: `go build -o bin/demo-ailang-solar ./cmd/demo-ailang-solar`

### Afternoon: Polish & Testing
- [ ] Add overlay UI showing:
  - [ ] Current planet count
  - [ ] LOD tier counts
  - [ ] Lighting status
  - [ ] SR/GR effect status
- [ ] Test each system:
  - [ ] LOD: Move camera to verify tier transitions
  - [ ] Lighting: Verify sun illuminates planets
  - [ ] SR: Increase velocity, verify Doppler shift
  - [ ] GR: Enable GR near object, verify lensing
- [ ] Add screenshot mode support
- [ ] Verify `--screenshot 60` works

### Day 3 Checkpoint
```bash
bin/demo-ailang-solar --screenshot 60 --output out/screenshots/ailang-solar.png
```

**Files created:**
- `cmd/demo-ailang-solar/main.go`

## Controls Reference

| Key | Action | State Field |
|-----|--------|-------------|
| WASD | Move camera XZ | cameraX, cameraZ |
| Q/E | Camera up/down | cameraY |
| [ ] | Ship velocity | shipVelocity |
| 1 | Toggle SR | (engine) |
| 2 | Toggle GR | grEnabled |
| 3 | Cycle GR intensity | grPhi |
| ; ' | Ambient light | ambientLevel |
| , . | Sun light | sunEnergy |
| Tab | Toggle overlay | (engine) |
| R | Reset camera | cameraX/Y/Z |

## Success Criteria

- [ ] `ailang check sim/solar_demo.ail` passes
- [ ] `make sim && go build ./...` succeeds
- [ ] Demo launches: `bin/demo-ailang-solar`
- [ ] All 8 planets render with TexturedPlanet
- [ ] Saturn rings visible when close
- [ ] Moons orbit parent planets
- [ ] LightingContext controls Tetra3D lights
- [ ] RelativityContext enables SR/GR shaders
- [ ] LOD transitions work smoothly
- [ ] No hardcoded planet data in Go
- [ ] Screenshot mode works

## AILANG Features Used

| Feature | Usage | Risk |
|---------|-------|------|
| Module imports | Import protocol types | Low (working) |
| Record types | Planet, Moon, State | Low |
| Tagged unions | DrawCmd generation | Low |
| List operations | Planet/moon iteration | Low |
| `std/math` | sqrt for gamma | Low (working) |
| Float comparison | Velocity thresholds | Low |

## Potential Blockers

| Risk | Mitigation | Fallback |
|------|------------|----------|
| List iteration limits | Keep planet count to 8 | Use smaller system |
| Math precision | Use pre-computed gamma table | Hardcode common values |
| Codegen issues | Report via ailang-feedback | Use mock sim_gen |

## Post-Sprint

After completion:
1. Move design doc to `design_docs/implemented/vX_Y_Z/`
2. Report any AILANG issues via `ailang-feedback`
3. Document patterns learned for future demos
4. Consider extracting planet data to shared `sim/celestial.ail`

## Notes

This sprint validates that AILANG can drive a complete rendering pipeline. The patterns established here will be reused for:
- Main game's arrival sequences
- Galaxy map system previews
- Any future celestial rendering
