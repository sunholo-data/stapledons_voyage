# Architecture Rules Reference

Detailed rules for Stapledon's Voyage three-layer architecture.

## The Key Question: AILANG or Engine?

**Ask yourself:** Does this affect gameplay outcomes?

| If YES → AILANG | If NO → Engine OK |
|-----------------|-------------------|
| Player position, health, inventory | Particle animation (dust, sparks) |
| NPC behavior, dialogue state | Screen transition effects (fade, wipe) |
| Time dilation, galactic time | Shader visual effects |
| Planet data, civilizations | UI layout math (positioning) |
| Game mode, what's happening | Asset loading, window management |

### The Rule of Thumb

```
AILANG owns WHAT is happening (state, logic, decisions)
Engine owns HOW it looks (rendering, animation, polish)
```

## Layer Responsibilities

### 1. AILANG Source Layer (`sim/*.ail`)

**Purpose:** ALL game logic and state

**MUST contain:**
- Game state types (World, NPC, Planet, Civilization)
- Step functions that update state
- Decision-making logic (AI, game rules)
- Anything that affects gameplay outcomes
- Time dilation calculations (affects galactic time!)

**Rules:**
- No side effects (except declared effects like Rand, Debug)
- Deterministic given same inputs
- Compiled to Go via `ailang compile --emit-go`

**Examples of what belongs here:**
```ailang
-- Game state that affects outcomes
type DomeState = {
    cruise_time: float,    -- Affects time dilation!
    velocity: float,       -- v=0.9c → γ=2.29
    journey_progress: float
}

-- Game logic
pure func gamma(velocity: float) -> float {
    1.0 / sqrt(1.0 - velocity * velocity)
}

-- Decisions
pure func should_npc_flee(npc: NPC, threat: Threat) -> bool {
    npc.courage < threat.level
}
```

### 2. Generated Layer (`sim_gen/*.go`)

**Purpose:** Generated Go code from AILANG

**Rules:**
- NEVER manually edit (auto-generated)
- Contains Go equivalents of AILANG types and functions
- NO rendering imports

### 3. Engine Layer (`engine/*.go`)

**Purpose:** IO bridging and VISUAL-ONLY effects

**Contains:**
- `engine/render/` - DrawCmd → pixels
- `engine/assets/` - Load sprites, fonts, sounds
- `engine/display/` - Window management
- `engine/view/` - Visual helpers (transitions, particles)
- `engine/shader/` - GPU effects

**CAN contain (purely visual):**
```go
// OK - Decorative particles (no gameplay impact)
type ParticleEmitter struct {
    particles []Particle  // Visual-only state
}

// OK - Transition animation (doesn't affect game mode)
type TransitionManager struct {
    progress float64  // Just animation timing
    effect   TransitionEffect
}

// OK - Shader parameters
type SRWarp struct {
    velocity float64  // Passed FROM AILANG, used for rendering
}
```

**CANNOT contain (affects gameplay):**
```go
// WRONG - Game state in engine
type DomeRenderer struct {
    cruiseVelocity float64  // NO! Affects time dilation
    planets []Planet        // NO! Game entities
}

// WRONG - Game logic in engine
func (d *DomeRenderer) Update() {
    d.cruiseTime += dt  // NO! This is game state
}

// WRONG - Decisions in engine
if player.Health < 10 { ... }  // NO! Game logic
```

### 4. Entry Layer (`cmd/*.go`)

**Purpose:** Wire everything together (minimal code)

**The ideal game loop:**
```go
func (g *Game) Update() error {
    input := render.CaptureInput()
    g.world, g.output, _ = sim_gen.Step(g.world, input)
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    render.RenderFrame(screen, g.output)
}
```

## Detailed Boundary Guide

### What AILANG Owns (Game State)

| Category | Examples | Why AILANG? |
|----------|----------|-------------|
| **Movement** | Player position, NPC positions | Affects collision, interactions |
| **Time** | Ship time, galactic time, velocity | Core gameplay mechanic |
| **Entities** | Planets, civilizations, crew | Game data with state |
| **Decisions** | AI behavior, dialogue choices | Affects outcomes |
| **Inventory** | Items, resources, cargo | Gameplay resources |
| **Progression** | Quests, flags, unlocks | Game state |
| **Mode** | What screen/view we're in | Game state machine |

### What Engine Can Own (Visual Polish)

| Category | Examples | Why Engine OK? |
|----------|----------|----------------|
| **Particles** | Dust, sparks, debris | Decorative, no gameplay |
| **Transitions** | Fade, wipe, dissolve | Visual polish |
| **Shaders** | SR warp, bloom, vignette | Pure rendering |
| **Animation** | Sprite frame timing | Visual timing |
| **Layout** | UI positioning math | Where things draw |
| **Audio** | Sound playback | Output only |

### Edge Cases

| Situation | AILANG or Engine? | Reason |
|-----------|-------------------|--------|
| Particle hits player | **AILANG** | Affects gameplay (damage) |
| Decorative dust | **Engine** | No gameplay impact |
| Transition progress | **Engine** | Just visual timing |
| Mode change | **AILANG** | Game state |
| Camera shake | **Engine** | Visual effect |
| Camera position | **AILANG** | Affects what player sees (gameplay) |
| Font rendering | **Engine** | Visual output |
| Dialogue text | **AILANG** | Game content |

## Common Violations

### 1. Game State in Engine (WRONG)

```go
// engine/view/dome_renderer.go
type DomeRenderer struct {
    cruiseVelocity float64  // WRONG! Affects time dilation
    planets []Planet        // WRONG! Game entities
}

func (d *DomeRenderer) Update(dt float64) {
    d.cruiseTime += dt  // WRONG! Game state update
}
```

**Fix:** Move to AILANG:
```ailang
-- sim/dome.ail
type DomeState = {
    cruise_time: float,
    velocity: float
}

pure func step_dome(state: DomeState, dt: float) -> DomeState {
    { state | cruise_time: state.cruise_time + dt }
}
```

### 2. Duplicate Types (WRONG)

```go
// engine/view/layer.go
type Camera struct { ... }  // WRONG! Duplicates sim_gen.Camera
type Input struct { ... }   // WRONG! Duplicates sim_gen.FrameInput
```

**Fix:** Use sim_gen types:
```go
import "stapledons_voyage/sim_gen"

func Render(screen *ebiten.Image, cam sim_gen.Camera) { ... }
```

### 3. Hardcoded Game Data (WRONG)

```go
// engine/view/dome_renderer.go
planetConfigs := []struct{
    name string
    radius float64
}{
    {"Earth", 1.0},   // WRONG! Game data
    {"Jupiter", 11.2},
}
```

**Fix:** Data comes from AILANG:
```ailang
-- sim/celestial.ail
export pure func init_sol_system() -> StarSystem {
    { planets: [...] }
}
```

### 4. Visual Effects Are OK

```go
// engine/particle/emitter.go
type Emitter struct {
    particles []Particle  // OK - decorative only
}

func (e *Emitter) Update(dt float64) {
    // OK - this is visual animation, not game logic
    for i := range e.particles {
        e.particles[i].life -= dt
    }
}
```

## Quick Reference Table

| Code Type | Location | Allowed? |
|-----------|----------|----------|
| NPC behavior | `sim/*.ail` | YES |
| NPC behavior | `engine/` | NO |
| Particle animation | `engine/particle/` | YES |
| Particle collision damage | `sim/*.ail` | YES |
| Screen fade effect | `engine/transition/` | YES |
| Game mode change | `sim/*.ail` | YES |
| Planet data | `sim/*.ail` | YES |
| Planet rendering | `engine/render/` | YES |
| Time dilation calc | `sim/*.ail` | YES |
| SR shader effect | `engine/shader/` | YES |
| Camera shake | `engine/` | YES |
| Camera target | `sim/*.ail` | YES |

## Checking Boundaries

Run the boundary check script:
```bash
.claude/skills/game-architect/scripts/check_layer_boundaries.sh
```

This checks for:
- Rendering imports in sim_gen/
- Game logic patterns in engine/
- Duplicate type definitions

## Migration Reference

See these design docs for ongoing migrations:
- [view-layer-ailang-migration.md](../../../design_docs/planned/next/view-layer-ailang-migration.md)
- [dome-state-migration.md](../../../design_docs/planned/next/dome-state-migration.md)
- [planet-data-migration.md](../../../design_docs/planned/next/planet-data-migration.md)
- [view-types-cleanup.md](../../../design_docs/planned/next/view-types-cleanup.md)
