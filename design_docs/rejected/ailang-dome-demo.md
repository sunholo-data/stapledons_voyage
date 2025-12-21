# AILANG Dome Demo - Player Movement with Solar System View

---

**⚠️ REJECTED: 2025-12-20**

**Reason:** 3D player movement approach abandoned in favor of scene-based navigation.

**Why rejected:**
- Interior experience doesn't need spatial exploration (player doesn't walk around)
- Scene-based selection (deck UI) better serves conversation-focused gameplay
- Dual coordinates still needed, but for scene rendering not player movement
- AILANG will generate DrawCmds for 2D deck scenes, not 3D navigation

**Replacement:** See design decision "[2025-12-20] Interior Ship Experience: Scene-Based Navigation" in [docs/vision/design-decisions.md](../../docs/vision/design-decisions.md)

---

**Status**: Rejected
**Target**: v0.5.0
**Priority**: P1 - High (demonstrates AILANG-first architecture with dual coordinates)
**Estimated**: 1 day
**Dependencies**: sim/solar_demo.ail (solar system scene), demo-engine-dome (Go reference implementation)

## Game Vision Alignment

**Score against core pillars** ([core-pillars.md](../../docs/vision/core-pillars.md)):

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Choices Are Final | N/A | 0 | Demo/infrastructure |
| The Game Doesn't Judge | N/A | 0 | Demo/infrastructure |
| Time Has Emotional Weight | Potential | +1 | Could show solar system evolving while player walks (time dilation demo) |
| The Ship Is Home | Direct | +1 | Demonstrates observation deck interior experience |
| Grounded Strangeness | Direct | +1 | Realistic solar system physics (orbital mechanics) |
| We Are Not Built For This | N/A | 0 | Demo/infrastructure |
| **Net Score** | | **+3** | **Decision: Move forward** |

**Feature type:** Demo/Infrastructure
- Validates AILANG-first architecture for dual coordinate systems
- Demonstrates pattern for observation deck gameplay
- Reuses existing solar_demo.ail scene

**Prior Decisions:** None directly relevant in design-decisions.md

## Problem Statement

Need to validate the AILANG-first architecture for dual coordinate system rendering:

**Current State:**
- [demo-engine-dome](../../cmd/demo-engine-dome/main.go) - Working Go implementation showing dual coordinates
- [sim/solar_demo.ail](../../sim/solar_demo.ail) - Working AILANG solar system with 60+ bodies

**Gap:**
- No AILANG demo showing dual coordinate system (player in meters, scene in astronomical units)
- Need to prove AILANG can handle:
  - Player movement state (ship-local meters)
  - Scene positioning (galactic light-years or AU)
  - DrawCmd generation separating these coordinate systems

**Why This Matters:**
- Validates AILANG architecture for observation deck gameplay
- Tests DrawCmd API for dual coordinate rendering
- Demonstrates pattern other features will follow

## Goals

**Primary Goal:** Create AILANG demo with player walking inside dome viewing solar system outside.

**Success Metrics:**
- Player walks around disc (WASD input, meters per second)
- Solar system visible through dome (planets orbit realistically)
- Dual coordinates work: player position ≠ solar system position
- All logic in AILANG, rendering via existing DrawCmds
- `ailang check sim/dome_demo.ail` passes
- Go host compiles and runs smoothly

## Solution Design

### Overview

Create `sim/dome_demo.ail` that:
1. Maintains player position state (ship-local, meters)
2. Reuses solar system from solar_demo.ail (galactic, AU/light-years)
3. Generates DrawCmds separating these coordinate systems
4. Updates player position based on WASD input

**Physics Basis:**
- Player movement: Normal walking physics (5 m/s)
- Solar system: Real orbital mechanics from solar_demo.ail
- Dual coordinates: Player walks in meters, solar objects positioned in AU

### Architecture

```
sim/dome_demo.ail (AILANG)           cmd/demo-ailang-dome/main.go (Go host)
├── DomeState                        ├── ebiten.Game implementation
│   ├── playerX/Y (meters)           ├── CaptureInput() → FrameInput
│   ├── playerYaw/Pitch (radians)    ├── InitDomeDemo() → DomeState
│   └── solarSystem (from solar)     ├── StepDomeDemo(state, input)
├── stepDomeDemo()                   └── Render DrawCmds from FrameOutput
│   ├── Process WASD input
│   ├── Update player position
│   └── Update solar system orbits
└── renderDomeDemo()
    ├── Generate dome interior DrawCmds
    └── Generate solar system DrawCmds
```

### AILANG Implementation

**File:** `sim/dome_demo.ail`

```ailang
module sim/dome_demo

import sim/protocol (
    FrameInput, FrameOutput, DrawCmd, Camera
)
import sim/solar_demo (
    SolarPlanet, getAllSolarSystemBodies,
    updatePlanetRotations
)

-- Player state INSIDE ship (meters, ship-local coordinates)
export type PlayerState = {
    posX: float,      -- Meters from dome center
    posY: float,      -- Fixed at floor level (1.5m)
    posZ: float,      -- Meters from dome center
    yaw: float,       -- Look direction (radians)
    pitch: float,     -- Look angle (radians)
    walkSpeed: float  -- Meters per second
}

-- Dome configuration
export type DomeConfig = {
    radius: float,         -- Dome radius in meters (50.0 = 100m diameter)
    floorY: float,         -- Floor height (-2.0m)
    showStruts: bool,      -- Render dome structure
    showFloorProps: bool   -- Render floor markers
}

-- Full demo state
export type DomeState = {
    tick: int,
    player: PlayerState,
    dome: DomeConfig,
    solarBodies: [SolarPlanet],  -- From solar_demo.ail
    timeScale: float              -- Solar system animation speed
}

-- Initialize player at dome center, looking at solar system
export pure func initPlayer() -> PlayerState =
    { posX: 0.0, posY: 1.5, posZ: 0.0,
      yaw: 3.14159, pitch: 0.0, walkSpeed: 5.0 }

export pure func initDomeConfig() -> DomeConfig =
    { radius: 50.0, floorY: -2.0,
      showStruts: true, showFloorProps: true }

export pure func initDomeDemo() -> DomeState =
    { tick: 0,
      player: initPlayer(),
      dome: initDomeConfig(),
      solarBodies: getAllSolarSystemBodies(),
      timeScale: 1.0 }

-- Process WASD input to update player position
pure func updatePlayerPosition(
    player: PlayerState,
    input: FrameInput,
    domeRadius: float
) -> PlayerState {
    -- Extract WASD from input.keys
    let moveSpeed = player.walkSpeed * 0.016;  -- Assume 60 FPS

    -- Forward/back (W/S)
    let dx1 = if input.wPressed then moveSpeed * sin(player.yaw) else 0.0;
    let dz1 = if input.wPressed then moveSpeed * cos(player.yaw) else 0.0;
    let dx2 = if input.sPressed then -moveSpeed * sin(player.yaw) else 0.0;
    let dz2 = if input.sPressed then -moveSpeed * cos(player.yaw) else 0.0;

    -- Strafe left/right (A/D)
    let dx3 = if input.aPressed then moveSpeed * cos(player.yaw) else 0.0;
    let dz3 = if input.aPressed then -moveSpeed * sin(player.yaw) else 0.0;
    let dx4 = if input.dPressed then -moveSpeed * cos(player.yaw) else 0.0;
    let dz4 = if input.dPressed then moveSpeed * sin(player.yaw) else 0.0;

    -- Sum all movements
    let newX = player.posX + dx1 + dx2 + dx3 + dx4;
    let newZ = player.posZ + dz1 + dz2 + dz3 + dz4;

    -- Clamp to dome boundary (leave 1m margin)
    let maxDist = domeRadius - 1.0;
    let dist = sqrt(newX * newX + newZ * newZ);
    let clampedX = if dist > maxDist then newX * maxDist / dist else newX;
    let clampedZ = if dist > maxDist then newZ * maxDist / dist else newZ;

    { posX: clampedX, posY: player.posY, posZ: clampedZ,
      yaw: player.yaw, pitch: player.pitch, walkSpeed: player.walkSpeed }
}

-- Generate floor disc DrawCmd (fixed in ship-local coordinates)
pure func drawFloorDisc(domeRadius: float, floorY: float) -> DrawCmd =
    DrawCmd.Circle(0.0, floorY, domeRadius * 2.0, 0x1F1F24FF, true)

-- Generate floor props (markers at various radii)
pure func drawFloorProps(domeRadius: float, floorY: float) -> [DrawCmd] =
    [ DrawCmd.Rect(0.0, floorY + 0.1, 1.0, 1.0, 0x4A5060FF),
      DrawCmd.Circle(10.0, floorY + 0.1, 1.0, 0x4A5060FF, true),
      DrawCmd.Circle(20.0, floorY + 0.1, 1.0, 0x4A5060FF, true),
      DrawCmd.Circle(35.0, floorY + 0.1, 1.0, 0x4A5060FF, true) ]

-- Map solar planet to DrawCmd (positioned in dome view)
-- Solar positions are in AU (40-600 AU), we need to map to visual distance
pure func solarPlanetToDrawCmd(p: SolarPlanet) -> DrawCmd {
    -- Map orbital radius (40-600 AU) to visual distance (80-460 meters)
    let minOrbit = 40.0;
    let maxOrbit = 600.0;
    let minRender = 80.0;
    let maxRender = 460.0;

    let orbitDist = p.orbitRadius;
    let clamped = if orbitDist < minOrbit then minOrbit
                  else if orbitDist > maxOrbit then maxOrbit
                  else orbitDist;
    let t = (clamped - minOrbit) / (maxOrbit - minOrbit);
    let renderDist = minRender + t * (maxRender - minRender);

    -- Calculate position based on orbital phase
    let x = renderDist * cos(p.orbitPhase);
    let z = renderDist * sin(p.orbitPhase);

    DrawCmd.Circle(x, 0.0, p.radius, p.colorRgba, true)
}

pure func mapSolarBodies(bodies: [SolarPlanet]) -> [DrawCmd] =
    match bodies {
        [] => [],
        p :: rest => solarPlanetToDrawCmd(p) :: mapSolarBodies(rest)
    }

-- Generate all DrawCmds for the dome view
export pure func renderDomeDemo(state: DomeState) -> [DrawCmd] {
    -- Background (space/stars)
    let bgCmd = DrawCmd.SpaceBg(0);

    -- Floor disc (fixed in ship-local coordinates)
    let floorCmd = drawFloorDisc(state.dome.radius, state.dome.floorY);

    -- Floor props (optional)
    let propCmds = if state.dome.showFloorProps
                   then drawFloorProps(state.dome.radius, state.dome.floorY)
                   else [];

    -- Solar system bodies (mapped to visual distance)
    let solarCmds = mapSolarBodies(state.solarBodies);

    -- Combine all (background first, then floor, then props, then solar)
    bgCmd :: floorCmd :: concatLists(propCmds, solarCmds)
}

-- Helper to concatenate lists
pure func concatLists(a: [DrawCmd], b: [DrawCmd]) -> [DrawCmd] =
    match a {
        [] => b,
        x :: xs => x :: concatLists(xs, b)
    }

-- Build camera from player state
pure func buildCamera(player: PlayerState) -> Camera =
    { x: player.posX, y: player.posY, zoom: 1.0 }

-- Step function: update state based on input
export pure func stepDomeDemo(
    state: DomeState,
    input: FrameInput
) -> (DomeState, FrameOutput) {
    -- Update player position based on WASD
    let newPlayer = updatePlayerPosition(
        state.player,
        input,
        state.dome.radius
    );

    -- Update solar system orbital phases
    let updatedSolar = updatePlanetRotations(state.solarBodies);

    -- Build new state
    let newState = { tick: state.tick + 1,
                     player: newPlayer,
                     dome: state.dome,
                     solarBodies: updatedSolar,
                     timeScale: state.timeScale };

    -- Generate DrawCmds
    let drawCmds = renderDomeDemo(newState);

    -- Build FrameOutput
    let output = { draw: drawCmds,
                   sounds: [],
                   debug: [],
                   camera: buildCamera(newPlayer),
                   relativity: buildDefaultRelativity(),
                   lighting: buildDefaultLighting(),
                   lod: buildDefaultLOD() };

    (newState, output)
}

-- Default contexts (minimal for dome demo)
pure func buildDefaultRelativity() -> RelativityContext =
    { sr: { enabled: false, velocity: 0.0, gamma: 1.0, viewAngle: 0.0 },
      gr: { enabled: false, centerX: 0.5, centerY: 0.5, phi: 0.0, rs: 0.0, objectType: "none" } }

pure func buildDefaultLighting() -> LightingContext =
    { enabled: true,
      ambient: { energy: 0.3, color: { r: 0.3, g: 0.3, b: 0.35 } },
      lights: [],
      lightMultiplier: 1.0 }

pure func buildDefaultLOD() -> LODConfig =
    { enabled: false, transitionTime: 0.0, hysteresis: 0.0,
      full3DPixels: 0.0, billboardPixels: 0.0, circlePixels: 0.0, pointPixels: 0.0,
      max3DObjects: 0 }
```

### Go Host Implementation

**File:** `cmd/demo-ailang-dome/main.go`

```go
package main

import (
    "github.com/hajimehoshi/ebiten/v2"
    "stapledons_voyage/engine/render"
    "stapledons_voyage/sim_gen"
)

type Game struct {
    state *sim_gen.DomeState
}

func (g *Game) Update() error {
    input := captureInput()  // Mouse, WASD keys
    newState, output := sim_gen.StepDomeDemo(g.state, input)
    g.state = newState
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    _, output := sim_gen.StepDomeDemo(g.state, emptyInput)
    render.RenderFrame(screen, output)
}

func main() {
    game := &Game{state: sim_gen.InitDomeDemo()}
    ebiten.RunGame(game)
}
```

## Implementation Plan

**Phase 1: Create AILANG Module** (~3 hours)
- [ ] Create `sim/dome_demo.ail` with types and functions
- [ ] Run `ailang check sim/dome_demo.ail`
- [ ] Fix any type errors or syntax issues

**Phase 2: Generate Go Code** (~1 hour)
- [ ] Run `make sim` to compile AILANG → Go
- [ ] Check `sim_gen/dome_demo.go` generated correctly
- [ ] Verify exported functions exist

**Phase 3: Create Go Host** (~2 hours)
- [ ] Create `cmd/demo-ailang-dome/main.go`
- [ ] Implement ebiten.Game interface
- [ ] Wire up input capture (WASD, mouse)
- [ ] Test basic rendering

**Phase 4: Test & Polish** (~1 hour)
- [ ] Test: Walk around dome - floor stays fixed
- [ ] Test: Solar system orbits correctly
- [ ] Test: Dual coordinates work (player ≠ solar positions)
- [ ] Add HUD showing player position
- [ ] Update documentation

## Files to Create/Modify

**New files:**
- `sim/dome_demo.ail` - AILANG demo module (~300 LOC)
- `cmd/demo-ailang-dome/main.go` - Go host (~150 LOC)

**Modified files:**
- `Makefile` - Add `demo-ailang-dome` target

## Success Criteria

- [ ] `ailang check sim/dome_demo.ail` passes
- [ ] `make sim` compiles without errors
- [ ] `go run ./cmd/demo-ailang-dome` runs smoothly
- [ ] Player can walk around disc with WASD
- [ ] Solar system orbits visibly through dome
- [ ] Dual coordinates demonstrated (player in meters, solar in AU)
- [ ] Frame rate > 30 FPS

## Testing Strategy

**AILANG validation:**
```bash
ailang check sim/dome_demo.ail
ailang run --entry initDomeDemo sim/dome_demo.ail
```

**Manual testing:**
```bash
make sim
go run ./cmd/demo-ailang-dome
```

- Walk with WASD - verify movement smooth
- Look around with mouse - verify view updates
- Watch solar system - verify planets orbit
- Check HUD - verify player position displayed

## Non-Goals

**Not in this demo:**
- 3D rendering (use 2D DrawCmds only)
- SR/GR effects (stationary observation)
- Multiple rooms/decks (single dome only)
- NPC crew (player only)
- Full game integration (standalone demo)

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| DrawCmd API insufficient for dual coordinates | High | Design doc validated with existing demo-engine-dome |
| AILANG list operations too slow for 60+ planets | Medium | Profile and optimize, or reduce planet count in demo |
| Input handling mismatch between AILANG and Go | Medium | Use existing FrameInput protocol from protocol.ail |

## References

- [sim/solar_demo.ail](../../sim/solar_demo.ail) - Solar system scene to reuse
- [demo-engine-dome/main.go](../../cmd/demo-engine-dome/main.go) - Go reference implementation
- [engine-capabilities.md](../reference/engine-capabilities.md) - Available DrawCmd types
- [design_docs/planned/bubble-ship-dome-system.md](./bubble-ship-dome-system.md) - Dual coordinate system design

## Future Work

- Extend to 3D rendering with Tetra3D DrawCmds
- Add SR/GR effects when ship moves through space
- Integrate with full game (observation deck feature)
- Add crew NPCs looking at solar system
- Add time dilation demo (planets age while player walks)

---

**Document created**: 2025-12-19
**Last updated**: 2025-12-19
