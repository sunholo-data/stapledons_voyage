# Dome State Migration to AILANG

## Status
- Status: **Implemented** ✅
- Priority: P1 (Architecture)
- Complexity: Medium
- Part of: [view-layer-ailang-migration.md](view-layer-ailang-migration.md)
- Estimated: 1 day
- Actual: 2 days (Dec 9-10, 2025)
- Sprint: [bridge-dome-migration-sprint.md](../../../sprints/bridge-dome-migration-sprint.md)

## Problem Statement

The `DomeRenderer` in `engine/view/dome_renderer.go` owns game state that affects gameplay:

```go
// Current (WRONG) - Go owns gameplay state
type DomeRenderer struct {
    cruiseTime     float64  // Affects time dilation display
    cruiseVelocity float64  // Affects SR effects, time dilation
    cruiseDuration float64  // Journey length
    // ... rendering state (OK to stay)
}
```

**Why this is wrong:**
- `cruiseVelocity` affects **time dilation** - this is core gameplay
- Time dilation determines how much galactic time passes
- If Go owns this state, AILANG can't properly calculate consequences

**Why this isn't just "visual":**
- At v=0.9c, γ=2.29 → 1 ship year = 2.29 galactic years
- This affects civilization evolution, crew aging, mission timeline
- The velocity IS the game mechanic, not just a visual parameter

## Target Architecture

```
AILANG (sim/dome.ail)          Engine (engine/view/dome_renderer.go)
├── DomeState                  ├── Receives DomeState from AILANG
│   ├── cruise_time: float     ├── Renders planets at positions
│   ├── velocity: float        ├── Applies SR shader with velocity
│   └── journey_progress: float└── No Update(), no state ownership
├── step_dome(state, dt)
└── render_dome(state) → [DrawCmd]
```

## AILANG Implementation

### Types (sim/dome.ail)

```ailang
module sim/dome

import std/prelude
import sim/protocol (DrawCmd)

-- Dome observation state
type DomeState = {
    cruise_time: float,       -- Current time in cruise animation
    velocity: float,          -- Ship velocity (0.0 - 0.99c)
    journey_progress: float,  -- 0.0 to 1.0 through current journey
    journey_duration: float,  -- Total journey time in ship-years
    target_system: Option(SystemID)
}

-- Initialize dome for a new journey
export pure func init_dome(velocity: float, duration: float, target: Option(SystemID)) -> DomeState {
    {
        cruise_time: 0.0,
        velocity: velocity,
        journey_progress: 0.0,
        journey_duration: duration,
        target_system: target
    }
}

-- Step dome state each frame
export pure func step_dome(state: DomeState, dt: float) -> DomeState {
    let new_time = state.cruise_time + dt;
    let new_progress = min(1.0, state.journey_progress + dt / state.journey_duration);

    { state |
        cruise_time: new_time,
        journey_progress: new_progress
    }
}

-- Calculate time dilation factor (gamma)
export pure func gamma(velocity: float) -> float {
    -- γ = 1 / sqrt(1 - v²/c²)
    -- velocity is already fraction of c
    let v_squared = velocity * velocity;
    1.0 / sqrt(1.0 - v_squared)
}

-- Calculate galactic time elapsed for ship time
export pure func galactic_time(ship_time: float, velocity: float) -> float {
    ship_time * gamma(velocity)
}
```

### Rendering (sim/dome.ail)

```ailang
-- Generate DrawCmds for dome view
export pure func render_dome(state: DomeState, screen_w: int, screen_h: int) -> [DrawCmd] {
    -- Dome background (shader will apply SR effects)
    let bg_cmd = GalaxyBg(state.velocity, 0.0);  -- velocity for SR, 0 for GR

    -- Target system indicator (if approaching)
    let target_cmds = match state.target_system {
        None => [],
        Some(sys) => render_target_indicator(sys, state.journey_progress)
    };

    -- Journey progress indicator
    let progress_cmd = render_progress_bar(state.journey_progress, screen_w);

    -- Time display
    let time_cmds = render_time_display(state, screen_w, screen_h);

    concat([bg_cmd], concat(target_cmds, concat([progress_cmd], time_cmds)))
}

pure func render_time_display(state: DomeState, w: int, h: int) -> [DrawCmd] {
    let ship_years = state.cruise_time;
    let galactic_years = galactic_time(ship_years, state.velocity);
    let gamma_val = gamma(state.velocity);

    [
        -- Ship time
        Text("Ship: " ++ formatYears(ship_years), 20.0, 20.0, 8),
        -- Galactic time
        Text("Galaxy: " ++ formatYears(galactic_years), 20.0, 40.0, 8),
        -- Gamma factor
        Text("γ=" ++ formatFloat(gamma_val, 2), 20.0, 60.0, 6),
        -- Velocity
        Text("v=" ++ formatFloat(state.velocity, 2) ++ "c", 20.0, 80.0, 6)
    ]
}
```

## Engine Changes

### Before (dome_renderer.go)

```go
type DomeRenderer struct {
    cruiseTime     float64  // DELETE - move to AILANG
    cruiseVelocity float64  // DELETE - move to AILANG
    cruiseDuration float64  // DELETE - move to AILANG

    // Keep these - pure rendering
    planetRenderer *PlanetRenderer
    srShader       *shader.SRWarp
    starLayers     []*StarLayer
}

func (d *DomeRenderer) Update(dt float64) {
    // DELETE this method - AILANG does the stepping
    d.cruiseTime += dt
    // ...
}
```

### After (dome_renderer.go)

```go
type DomeRenderer struct {
    // Only rendering resources
    planetRenderer *PlanetRenderer
    srShader       *shader.SRWarp
    starLayers     []*StarLayer
}

// No Update() method - stateless

func (d *DomeRenderer) Render(screen *ebiten.Image, state sim_gen.DomeState) {
    // Apply SR shader with velocity from AILANG
    d.srShader.SetVelocity(state.Velocity)

    // Render background
    d.renderStars(screen, state.CruiseTime)

    // Render target system if approaching
    if state.TargetSystem != nil {
        d.renderTargetSystem(screen, state.TargetSystem, state.JourneyProgress)
    }
}
```

## Integration with World State

The dome state should be part of the world when in cruise mode:

```ailang
-- In sim/world.ail
type World = {
    -- ... existing fields
    mode: GameMode,
    dome: Option(DomeState)  -- Present when cruising
}

-- In sim/step.ail
pure func step(world: World, input: FrameInput) -> World {
    match world.mode {
        ModeCruise => {
            let new_dome = match world.dome {
                Some(d) => Some(step_dome(d, input.dt)),
                None => None
            };
            { world | dome: new_dome }
        },
        _ => world
    }
}
```

## Migration Steps

### Phase 1: Add AILANG Types ✅
- [x] Create `sim/bridge.ail` with DomeState type (in bridge module, not separate dome.ail)
- [ ] Add `gamma()` and `galactic_time()` functions (deferred - needs floatToStr for display)
- [x] Add `stepDome()` function
- [x] Run `make sim` to generate Go code

### Phase 2: Wire Into World ✅
- [x] Add `domeState: DomeState` to BridgeState type
- [x] Update `stepBridge()` to call `stepDome()`
- [x] Pass DomeState to `renderDome()` function

### Phase 3: Refactor Engine ✅
- [x] Disable planet rendering in `DomeRenderer` (Go planets commented out)
- [x] AILANG renders planets via CircleRGBA DrawCmd
- [x] Galaxy background via GalaxyBg DrawCmd
- [x] Go engine acts as "dumb renderer" for AILANG DrawCmds

### Phase 4: Cleanup (Partial)
- [x] Go dome_renderer no longer owns cruise animation state
- [ ] Remove remaining Go time dilation calculations (deferred)
- [x] Visual verification via screenshots at frames 60, 600, 900

## Success Criteria

- [x] DomeState defined in AILANG (`sim/bridge.ail`)
- [x] `stepDome()` called from AILANG step function
- [x] `DomeRenderer` planet rendering disabled (AILANG renders planets)
- [ ] Time dilation calculated by AILANG `gamma()` function (needs floatToStr - GitHub #29)
- [ ] Ship time and galactic time displayed correctly (needs floatToStr)
- [x] Galaxy background renders via AILANG GalaxyBg DrawCmd
- [x] Planets fly by as cruise animation progresses
- [x] 60 FPS maintained

## Testing

```bash
# Verify gamma calculations
ailang run sim/dome.ail --entry test_gamma
# Expected: gamma(0.9) ≈ 2.294, gamma(0.99) ≈ 7.089

# Visual test
make run
# Start journey, verify time displays update correctly
```

## References

- [view-layer-ailang-migration.md](view-layer-ailang-migration.md) - Parent migration doc
- [CLAUDE.md](../../../CLAUDE.md) - AILANG-first architecture

## Implementation Notes (Dec 2025)

### What Was Implemented

1. **DomeState type** in `sim/bridge.ail`:
   - `cruiseTime`, `cruiseDuration`, `cruiseVelocity`, `cameraZ`, `targetPlanet`
   - Integrated into `BridgeState` rather than separate `dome.ail` module

2. **Planet rendering** via AILANG:
   - 5 planets (Neptune → Saturn → Jupiter → Mars → Earth)
   - Perspective projection based on `cameraZ`
   - `CircleRGBA` DrawCmd for filled circles

3. **Cruise animation**:
   - 60-second loop with smooth-step easing
   - Camera moves from z=10 to z=-155
   - Planets grow as camera approaches

4. **Galaxy background**:
   - `GalaxyBg` DrawCmd renders Milky Way
   - Go engine loads `galaxy_4k.jpg` texture

5. **Strut parallax**:
   - `strutParallax()` function moves strut tops based on `cameraZ`
   - Creates depth illusion during cruise

### Deferred Items

- **HUD display**: Blocked on `floatToStr` function (GitHub issue #29)
- **gamma() function**: Implemented but can't display without floatToStr
- **Complete Go dome_renderer removal**: Kept for fallback/star layers
