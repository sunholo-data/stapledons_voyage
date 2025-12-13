# Engine Genericization

**Status:** Planned
**Priority:** P1 - Architecture
**Goal:** Make the Go/Ebiten engine fully reusable for any AILANG game

## Problem Statement

The engine layer (`engine/`) currently contains game-specific knowledge about Stapledon's Voyage. This violates the architectural principle that the engine should be a "dumb renderer" that could be swapped into another project with different AILANG code to make an entirely different game.

**Key insight:** `sim_gen/` is fine since it's generated from AILANG - the problem is game-specific content baked into `engine/`.

## Current State Analysis

### Game-Specific Content in Engine (Must Move)

| File | Game-Specific Content | Priority |
|------|----------------------|----------|
| [engine/view/dome_renderer.go](../../engine/view/dome_renderer.go) | Solar system with Neptune, Saturn, Jupiter, Mars, Earth; Saturn's rings; `DomeViewState` | P1 |
| [engine/render/deck_stack.go](../../engine/render/deck_stack.go) | Hardcoded 5 decks; DeckCore/Engineering/Culture/Habitat/Bridge names | P1 |
| [engine/render/draw.go:478-518](../../engine/render/draw.go#L478-L518) | `getBridgeSpriteColor()` with pilot/comms/engineer/scientist/captain roles | P2 |
| [engine/render/deck_preview.go](../../engine/render/deck_preview.go) | Deck names and colors | P2 |
| [engine/screenshot/screenshot.go](../../engine/screenshot/screenshot.go) | `ArrivalState`, `GetArrivalPlanetName()` | P3 |

### Generic Content in Engine (Keep)

| File | Content | Status |
|------|---------|--------|
| `engine/render/input.go` | `CaptureInput()` → `FrameInput` | Generic |
| `engine/render/draw.go` (DrawCmd switch) | Renders DrawCmd variants | Generic |
| `engine/assets/` | Sprite/audio/font loading | Generic |
| `engine/camera/` | Camera transforms | Generic |
| `engine/shader/` | Visual effects | Generic |
| `engine/display/` | Resolution/fullscreen | Generic |

## Target Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  Game (e.g., Stapledon's Voyage)                                │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  sim/*.ail - World, NPC, Ship, Decks, Planets            │   │
│  │  (ALL game concepts defined here)                         │   │
│  └──────────────────────────────────────────────────────────┘   │
│              ↓ generates                                         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  sim_gen/*.go - Generated types (World, DrawCmd, etc.)   │   │
│  │  (OK - generated from AILANG, contains game types)        │   │
│  └──────────────────────────────────────────────────────────┘   │
│              ↓ uses                                              │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  game_views/*.go - Game-specific renderers (NEW)          │   │
│  │  DomeRenderer, DeckStackRenderer moved here               │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
              ↓ uses
┌─────────────────────────────────────────────────────────────────┐
│  Generic Engine (REUSABLE)                                      │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  engine/render/ - DrawCmd rendering (generic)             │   │
│  │  engine/assets/ - Asset loading (generic)                 │   │
│  │  engine/camera/ - Camera transforms (generic)             │   │
│  │  engine/shader/ - Visual effects (generic)                │   │
│  │  engine/display/ - Window management (generic)            │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  Engine only knows: DrawCmd, FrameInput, FrameOutput, Camera    │
│  Engine does NOT know: World, Deck, Ship, Planet, NPC           │
└─────────────────────────────────────────────────────────────────┘
```

## The Key Question

Before adding code to `engine/`:

> **Does this affect gameplay outcomes?**

| Answer | Where to Put It |
|--------|-----------------|
| YES | AILANG (`sim/*.ail`) |
| NO, but game-specific visual | `game_views/*.go` (NEW) |
| NO, generic rendering | `engine/*.go` |

## Implementation Plan

### Phase 1: Create game_views Layer

1. Create `game_views/` directory at project root
2. Move game-specific renderers there:
   - `engine/view/dome_renderer.go` → `game_views/dome_renderer.go`
   - `engine/render/deck_stack.go` → `game_views/deck_stack.go`
   - `engine/render/deck_preview.go` → `game_views/deck_preview.go`

### Phase 2: Remove Sprite ID Semantics from Engine

1. Remove `getBridgeSpriteColor()` from `engine/render/draw.go`
2. Replace with: lookup from asset manifest or pass color in DrawCmd
3. Engine should only know "sprite ID N" → "render these pixels"

### Phase 3: Abstract Game Types from Engine

1. Create interface in engine for what it needs:
   ```go
   // engine/sim.go
   type GameSimulation interface {
       Init(seed int64) interface{}
       Step(state interface{}, input FrameInput) (interface{}, FrameOutput)
   }
   ```

2. Engine calls this interface, doesn't import `sim_gen.World` directly

### Phase 4: Move Planet Data to AILANG

1. Solar system definitions → `sim/celestial.ail` (already partially done)
2. Remove `createSolarSystem()` from dome_renderer.go
3. Planets rendered via DrawCmds from AILANG

## Acceptance Criteria

1. **Engine imports only generic types from sim_gen:**
   - `FrameInput`, `FrameOutput`, `DrawCmd*`, `Camera`, `Coord`
   - NOT: `World`, `DeckType`, `ArrivalState`, `DomeViewState`, `Planet`, `NPC`

2. **No hardcoded game concepts in engine:**
   - No deck names (Core, Engineering, Bridge, etc.)
   - No crew roles (pilot, comms, scientist, etc.)
   - No planet names (Saturn, Earth, etc.)
   - No sprite ID semantic ranges (1000-1099 = bridge tiles)

3. **game_views/ exists and contains:**
   - All game-specific rendering helpers
   - Types that reference sim_gen game types

4. **Reusability test:**
   - Could theoretically swap `sim/*.ail` + `sim_gen/` + `game_views/` for a different game
   - `engine/` would work unchanged

## Migration Checklist

### Files to Move to game_views/

- [x] `engine/view/dome_renderer.go` → `game_views/dome_renderer.go`
- [x] `engine/render/deck_stack.go` → `game_views/deck_stack.go`
- [x] `engine/render/deck_preview.go` → `game_views/deck_preview.go`
- [x] `engine/render/deck_transition.go` → `game_views/deck_transition.go`

### Code to Remove from Engine

- [ ] `getBridgeSpriteColor()` in draw.go (P2 - bridge sprite colors hardcoded)
- [ ] Sprite ID range comments (bridge tiles, crew sprites)
- [ ] `registerAnimations()` hardcoded sprite ID ranges

### Types to Stop Importing in Engine

- [x] `sim_gen.DeckType` (moved with deck_stack.go)
- [x] `sim_gen.DomeViewState` (moved with dome_renderer.go)
- [ ] `sim_gen.ArrivalState` (in screenshot.go - P3, testing only)
- [ ] `sim_gen.DeckInfo` (moved with deck_stack.go)
- [ ] `sim_gen.World` (still referenced in some places)

## Notes

- `sim_gen/` is fine because it's generated - game types belong there
- Engine can render ANY DrawCmd without knowing what game produced it
- Think of engine like a GPU driver - it doesn't know Minecraft from Fortnite
