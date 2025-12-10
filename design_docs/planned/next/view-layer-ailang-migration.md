# View Layer AILANG Migration

## Status
- Status: Planned
- Priority: P0 (Architecture Fix)
- Complexity: High
- Supersedes: `01-view-system.md` (Go-centric approach)
- Blocks: All future game features

## Problem Statement

The `engine/view/` directory has become a parallel game engine that violates the AILANG-first architecture mandated in CLAUDE.md.

### Current State (WRONG)

```
engine/view/
├── bridge_view.go       # 433 lines - has state, Update(), Init()
├── dome_renderer.go     # 396 lines - owns cruise state, planet data
├── bubble_arc.go        # 391 lines - particle simulation
├── layer.go             # 170 lines - defines Camera, Input, Dialogue types
├── manager.go           # 211 lines - ViewManager coordinates views
├── planet_layer.go      # 163 lines - owns 3D planet state
├── space_view.go        # 175 lines - owns space background state
└── ... (3000+ lines total)
```

**What's Wrong:**
1. **Views own state** - BridgeView has `state *sim_gen.BridgeState` plus its own `frameCount`, `domeRenderer`, etc.
2. **Views have Update() methods** - Game logic running in Go, not AILANG
3. **Duplicate types** - `Camera`, `Input`, `Dialogue` duplicate AILANG types
4. **Hardcoded game data** - `dome_renderer.go` has planet configs, orbital distances
5. **Cruise state affects gameplay** - `cruiseTime`, `velocity` affect time dilation (game logic)

**What's OK (purely visual, no gameplay impact):**
- **Decorative particles** - `BubbleArc` debris is visual-only, engine can manage
- **Screen transitions** - Fade/wipe animations don't affect game state
- **Shader effects** - SR/GR visual distortion is pure rendering

### Required State (CORRECT)

```
sim/*.ail           # ALL game state, logic, visual state
sim_gen/*.go        # Generated Go code from AILANG
engine/render/      # Stateless DrawCmd renderer
```

**The engine should be "dumb":**
```go
// This is ALL the game loop should be:
for {
    input := render.CaptureInput()
    state = sim_gen.Step(state, input)
    cmds := sim_gen.Render(state)
    render.RenderFrame(screen, cmds)
}
```

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| All Pillars | N/A | Infrastructure/architecture fix |

**This is a technical debt issue, not a feature.** It blocks proper game development because:
- Game logic is split between AILANG and Go
- Features get implemented in the wrong layer
- Testing and debugging is harder
- AILANG codegen benefits are lost

## Architecture Target

### Layer Boundaries

| Layer | Location | Contains | Owns State? |
|-------|----------|----------|-------------|
| **AILANG Source** | `sim/*.ail` | All game logic, visual state | YES |
| **Generated Code** | `sim_gen/*.go` | Go bridge to AILANG | YES (generated) |
| **Engine** | `engine/render/` | DrawCmd → pixels | NO |
| **Entry** | `cmd/game/` | Wiring only | NO |

### What Moves to AILANG

| Current Go Code | Move To | AILANG Type |
|-----------------|---------|-------------|
| `DomeRenderer.cruiseTime` | `sim/dome.ail` | `DomeState { cruise_time: float, velocity: float }` |
| `DomeRenderer.planets` | `sim/dome.ail` | `[Planet]` with positions from world state |
| `BridgeView.frameCount` | Already in AILANG | Part of `BridgeState` |
| `layer.go Camera` | DELETE | Use `sim_gen.Camera` |
| `layer.go Input` | DELETE | Use `sim_gen.FrameInput` |
| `layer.go Dialogue` | `sim/dialogue.ail` | AILANG `Dialogue` type |

### What Stays in Engine (Purely Visual)

| Go Code | Why It Stays |
|---------|--------------|
| `BubbleArc.debris` | Decorative particles, no gameplay impact |
| `transition.go` | Visual fade/wipe, doesn't affect game state |
| Shader effects | Pure rendering math |

### What Stays in Go

| Go Code | Why It Stays |
|---------|--------------|
| `render/draw.go` | Renders DrawCmds to pixels - pure rendering |
| `render/input.go` | Captures OS input to FrameInput - pure IO |
| `assets/` | Loads sprites, fonts, audio - pure IO |
| `display/` | Window management - pure OS integration |
| `shader/` | GPU shader programs - pure rendering |
| `relativity/transform.go` | Math utilities for SR rendering - pure math |

### What Gets Deleted or Reduced

| File | Action | Reason |
|------|--------|--------|
| `engine/view/layer.go` | DELETE | Duplicate types (Camera, Input, Dialogue) - use sim_gen |
| `engine/view/manager.go` | REDUCE | Keep transition coordination, remove state ownership |
| `engine/view/ui_layer.go` | REDUCE | Keep layout helpers, UI state comes from AILANG |
| `engine/view/dome_renderer.go` | REDUCE | Keep rendering, move cruise/planet state to AILANG |

## AILANG Implementation

### New AILANG Modules

```ailang
-- sim/view_state.ail
module sim/view_state

import std/prelude

-- Current view type
type ViewType =
    | ViewSpace
    | ViewBridge
    | ViewGalaxyMap
    | ViewShip

-- Transition between views
type ViewTransition = {
    from: ViewType,
    to: ViewType,
    progress: float,      -- 0.0 to 1.0
    duration: float,
    effect: TransitionEffect
}

type TransitionEffect =
    | FadeToBlack
    | Crossfade
    | Wipe(Direction)

-- Master view state
type ViewState = {
    current: ViewType,
    transition: Option(ViewTransition),
    space: SpaceViewState,
    bridge: BridgeState,
    galaxy_map: GalaxyMapState
}
```

```ailang
-- sim/particles.ail
module sim/particles

import std/prelude
import std/rand

type Particle = {
    x: float, y: float,
    vx: float, vy: float,
    size: float,
    brightness: float,
    lifetime: float
}

-- Pure function: step all particles
export pure func step_particles(particles: [Particle], dt: float) -> [Particle] {
    filter_map(particles, \p. step_one(p, dt))
}

pure func step_one(p: Particle, dt: float) -> Option(Particle) {
    let new_life = p.lifetime - dt;
    match new_life < 0.0 {
        true => None,
        false => Some({
            p |
            x: p.x + p.vx * dt,
            y: p.y + p.vy * dt,
            lifetime: new_life
        })
    }
}

-- Spawn new particles (uses Rand effect)
export func spawn_debris(count: int, velocity: float) -> [Particle] ! {Rand} {
    gen_list(count, \_ . spawn_one(velocity))
}

func spawn_one(velocity: float) -> Particle ! {Rand} {
    let speed = 100.0 + velocity * 400.0;
    {
        x: rand_float(600.0, 680.0),
        y: rand_float(200.0, 280.0),
        vx: rand_float(-1.0, 1.0) * speed * 0.5,
        vy: rand_float(-1.0, 0.0) * speed,
        size: rand_float(1.5, 3.5),
        brightness: rand_float(0.5, 1.0),
        lifetime: rand_float(2.0, 5.0)
    }
}
```

```ailang
-- sim/dome.ail
module sim/dome

import std/prelude
import sim/particles (Particle, step_particles, spawn_debris)

type DomeState = {
    cruise_time: float,
    cruise_velocity: float,
    cruise_duration: float,
    debris: [Particle],
    spawn_accum: float
}

export pure func init_dome() -> DomeState {
    {
        cruise_time: 0.0,
        cruise_velocity: 0.15,
        cruise_duration: 60.0,
        debris: [],
        spawn_accum: 0.0
    }
}

export func step_dome(state: DomeState, dt: float) -> DomeState ! {Rand} {
    -- Update cruise time (loop)
    let new_time = match state.cruise_time + dt > state.cruise_duration {
        true => 0.0,
        false => state.cruise_time + dt
    };

    -- Step existing particles
    let stepped = step_particles(state.debris, dt);

    -- Spawn new particles based on accumulator
    let spawn_rate = 1.0 + state.cruise_velocity * 5.0;
    let new_accum = state.spawn_accum + spawn_rate * dt;
    let (spawned, final_accum) = spawn_while(new_accum, state.cruise_velocity);

    {
        state |
        cruise_time: new_time,
        debris: append(stepped, spawned),
        spawn_accum: final_accum
    }
}

-- Spawn particles while accumulator >= 1.0
func spawn_while(accum: float, vel: float) -> ([Particle], float) ! {Rand} {
    match accum >= 1.0 {
        true => {
            let p = spawn_debris(1, vel);
            let (more, final) = spawn_while(accum - 1.0, vel);
            (append(p, more), final)
        },
        false => ([], accum)
    }
}

-- Render dome to DrawCmds
export pure func render_dome(state: DomeState) -> [DrawCmd] {
    -- Background stars, planets would come from world state
    -- Debris particles
    let debris_cmds = map(state.debris, render_particle);

    -- Bubble arc edge (shimmer effect)
    let arc_cmds = render_bubble_arc(state.cruise_time);

    append(debris_cmds, arc_cmds)
}

pure func render_particle(p: Particle) -> DrawCmd {
    let alpha = floatToInt(p.brightness * 255.0);
    Circle(p.x, p.y, p.size, rgba(200, 200, 220, alpha))
}
```

### Engine Simplification

After migration, `engine/view/` should contain ONLY:

```go
// engine/view/helpers.go - 50 lines max
package view

// ComputePanelBounds - layout helper for UI DrawCmds
// This is pure math, no state
func ComputePanelBounds(anchor int, x, y, w, h, screenW, screenH float64) Rect {
    // ... layout calculation
}

// Easing functions - pure math
func EaseInOutCubic(t float64) float64 {
    // ... math
}
```

Everything else is deleted or moved.

## Migration Plan

### Phase 1: Add AILANG State (No Removal)

1. Create `sim/view_state.ail` with ViewType, ViewTransition
2. Create `sim/dome.ail` with DomeState, particles
3. Create `sim/particles.ail` with particle system
4. Update `sim/bridge.ail` to own dome state
5. Generate and verify compilation

### Phase 2: Wire AILANG to Engine

1. Update `cmd/game/main.go` to call AILANG step/render
2. Pass dome DrawCmds through existing renderer
3. Verify particles render correctly from AILANG
4. Verify cruise animation works from AILANG

### Phase 3: Remove Go State

1. Delete `BubbleArc.debris` - use AILANG particles
2. Delete `DomeRenderer.cruiseTime` - use AILANG state
3. Delete `DomeRenderer.planets` - planets from world state
4. Delete `layer.go` duplicate types
5. Delete `manager.go`, `transition.go`, `ui_layer.go`

### Phase 4: Cleanup

1. Remove dead code
2. Rename `engine/view/` to `engine/view_helpers/` (only math utilities remain)
3. Update imports
4. Run architecture check to verify

## Affected Files

### AILANG (Create/Modify)
- `sim/view_state.ail` - NEW
- `sim/dome.ail` - NEW
- `sim/particles.ail` - NEW
- `sim/bridge.ail` - Modify to include dome
- `sim/protocol.ail` - Add new DrawCmd variants if needed

### Go Engine (Delete)
- `engine/view/manager.go` - DELETE
- `engine/view/layer.go` - DELETE (keep ComputePanelBounds as helper)
- `engine/view/transition.go` - DELETE
- `engine/view/ui_layer.go` - DELETE
- `engine/view/bubble_arc.go` - DELETE (logic moves to AILANG)
- `engine/view/dome_renderer.go` - REDUCE to pure rendering helper
- `engine/view/bridge_view.go` - REDUCE to pure rendering

### Go Engine (Modify)
- `engine/render/draw.go` - May need new DrawCmd handlers
- `cmd/game/main.go` - Simplify to pure step/render loop

## Success Criteria

- [ ] All view state defined in `sim/*.ail`
- [ ] `engine/view/` contains < 200 lines total (helpers only)
- [ ] No Go code has Update() methods that modify game state
- [ ] Game loop is pure: input → step → render → draw
- [ ] Particle systems run in AILANG
- [ ] View transitions controlled by AILANG state
- [ ] Architecture check passes with 0 violations

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| AILANG particle performance | Slow at high counts | Limit particle count, optimize AILANG |
| Complex rendering helpers | May need Go helpers | Keep rendering math in Go, state in AILANG |
| Breaking existing demos | Demo commands fail | Update demos after migration |

## References

- [CLAUDE.md](../../../CLAUDE.md) - AILANG-first architecture mandate
- [01-view-system-DEFUNCT.md](../../archive/01-view-system-DEFUNCT.md) - Superseded by this doc
- [engine-capabilities.md](../reference/engine-capabilities.md) - What engine CAN do
