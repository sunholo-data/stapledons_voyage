# AILANG Solar System Demo

## Status
- **Status**: Planned
- **Priority**: P1 (AILANG integration validation)
- **Estimated**: 2-3 days
- **Location**: `cmd/demo-ailang-solar/`, `sim/solar_demo.ail`

## Game Vision Alignment

| Pillar | Alignment | Notes |
|--------|-----------|-------|
| Choices Are Final | N/A | Tech demo, not gameplay |
| The Game Doesn't Judge | N/A | Tech demo |
| Time Has Emotional Weight | **Supports** | SR/GR effects visualize time dilation |
| The Ship Is Home | N/A | Tech demo |
| Grounded Strangeness | **Supports** | Accurate celestial rendering (rings, moons, lighting) |
| We Are Not Built For This | **Supports** | Overwhelming scale of solar system visible |
| **Overall** | **Validation Demo** | Proves AILANG can control full rendering pipeline |

**This demo validates the AILANG-first architecture** by having AILANG control all celestial data (planets, moons, rings, lighting) while the Go engine only handles rendering.

## Problem Statement

We have two separate tech demos:
- **demo-lod**: Tests LOD system, lighting, SR/GR effects - but uses hardcoded Go data
- **demo-game-saturn**: Tests ring rendering, moons - AILANG controlled but isolated

**Need**: A unified demo that:
1. Proves AILANG can control a full solar system
2. Tests LOD, lighting, and relativity effects together
3. Validates the protocol types (LightingContext, RelativityContext, TexturedPlanet)
4. Serves as integration test for the engine capabilities documented in protocol.ail

## Proposed Solution

### New Demo: `demo-ailang-solar`

A solar system demo where:
- **AILANG owns ALL data**: Planet positions, moons, rings, lighting, SR/GR context
- **Go engine is purely rendering**: Reads AILANG output, draws to screen
- **Tests every major system**: LOD, lighting, relativity, TexturedPlanet DrawCmd

### Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│ AILANG (sim/solar_demo.ail)                                          │
├──────────────────────────────────────────────────────────────────────┤
│ SolarDemoState                                                        │
│   - tick: int                                                         │
│   - cameraPos: Vector3                                                │
│   - shipVelocity: float (0.0-0.99c for SR)                           │
│   - nearGRObject: Option<GRObject>                                    │
│   - planets: [Planet] (with moons, rings)                            │
│   - starLight: LightSource                                            │
│                                                                       │
│ init_solar_demo() -> SolarDemoState                                   │
│ step_solar_demo(state, input) -> (SolarDemoState, FrameOutput)       │
│                                                                       │
│ FrameOutput includes:                                                 │
│   - draw: [DrawCmd] with TexturedPlanet, Star, GalaxyBg              │
│   - lighting: LightingContext (star position, color, energy)         │
│   - relativity: RelativityContext (SR from velocity, GR from objects)│
└──────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────────┐
│ Go Engine (cmd/demo-ailang-solar/main.go)                             │
├──────────────────────────────────────────────────────────────────────┤
│ 1. Call sim_gen.InitSolarDemo() → state                              │
│ 2. Each frame:                                                        │
│    - Capture input → FrameInput                                       │
│    - Call sim_gen.StepSolarDemo(state, input) → state, output        │
│    - Apply output.lighting to Tetra3D lights                          │
│    - Apply output.relativity to shader system                         │
│    - Render output.draw commands                                      │
│ 3. Handle TexturedPlanet DrawCmd with LOD system:                     │
│    - Full3D: Tetra3D sphere with texture                             │
│    - Billboard/Circle/Point: 2D fallback based on distance           │
└──────────────────────────────────────────────────────────────────────┘
```

### Solar System Layout

The demo solar system (not our real solar system, but representative):

| Object | Position (relative) | Radius | Features | LOD Test |
|--------|---------------------|--------|----------|----------|
| **Sun** | Origin (0,0,0) | 30 | Light source, yellow-white | Light emitter |
| **Mercury** | 40 units | 2 | Gray, no moons | Small/distant |
| **Venus** | 60 units | 4 | Yellow-white | Medium |
| **Earth** | 80 units | 4.5 | Blue-white, 1 moon | Moon orbit |
| **Mars** | 100 units | 3 | Red, 2 small moons | Multiple moons |
| **Jupiter** | 150 units | 15 | Banded, 4 major moons | Large + moons |
| **Saturn** | 200 units | 12 | Rings, 6 moons | Ring rendering |
| **Uranus** | 280 units | 8 | Tilted, thin rings | Dark rings |
| **Neptune** | 350 units | 7 | Blue, 1 moon | Distant |

### AILANG Types Required

```ailang
-- sim/solar_demo.ail

import sim/protocol (
    FrameInput, FrameOutput, DrawCmd, Camera,
    LightingContext, LightSource, AmbientSettings, RGBColor,
    RelativityContext, SRContext, GRContext
)

-- Vector for 3D positions
export type Vector3 = { x: float, y: float, z: float }

-- Moon definition
export type Moon = {
    name: string,
    radius: float,
    color: RGBColor,
    orbitRadius: float,
    orbitSpeed: float,
    orbitPhase: float
}

-- Ring band for ringed planets
export type RingBand = {
    innerRadius: float,
    outerRadius: float,
    color: int,        -- 0xRRGGBBAA
    opacity: float
}

-- Planet with all visual properties
export type Planet = {
    name: string,
    position: Vector3,
    radius: float,
    color: int,        -- For Circle/Point LOD tiers
    textureName: string,  -- For TexturedPlanet DrawCmd
    rotation: float,   -- Current rotation angle
    rotationSpeed: float,
    moons: [Moon],
    rings: [RingBand],
    hasRings: bool,
    ringColor: int     -- For TexturedPlanet ringRgba
}

-- Demo state
export type SolarDemoState = {
    tick: int,
    cameraX: float,
    cameraY: float,
    cameraZ: float,
    shipVelocity: float,    -- 0.0-0.99 for SR effects
    grEnabled: bool,
    grCenterX: float,       -- Screen-space for GR effect
    grCenterY: float,
    grPhi: float,           -- Gravitational potential
    planets: [Planet],
    sunEnergy: float,
    ambientLevel: float
}
```

### Key AILANG Functions

```ailang
-- Initialize the solar system
export pure func init_solar_demo() -> SolarDemoState {
    {
        tick: 0,
        cameraX: 300.0,
        cameraY: 100.0,
        cameraZ: 200.0,
        shipVelocity: 0.0,
        grEnabled: false,
        grCenterX: 0.5,
        grCenterY: 0.5,
        grPhi: 0.001,
        planets: create_planets(),
        sunEnergy: 8000.0,
        ambientLevel: 0.2
    }
}

-- Create all planets with their moons and rings
pure func create_planets() -> [Planet] {
    [
        create_mercury(),
        create_venus(),
        create_earth(),
        create_mars(),
        create_jupiter(),
        create_saturn(),
        create_uranus(),
        create_neptune()
    ]
}

-- Saturn with rings and moons
pure func create_saturn() -> Planet {
    {
        name: "Saturn",
        position: { x: 200.0, y: 0.0, z: 0.0 },
        radius: 12.0,
        color: 0xDCBE8CFF,  -- Tan
        textureName: "saturn",
        rotation: 0.0,
        rotationSpeed: 0.01,
        moons: saturn_moons(),
        rings: saturn_rings(),
        hasRings: true,
        ringColor: 0xD4C8A880  -- Semi-transparent tan
    }
}

pure func saturn_rings() -> [RingBand] {
    [
        { innerRadius: 1.24, outerRadius: 1.53, color: 0xB4A08240, opacity: 0.3 },
        { innerRadius: 1.53, outerRadius: 1.95, color: 0xDCCDAA80, opacity: 0.7 },
        { innerRadius: 2.03, outerRadius: 2.27, color: 0xD2BE9660, opacity: 0.5 }
    ]
}

pure func saturn_moons() -> [Moon] {
    [
        { name: "Titan", radius: 0.4, color: { r: 0.82, g: 0.63, b: 0.39 },
          orbitRadius: 4.0, orbitSpeed: 0.15, orbitPhase: 0.0 },
        { name: "Enceladus", radius: 0.15, color: { r: 0.94, g: 0.96, b: 1.0 },
          orbitRadius: 2.8, orbitSpeed: 0.4, orbitPhase: 1.57 }
    ]
}

-- Step function: update state and generate draw commands
export func step_solar_demo(state: SolarDemoState, input: FrameInput)
    -> (SolarDemoState, FrameOutput) ! {Rand} {

    -- Update planet rotations
    let newPlanets = update_planet_rotations(state.planets)

    -- Handle camera movement (from input)
    let (newCamX, newCamY, newCamZ) = handle_camera_input(
        state.cameraX, state.cameraY, state.cameraZ, input)

    -- Handle velocity changes (for SR effects)
    let newVelocity = handle_velocity_input(state.shipVelocity, input)

    -- Generate draw commands
    let drawCmds = generate_draw_commands(newPlanets, state.sunEnergy)

    -- Build lighting context
    let lighting = build_lighting_context(state.sunEnergy, state.ambientLevel)

    -- Build relativity context
    let relativity = build_relativity_context(
        newVelocity, state.grEnabled, state.grCenterX, state.grCenterY, state.grPhi)

    let newState = { state |
        tick: state.tick + 1,
        cameraX: newCamX,
        cameraY: newCamY,
        cameraZ: newCamZ,
        shipVelocity: newVelocity,
        planets: newPlanets
    }

    let output = {
        draw: drawCmds,
        sounds: [],
        debug: [],
        camera: { x: newCamX, y: newCamY, zoom: 1.0 },
        relativity: relativity,
        lighting: lighting
    }

    (newState, output)
}

-- Generate TexturedPlanet commands
pure func generate_draw_commands(planets: [Planet], sunEnergy: float) -> [DrawCmd] {
    -- Sun as a star sprite (not TexturedPlanet - it's a light source)
    let sunCmd = Star(640.0, 360.0, 0, 1.5, 1.0, 0)

    -- Each planet as TexturedPlanet
    let planetCmds = map_planets_to_draw_cmds(planets)

    -- Space background
    let bgCmd = SpaceBg(0)

    [bgCmd] ++ [sunCmd] ++ planetCmds
}

-- Build LightingContext for engine
pure func build_lighting_context(sunEnergy: float, ambientLevel: float) -> LightingContext {
    {
        enabled: true,
        ambient: {
            energy: ambientLevel,
            color: { r: 0.08, g: 0.08, b: 0.1 }  -- Deep space blue-black
        },
        lights: [
            {
                id: "sun",
                x: 0.0, y: 0.0, z: 0.0,
                energy: sunEnergy,
                color: { r: 1.0, g: 0.95, b: 0.85 },  -- G-type star
                range: 0.0  -- Infinite
            }
        ],
        lightMultiplier: 1.0
    }
}

-- Build RelativityContext for shader effects
pure func build_relativity_context(
    velocity: float,
    grEnabled: bool,
    grCenterX: float,
    grCenterY: float,
    grPhi: float
) -> RelativityContext {
    let gamma = if velocity > 0.01 then
        1.0 / sqrt(1.0 - velocity * velocity)
    else 1.0

    {
        sr: {
            enabled: velocity > 0.01,
            velocity: velocity,
            gamma: gamma,
            viewAngle: 0.0  -- Looking forward
        },
        gr: {
            enabled: grEnabled,
            centerX: grCenterX,
            centerY: grCenterY,
            phi: grPhi,
            rs: 0.05,
            objectType: "bh"
        }
    }
}
```

## Testing Each System

### 1. LOD System Test
- Camera starts far from Saturn → Point tier
- Move closer → Circle → Billboard → Full3D
- Verify tier transitions are smooth (no popping)
- Verify moons transition independently

### 2. Lighting System Test
- Sun light illuminates planets correctly
- Day/night terminator visible on planets
- Ambient level adjustable (AILANG controls ambientLevel)
- Star color tints planet illumination

### 3. SR Effects Test
- Increase shipVelocity → enable SR shader
- Verify Doppler shift (blue forward, red backward)
- Verify aberration (stars compress toward forward direction)
- Test at 0.5c, 0.9c, 0.99c

### 4. GR Effects Test
- Enable grEnabled near dense object
- Verify gravitational lensing around center point
- Test Faint, Subtle, Strong, Extreme intensity levels
- Verify works at different grCenterX/Y positions

### 5. Ring Rendering Test
- Saturn rings visible when close
- Rings occlude correctly (front/back)
- Ring transparency works
- Uranus dark rings render differently

### 6. Moon Orbit Test
- Moons orbit their parent planet
- Moon orbits controlled by AILANG (orbitSpeed, orbitPhase)
- Moons have independent LOD

## Controls

| Key | Action | AILANG State Change |
|-----|--------|---------------------|
| WASD | Move camera | cameraX/Y/Z |
| Q/E | Camera up/down | cameraZ |
| 1-4 | Warp to planet group | cameraX/Y/Z preset |
| [ ] | Adjust ship velocity | shipVelocity +/- 0.1c |
| G | Toggle GR effect | grEnabled |
| I/K | Adjust GR intensity | grPhi |
| ; ' | Adjust ambient light | ambientLevel |
| , . | Adjust sun light | sunEnergy |
| Tab | Toggle overlay | (engine only) |

## Files to Create

| File | Purpose |
|------|---------|
| `sim/solar_demo.ail` | AILANG solar system state and step function |
| `cmd/demo-ailang-solar/main.go` | Go entry point, render loop |
| `sim_gen/solar_demo.go` | Generated code (via `make sim`) |

## Success Criteria

- [ ] `ailang check sim/solar_demo.ail` passes
- [ ] `make sim && go build ./...` succeeds
- [ ] Demo launches: `bin/demo-ailang-solar`
- [ ] All 8 planets render with TexturedPlanet DrawCmd
- [ ] Saturn rings render correctly
- [ ] Moons orbit their parent planets
- [ ] Lighting controlled by AILANG (LightingContext)
- [ ] SR effects activate when velocity > 0.01c
- [ ] GR effects activate when grEnabled is true
- [ ] LOD transitions work (verified by moving camera)
- [ ] No hardcoded planet/moon/ring data in Go
- [ ] Screenshot mode works: `--screenshot 60`

## Dependencies

### On Existing Features
- [ ] TexturedPlanet DrawCmd rendering (engine/render/draw.go)
- [ ] LightingContext handling (engine integration)
- [ ] RelativityContext → shader pipeline (engine/shader/)
- [ ] LOD system (engine/lod/)
- [ ] Tetra3D RingSystem (engine/tetra/ring.go)

### On AILANG
- [ ] `std/math` for sqrt, trig functions
- [ ] List operations for planet/moon iteration
- [ ] Record update syntax
- [ ] Float comparisons

## Builds On

| Demo | What We Take |
|------|--------------|
| `demo-lod` | LOD system integration, lighting setup, SR/GR demo mode |
| `demo-game-saturn` | Ring rendering, moon orbit math, TexturedPlanet usage |
| `demo-orbital` | Camera movement patterns |

## Migration Path

After this demo validates the architecture:
1. Move planet data to `sim/celestial.ail` (shared with main game)
2. Main game uses same patterns for arrival sequences
3. Galaxy map can call same functions for system preview

## References

- [celestial-lod-system.md](celestial-lod-system.md) - LOD tier design
- [light-lod-system.md](light-lod-system.md) - Light LOD (deferred)
- [ailang-planet-ring-moon-data.md](ailang-planet-ring-moon-data.md) - AILANG data types
- [protocol.ail](../../sim/protocol.ail) - DrawCmd, LightingContext, RelativityContext
- [engine-capabilities.md](reference/engine-capabilities.md) - Engine features
