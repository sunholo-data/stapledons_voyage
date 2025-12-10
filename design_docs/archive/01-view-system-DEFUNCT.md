# View System Architecture

> **DEFUNCT - ARCHIVED 2025-12-10**
>
> This document describes a Go-centric view system that violates the AILANG-first architecture.
> It has been superseded by [view-layer-ailang-migration.md](../planned/next/view-layer-ailang-migration.md).
>
> **Why this was wrong:**
> - Defines Go interfaces (View, BackgroundLayer, ContentLayer, UILayer) instead of AILANG types
> - Views own state and have Update() methods - game logic should be in AILANG
> - Duplicate types (Camera, Input, Dialogue) that should come from sim_gen
> - Engine should be stateless - just render DrawCmds from AILANG

## Status
- Status: ~~Planned~~ **DEFUNCT**
- Priority: P0 (Foundation)
- Complexity: Medium
- Estimated: 3-4 days
- Blocks: All game views, arrival sequence

## Problem Statement

The game needs multiple view types that share common elements:
- **Space exterior** - Starfield background with 3D planets
- **Bridge interior** - Isometric view with crew stations
- **Galaxy map** - 2D/3D star navigation
- **Ship interior** - Isometric exploration

Currently these are ad-hoc. We need a composable view system.

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Hard Sci-Fi Authenticity | +++ | Views show realistic space, proper physics |
| Ship Is Home | ++ | Bridge and ship views establish home |
| Time Has Emotional Weight | + | Views support time displays |

## Architecture Overview

### Layer Model

Every game view composes three layers:

```
┌─────────────────────────────────────────────────┐
│                    UI LAYER                      │  Z: 100-199
│         (HUD, panels, dialogue, menus)           │
├─────────────────────────────────────────────────┤
│                 CONTENT LAYER                    │  Z: 10-99
│    (3D planets, isometric tiles, entities)       │
├─────────────────────────────────────────────────┤
│               BACKGROUND LAYER                   │  Z: 0-9
│         (starfield, nebulae, gradients)          │
└─────────────────────────────────────────────────┘
```

### View Types

```go
type ViewType int
const (
    ViewSpace       ViewType = iota  // Exterior space with planets
    ViewBridge                        // Bridge interior (isometric)
    ViewShip                          // Ship exploration (isometric)
    ViewGalaxyMap                     // Star navigation
    ViewPlanetSurface                 // Ground exploration (isometric)
)
```

### View Interface

```go
type View interface {
    // Lifecycle
    Init() error
    Enter(from ViewType)
    Exit(to ViewType)

    // Update/Draw
    Update(dt float64, input *Input) ViewTransition
    Draw(screen *ebiten.Image)

    // Layer management
    Background() BackgroundLayer
    Content() ContentLayer
    UI() UILayer
}

type ViewTransition struct {
    To       ViewType
    Duration float64           // Transition time in seconds
    Effect   TransitionEffect  // Fade, wipe, etc.
}
```

## Layer Specifications

### Background Layer

Renders behind everything. Usually space/starfield.

```go
type BackgroundLayer interface {
    SetParallax(depth float64)    // 0=static, 1=full camera motion
    SetVelocity(v float64)        // For SR aberration effects
    Draw(screen *ebiten.Image, camera *Camera)
}

// Implementation
type SpaceBackground struct {
    starLayers   []*StarLayer      // Multiple parallax depths
    nebulae      []*NebulaSprite   // Optional nebula overlays
    srWarp       *shader.SRWarp    // SR effects applied
    grWarp       *shader.GRWarp    // GR effects applied
}
```

**Star layers** (physics-based parallax):
| Layer | Stars | Parallax | Purpose |
|-------|-------|----------|---------|
| Far   | 500   | 0.0      | Fixed distant stars |
| Mid   | 300   | 0.3      | Slight motion |
| Near  | 100   | 0.7      | Foreground stars |

### Visual Physics Design Decisions

> **Note**: These are initial design choices balancing realism vs gameplay feel. Subject to modification after playtesting.

#### The Reality Problem

Real stellar parallax is nearly imperceptible at human timescales:
- Earth's entire orbit (2 AU) produces only ~0.77 arcseconds parallax for the nearest star
- At 0.9c, you'd need hours of travel to see noticeable shift for nearby stars
- SR effects (aberration, Doppler) dominate the visual experience long before parallax becomes visible

#### Speed Thresholds

| Speed | Primary Visual Effect | Parallax Visibility |
|-------|----------------------|---------------------|
| < 0.1c | Subtle star motion | None (use dust/particles for motion cue) |
| 0.1c - 0.3c | Minor aberration starting | Foreground layers only |
| **0.3c - 0.5c** | Noticeable aberration | **Nearby stars begin shifting** |
| 0.5c - 0.9c | Strong aberration (60°→26° cone) | Visible for stars <20 ly |
| > 0.9c | Extreme "starbow" effect | Rapid parallax, but aberration dominates |

#### Dual View Mode (Recommended Approach)

1. **"Raw" SR View** - What eyes would actually see:
   - Aberration: Stars bunch toward direction of travel
   - Doppler: Blue-shift ahead, red-shift behind
   - Parallax only visible on long journeys
   - Authentic but potentially disorienting

2. **"Navigation" View** - Computer-enhanced display:
   - Compensates for aberration (shows "true" star positions)
   - Exaggerated parallax for nearby stars (<20 ly)
   - Artificial depth layers (nebula wisps, dust) for motion feedback
   - More intuitive for gameplay

#### Parallax Layer Design

```
Speed < 0.3c:
┌────────────────────────────────────┐
│  ✦    ✦   ✦     ✦   ✦    ✦   ✦    │  Background stars (static)
│    ✦      ✦   ✦      ✦      ✦     │
│  ░░░░░  ░░░░░    ░░░░░░░░  ░░░    │  Dust/particle layer (fast parallax)
└────────────────────────────────────┘

Speed 0.3c - 0.5c:
┌────────────────────────────────────┐
│  ✦    ✦   ✦     ✦   ✦    ✦   ✦    │  Distant stars (static)
│    ★      ★   ★      ★      ★     │  Nearby stars (subtle parallax)
│  ░░░░░  ░░░░░    ░░░░░░░░  ░░░    │  Dust layer (fast parallax)
└────────────────────────────────────┘

Speed > 0.5c (Raw view):
┌────────────────────────────────────┐
│           ★✦✦★✦★✦★                 │  Aberration cone (stars bunch ahead)
│         ★  ✦✦  ★  ✦                │  Strong blue-shift forward
│                              ✦  ✦  │  Red-shifted trailing stars
└────────────────────────────────────┘
```

#### Design Rationale

- **Why 0.3c threshold?** Below this, SR effects are minimal (~6° aberration). Parallax becomes the primary "you're moving fast" indicator, so we enhance it for gameplay.
- **Why dual views?** Gives hard sci-fi authenticity (raw view) while maintaining playability (navigation view). Player can toggle based on preference.
- **Why artificial dust layers?** At low speeds, real parallax is invisible. Dust/particles provide immediate motion feedback without breaking physics.

#### Open Questions (To Resolve in Playtesting)

- [ ] Is 0.3c the right threshold, or should parallax start earlier for game feel?
- [ ] How disorienting is the raw SR view? Do players prefer it or avoid it?
- [ ] Should "navigation view" be default, with raw view as optional hardcore mode?
- [ ] Do we need gradual aberration transitions, or can we use discrete thresholds?

#### Boundary Glow Effect (Higgs Bubble Motion Cue)

The Higgs bubble provides a **physically-justified motion cue** that works at all speeds. Per [bubble-constraint.md](../../planned/future/bubble-constraint.md), mass cannot pass through the bubble boundary - only information (light/EM) crosses. This creates a visible effect when traveling through the interstellar medium.

**The Physics:**
- ISM contains ~1 atom/cm³ (hydrogen, helium, trace dust)
- At relativistic speeds, particles hit the bubble with enormous kinetic energy
- Since mass cannot pass, energy converts to EM radiation at the boundary
- The bubble's energy-dependent filtering means high-energy emissions (X-ray, gamma) are absorbed, but visible light escapes inward
- Result: a **glow at the forward boundary** proportional to speed and local ISM density

**Intensity by Speed:**

| Speed | Particle Flux | Boundary Effect |
|-------|---------------|-----------------|
| < 0.1c | Low | Invisible or very faint shimmer |
| 0.1c - 0.3c | Moderate | Subtle forward glow |
| 0.3c - 0.5c | High | Visible "bow glow" |
| 0.5c - 0.9c | Very high | Bright forward corona |
| > 0.9c | Extreme | Intense "plasma crown" |

**Intensity by Region:**

| Region | ISM Density | Boundary Effect |
|--------|-------------|-----------------|
| Deep interstellar | ~1 atom/cm³ | Standard glow |
| Hot ionized bubble | ~0.01 atom/cm³ | Very faint |
| Near star (stellar wind) | 10-100 atom/cm³ | Bright |
| Oort cloud / debris disk | Variable | Flickering, occasional flashes |
| Nebula / molecular cloud | 100-1000+ atom/cm³ | Intense "plasma rain" |

**Visual Design:**

```
Deep interstellar at 0.5c:
┌─────────────────────────────────────────────────────┐
│        · · · ·                                      │
│      ·╭━━━━━━━━━━━━━━━━━━━━━━━━━╮                  │
│     · │                         │                   │
│    ·  │     [SHIP INTERIOR]     │                   │
│     · │                         │                   │
│      ·╰━━━━━━━━━━━━━━━━━━━━━━━━━╯                  │
│         (subtle forward glow)                       │
└─────────────────────────────────────────────────────┘

Approaching star system at 0.3c:
┌─────────────────────────────────────────────────────┐
│      ✦·✦·✦·✦                                       │
│    ✦·╭━━━━━━━━━━━━━━━━━━━━━━━━━╮·                  │
│   ✦· │                         │ ·                  │
│  ✦·  │     [SHIP INTERIOR]     │  ·                 │
│   ✦· │                         │ ·                  │
│    ✦·╰━━━━━━━━━━━━━━━━━━━━━━━━━╯·                  │
│       (bright bow glow from stellar wind)           │
└─────────────────────────────────────────────────────┘

Inside nebula at any speed:
┌─────────────────────────────────────────────────────┐
│  ✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦✦                   │
│ ✦✦✦╭━━━━━━━━━━━━━━━━━━━━━━━━━╮✦✦✦                  │
│ ✦✦ │                         │ ✦✦                   │
│ ✦✦ │     [SHIP INTERIOR]     │ ✦✦                   │
│ ✦✦ │                         │ ✦✦                   │
│ ✦✦✦╰━━━━━━━━━━━━━━━━━━━━━━━━━╯✦✦✦                  │
│     (intense all-around glow, "plasma rain")        │
└─────────────────────────────────────────────────────┘
```

**Why This Works for Gameplay:**

1. **Always visible** - Unlike stellar parallax, boundary glow works at ALL speeds because you're always hitting something
2. **Direction-dependent** - Forward glow is strongest, gives heading feedback
3. **Contextual information** - Intensity tells you about local ISM density
4. **Physically justified** - Emerges directly from the Higgs bubble constraint (mass cannot pass)
5. **Narratively rich** - Crew culture could develop around the glow ("we wear a crown of fire")

**Narrative Opportunities:**

- **The Watch** - Crew tradition of watching the boundary glow during long journeys
- **Density Warnings** - Bright glow means dense region (navigation hazard? fuel opportunity?)
- **Color Shifts** - Different elements in ISM could tint the glow (hydrogen pink, helium blue)
- **Milestone Marker** - Glow intensity drop signals leaving a system's influence

**Integration with Other Effects:**

| Speed | Primary Cue | Secondary Cue |
|-------|-------------|---------------|
| < 0.1c | Boundary glow | Local parallax (if near planets) |
| 0.1c - 0.3c | Boundary glow + faint aberration | Stellar parallax for nearby stars |
| 0.3c - 0.9c | SR aberration + boundary glow | Color Doppler shift |
| > 0.9c | SR "starbow" + intense boundary glow | Doppler dominates |

**Open Questions:**

- [ ] Should glow color vary with ISM composition?
- [ ] Add occasional "flashes" from larger dust grain impacts?
- [ ] Should the glow be visible from interior views (window reflections)?
- [ ] Tie glow intensity to the trace mass absorption mechanic?

### Content Layer

The main interactive content.

```go
type ContentLayer interface {
    Draw(screen *ebiten.Image, camera *Camera)
    HandleInput(input *Input) bool
}

// Implementations
type SpaceContent struct {
    planets     []*Planet3D        // Tetra3D rendered spheres
    ship        *ShipModel         // Player ship (if visible)
    effects     []*SpaceEffect     // Engine glow, etc.
}

type IsometricContent struct {
    tiles       [][]*Tile
    entities    []*Entity
    camera      *IsoCamera
}

type GalaxyMapContent struct {
    stars       []*StarSystem
    routes      []*TradeRoute
    selection   *StarSystem
}
```

### UI Layer

HUD, panels, dialogue - always on top.

```go
type UILayer interface {
    AddPanel(panel *UIPanel)
    RemovePanel(id string)
    ShowDialogue(dialogue *Dialogue)
    Draw(screen *ebiten.Image)
}

type UIPanel struct {
    ID       string
    Position Vec2
    Size     Vec2
    Anchor   Anchor  // TopLeft, Center, BottomRight, etc.
    Draw     func(screen *ebiten.Image, bounds Rect)
}
```

## View Composition

### Space View (Exterior)

```
┌─────────────────────────────────────────────────┐
│  Ship Time: 47y 3mo          Galaxy: 2157 CE    │ ← UI: Time display
├─────────────────────────────────────────────────┤
│                                                 │
│                    🪐                           │ ← Content: 3D planet
│                                                 │
│            ✦  ✦    ✦      ✦    ✦               │
│         ✦      ✦  ✦    ✦      ✦   ✦            │ ← Background: Stars
│              ✦       ✦    ✦        ✦            │
│                                                 │
├─────────────────────────────────────────────────┤
│  v=0.3c  γ=1.05  [DECELERATE]                   │ ← UI: Controls
└─────────────────────────────────────────────────┘
```

### Bridge View (Interior)

```
┌─────────────────────────────────────────────────┐
│  Ship Time: 47y 3mo          Galaxy: 2157 CE    │ ← UI: Time display
├───────────────────────────────┬─────────────────┤
│       OBSERVATION DOME        │                 │
│    (Space view, smaller)      │    ARCHIVE      │ ← UI: Side panel
│         🌍 Earth              │    DIALOGUE     │
│            ✦  ✦               │                 │
├───────────────────────────────┴─────────────────┤
│                                                 │
│     [Helm]  [Comms]  [Systems]  [Galaxy Map]    │ ← Content: Isometric
│       👤      👤        👤                       │    bridge with
│    ═══════════════════════════════════════      │    crew stations
│                                                 │
└─────────────────────────────────────────────────┘
```

## Transitions

### Supported Effects

```go
type TransitionEffect int
const (
    TransitionNone      TransitionEffect = iota
    TransitionFade                        // Fade to black and back
    TransitionCrossfade                   // Blend between views
    TransitionWipe                        // Directional wipe
    TransitionZoom                        // Zoom in/out
)
```

### Transition Manager

```go
type TransitionManager struct {
    current     View
    next        View
    effect      TransitionEffect
    progress    float64  // 0.0 to 1.0
    duration    float64
}

func (tm *TransitionManager) Update(dt float64) {
    if tm.next == nil {
        return
    }

    tm.progress += dt / tm.duration
    if tm.progress >= 1.0 {
        tm.current.Exit(tm.next.Type())
        tm.current = tm.next
        tm.next = nil
        tm.progress = 0
    }
}

func (tm *TransitionManager) Draw(screen *ebiten.Image) {
    switch tm.effect {
    case TransitionFade:
        if tm.progress < 0.5 {
            tm.current.Draw(screen)
            drawFade(screen, tm.progress*2)  // Fade out
        } else {
            tm.next.Draw(screen)
            drawFade(screen, 2-tm.progress*2)  // Fade in
        }
    case TransitionCrossfade:
        tm.current.Draw(screen)
        tm.next.Draw(tm.buffer)
        drawBlend(screen, tm.buffer, tm.progress)
    }
}
```

## Implementation Plan

### Phase 1: Core Framework (2 days)

```
engine/
├── view/
│   ├── view.go           # View interface, ViewType enum
│   ├── manager.go        # ViewManager, transitions
│   ├── layer.go          # Layer interfaces
│   └── transition.go     # Transition effects
```

### Phase 2: Background Layer (1 day)

```
engine/
├── view/
│   └── background/
│       ├── space.go      # SpaceBackground implementation
│       └── stars.go      # Parallax star layers
```

### Phase 3: Integration (1 day)

- Wire into `cmd/game/main.go`
- Replace current ad-hoc rendering
- Test view transitions

## DrawCmd Integration

Views generate DrawCmds for the renderer:

```ailang
-- New DrawCmd variants for views
type DrawCmd =
    -- Existing...
    | Sprite(...)
    | RectRGBA(...)

    -- View system
    | ViewBackground(velocity: float, gr_intensity: float)
    | ViewTransition(effect: int, progress: float)
```

The engine interprets these and renders appropriately.

## Success Criteria

- [x] Views compose background + content + UI layers
- [ ] Transitions between views are smooth (implemented, not integrated into main game)
- [x] Space background renders with parallax stars
- [x] SR/GR effects apply to background layer
- [x] UI panels can be added/removed dynamically (BasicUILayer implemented)
- [x] 60fps maintained during transitions

## Sprint Progress

**Sprint ID**: view-system-v1
**Status**: In Progress (90% complete)
**Tracking**: [sprints/sprint_view-system-v1.json](../../../sprints/sprint_view-system-v1.json)

### Completed
- `engine/view/view.go` - View interface, ViewType enum
- `engine/view/layer.go` - Layer interfaces (Background, Content, UI)
- `engine/view/manager.go` - ViewManager to coordinate views
- `engine/view/transition.go` - Transition effects (fade, crossfade, wipe, zoom)
- `engine/view/easing.go` - Easing functions for smooth animations
- `engine/view/ui_layer.go` - BasicUILayer implementation
- `engine/view/background/space.go` - SpaceBackground with parallax stars
- `engine/view/view_test.go` - Unit tests
- Demo commands: `demo-planet-view`, `demo-sr-flyby`
- ViewSpace implementation

### Remaining
- [ ] Integrate ViewManager into `cmd/game/main.go` (DEFERRED: current GameMode system works)

### Notes
- Main game uses `GameMode` enum (ModeArrival, ModePlaying) instead of ViewManager
- ViewManager infrastructure is ready for future views
- Transition effects work but aren't wired into main game yet

## Dependencies

- **Requires**: Existing DrawCmd system, SR/GR shaders
- **Enables**: All game views, arrival sequence

## Next Steps After This

1. **space-background.md** - Detailed starfield implementation
2. **tetra3d-integration.md** - Add 3D rendering capability
3. **isometric-view.md** - Tile-based interior rendering
