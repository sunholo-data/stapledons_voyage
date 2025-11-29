# Stapledons Voyage Architecture

**Version:** 0.1.0
**Status:** Planned
**Priority:** P0 (High)
**Complexity:** Complex
**AILANG Workarounds:** Module imports, duplicate types

## Related Documents

- [Engine Layer Design](engine-layer.md) - Go/Ebiten implementation details
- [Evaluation System Design](eval-system.md) - Benchmarks and scenarios

## Overview

Stapledons Voyage is a 2D game engine that serves as an integration test for AILANG. The architecture cleanly separates simulation logic (AILANG) from rendering/IO (Go/Ebiten).

## Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    User / Display                        │
└─────────────────────────────────────────────────────────┘
                          ▲
                          │ Screen pixels
                          ▼
┌─────────────────────────────────────────────────────────┐
│              ENGINE LAYER (Go/Ebiten)                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │    Input     │  │   Render     │  │   Assets     │   │
│  │ CaptureInput │  │ RenderFrame  │  │ AssetManager │   │
│  └──────────────┘  └──────────────┘  └──────────────┘   │
└─────────────────────────────────────────────────────────┘
        │                    ▲
        │ FrameInput         │ FrameOutput
        ▼                    │
┌─────────────────────────────────────────────────────────┐
│             SIMULATION LAYER (AILANG → Go)               │
│  ┌──────────────────────────────────────────────────┐   │
│  │                   sim_gen/                        │   │
│  │    InitWorld(seed) → World                        │   │
│  │    Step(world, input) → (World, FrameOutput)      │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
        ▲
        │ Generated from
        │
┌─────────────────────────────────────────────────────────┐
│              SOURCE LAYER (AILANG)                       │
│  sim/protocol.ail  - Types: FrameInput, FrameOutput     │
│  sim/world.ail     - Types: World, Entity, Tile         │
│  sim/step.ail      - Functions: init_world, step        │
└─────────────────────────────────────────────────────────┘
```

## Data Flow

Each frame follows this cycle:

```
1. Ebiten calls Update()
2. CaptureInput() → FrameInput (mouse, keyboard state)
3. sim_gen.Step(world, input) → (newWorld, output)
4. Store newWorld for next frame
5. Ebiten calls Draw()
6. RenderFrame(screen, output) → pixels on screen
```

## Key Design Decisions

### Pure Simulation

The simulation layer is pure: given the same `(World, FrameInput)`, it always produces the same `(World, FrameOutput)`. No side effects, no IO, no randomness (until RNG effect in v0.5.1).

This enables:
- Deterministic replay
- Headless testing
- Time-travel debugging (future)

### Discriminator-Based ADTs

AILANG generates Go structs with discriminator enums instead of interfaces:

```go
type DrawCmd struct {
    Kind   DrawCmdKind
    Rect   *DrawCmdRect
    Sprite *DrawCmdSprite
    Text   *DrawCmdText
}
```

This provides:
- No interface dispatch in hot loops
- Cache-friendly memory layout
- Predictable switch-based pattern matching

### Thin Engine Layer

The Go engine layer is intentionally minimal:
- No game logic
- No state beyond current `World`
- Only IO bridging (input capture, rendering)

All game logic lives in AILANG, making this a true integration test.

## Directory Structure

```
stapledons_voyage/
├── sim/                    # AILANG source (manually edited)
│   ├── protocol.ail        # Stable API types
│   ├── world.ail           # World state types
│   └── step.ail            # Core game loop
├── sim_gen/                # Generated Go (never edit)
│   └── *.go                # Compiled from sim/*.ail
├── engine/                 # Go engine layer
│   ├── render/             # Input capture, rendering
│   ├── scenario/           # Evaluation runner
│   └── bench/              # Benchmarks
├── cmd/
│   ├── game/               # Main game executable
│   └── eval/               # Benchmark runner
├── design_docs/            # Design documentation
└── ailang_resources/       # AILANG contract, refs
```

## Build Pipeline

```
sim/*.ail  →  ailang compile  →  sim_gen/*.go  →  go build  →  bin/game
```

The Makefile enforces this:
- `make sim` - Compile AILANG to Go
- `make game` - Build executable (depends on sim)
- `make run` - Run game (depends on sim)
- `make eval` - Run benchmarks and scenarios

## AILANG Constraints

**Known limitations to work around:**

| Limitation | Impact | Workaround |
|------------|--------|------------|
| Module imports not working | Cannot share types across files | Duplicate type definitions in each .ail file |
| No RNG effect | World generation deterministic | Use seed parameter for pseudo-randomness |
| Lists only (no arrays) | O(n) tile access | Keep world small (64x64 max) for v0.1.0 |
| Recursion depth limits | Cannot iterate all tiles | Use bounded recursion patterns |

**Reported to AILANG core:** See [ailang_resources/](../../../ailang_resources/) for feedback sent.

## Success Criteria

### Architecture
- [ ] Three-layer separation (AILANG → Go codegen → Ebiten)
- [ ] Pure simulation layer (deterministic, no side effects)
- [ ] Thin engine layer (input capture + rendering only)
- [ ] Clean code generation boundary (never edit sim_gen/)

### Build Pipeline
- [ ] `make sim` compiles AILANG to Go
- [ ] `make game` produces working executable
- [ ] `make run` launches game window
- [ ] `make eval` generates evaluation report

### Integration
- [ ] FrameInput/FrameOutput bridge working
- [ ] DrawCmd rendering functional
- [ ] World state persists between frames

## Version History

| Version | Changes |
|---------|---------|
| 0.1.0 | Initial architecture, basic rendering, 2x2 tile world |
