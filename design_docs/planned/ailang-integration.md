# AILANG Integration Architecture

**Status**: Planned
**Target**: v0.5.0 (tracks AILANG release)
**Priority**: P0 - Critical
**Estimated**: Ongoing (tied to AILANG releases)
**Dependencies**: AILANG v0.5.0+ with Go codegen

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Infrastructure |
| Civilization Simulation | + | +1 | AILANG enables complex sim logic |
| Philosophical Depth | N/A | 0 | Infrastructure |
| Ship & Crew Life | N/A | 0 | Infrastructure |
| Legacy Impact | N/A | 0 | Infrastructure |
| Hard Sci-Fi Authenticity | + | +1 | Deterministic simulation via pure functions |
| **Net Score** | | **+2** | **Decision: Move forward (critical infrastructure)** |

**Feature type:** Infrastructure (enables everything else)

## Problem Statement

Stapledon's Voyage uses AILANG for all simulation logic. This document defines how we integrate AILANG-generated Go code with our Ebiten game engine.

**Current State:**
- Using mock `sim_gen/` package (hand-written Go)
- Awaiting AILANG v0.5.0 with Go codegen

**Impact:**
- Foundation for all game logic
- Determines build pipeline
- Affects iteration speed during development

## Goals

**Primary Goal:** Seamlessly integrate AILANG-generated Go code with our game engine, maintaining the same interface as the mock implementation.

**Success Metrics:**
- `make sim` produces working Go code from `sim/*.ail`
- Generated code passes all existing tests
- No changes required to `engine/` layer
- Build time < 30 seconds for full recompile

## Solution Design

### Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Build Pipeline                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   sim/*.ail ──► ailang compile ──► sim_gen/*.go ──► go build   │
│                     │                    │                      │
│                     │                    │                      │
│                     ▼                    ▼                      │
│              Type checking         Links with engine/           │
│              Effect checking       and cmd/game/                │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Layer Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     cmd/game/main.go                            │
│                     (Ebiten game loop)                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   engine/render/         engine/input/         engine/assets/   │
│   (DrawCmd → pixels)     (keys → FrameInput)   (load sprites)   │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│                      sim_gen/*.go                               │
│                 (AILANG-generated code)                         │
│                                                                 │
│         InitWorld()    Step()    Types (World, DrawCmd, etc)    │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│                       sim/*.ail                                 │
│                   (AILANG source files)                         │
│                                                                 │
│         protocol.ail    world.ail    step.ail    npc.ail       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### File Structure

```
stapledons_voyage/
├── sim/                          # AILANG source (manually edited)
│   ├── protocol.ail              # FrameInput, FrameOutput, DrawCmd
│   ├── world.ail                 # World, Entity, Planet, Civ types
│   ├── step.ail                  # init_world(), step() exports
│   ├── npc.ail                   # NPC AI logic
│   ├── civ.ail                   # Civilization simulation
│   └── journey.ail               # Journey planning logic
│
├── sim_gen/                      # Generated Go (never manually edit)
│   ├── protocol.go               # Generated from protocol.ail
│   ├── world.go                  # Generated from world.ail
│   ├── step.go                   # Generated from step.ail
│   └── ...
│
├── engine/                       # Pure Go (manually edited)
│   ├── render/                   # DrawCmd → Ebiten
│   ├── input/                    # Keyboard/Mouse → FrameInput
│   └── assets/                   # Sprite/sound loading
│
└── cmd/
    └── game/
        └── main.go               # Ebiten entry point
```

### Build Commands

**Makefile targets:**

```makefile
# Compile AILANG to Go
sim:
	ailang compile --emit-go \
		--package-name sim_gen \
		--out sim_gen/ \
		sim/*.ail

# Build game (depends on sim)
game: sim
	go build -o bin/game ./cmd/game

# Run with mock (no ailc needed)
game-mock:
	go build -tags mock -o bin/game ./cmd/game

# Full clean + rebuild
rebuild: clean sim game
```

### Generated Code Contract

AILANG v0.5.0+ guarantees this output structure:

**ADT Types:**
```go
// From: type DrawCmd = | Sprite(...) | Rect(...) | Text(...)
type DrawCmdKind int
const (
    DrawCmdKindSprite DrawCmdKind = iota
    DrawCmdKindRect
    DrawCmdKindText
)
type DrawCmd struct {
    Kind   DrawCmdKind
    Sprite *DrawCmdSprite
    Rect   *DrawCmdRect
    Text   *DrawCmdText
}
```

**Exported Functions:**
```go
// From: export func init_world(seed: int) -> World
func InitWorld(seed int64) World

// From: export func step(world: World, input: FrameInput) -> (World, FrameOutput) ! {RNG, Debug}
func Step(world World, input FrameInput) (World, FrameOutput, error)
```

### Engine Integration

**Current mock interface (must be preserved):**

```go
// engine/render/render.go
func RenderFrame(screen *ebiten.Image, output sim_gen.FrameOutput) {
    for _, cmd := range output.DrawCmds {
        switch cmd.Kind {
        case sim_gen.DrawCmdKindSprite:
            drawSprite(screen, cmd.Sprite)
        case sim_gen.DrawCmdKindRect:
            drawRect(screen, cmd.Rect)
        case sim_gen.DrawCmdKindText:
            drawText(screen, cmd.Text)
        }
    }
}

// engine/input/input.go
func CaptureInput() sim_gen.FrameInput {
    return sim_gen.FrameInput{
        Tick:        currentTick,
        Dt:          1.0 / 60.0,
        KeysPressed: getKeysPressed(),
        MousePos:    getMousePos(),
    }
}
```

### Effect Handling

**RNG Effect (v0.5.1+):**
```go
// Set deterministic seed for replay/testing
os.Setenv("AILANG_SEED", "42")

// RNG calls in AILANG will use this seed
world := sim_gen.InitWorld(42)
```

**Debug Effect (v0.5.1+):**
```go
world, output, err := sim_gen.Step(world, input)
if err != nil {
    // Effect capability error
    log.Fatal(err)
}

// Access debug output
for _, log := range output.Debug.Logs {
    fmt.Printf("[%s] %s\n", log.Location, log.Message)
}
for _, assert := range output.Debug.Assertions {
    if !assert.Passed {
        fmt.Printf("ASSERTION FAILED: %s at %s\n", assert.Message, assert.Location)
    }
}
```

### Migration Strategy

**Phase 1: Parallel Development**
- Keep mock `sim_gen/` working
- Develop `sim/*.ail` alongside
- Use `make game-mock` for day-to-day dev

**Phase 2: Codegen Validation**
- When AILANG v0.5.0 ships, run `make sim`
- Compare generated code with mock
- Fix any type mismatches

**Phase 3: Switch Over**
- Remove mock code
- `make game` becomes default
- Keep mock target for emergency fallback

### Implementation Plan

**Phase 1: Protocol Definition** (~1 day)
- [ ] Define `sim/protocol.ail` matching mock types exactly
- [ ] Verify with `ailang check sim/protocol.ail`

**Phase 2: World Types** (~2 days)
- [ ] Define `sim/world.ail` with all game types
- [ ] Define `sim/npc.ail`, `sim/civ.ail`, etc.

**Phase 3: Step Function** (~2 days)
- [ ] Implement `init_world` in AILANG
- [ ] Implement `step` in AILANG
- [ ] Add effect annotations

**Phase 4: Integration** (~1 day)
- [ ] Run `ailang compile --emit-go`
- [ ] Verify generated code builds
- [ ] Run tests against generated code

### Files to Modify/Create

**AILANG source (new):**
- `sim/protocol.ail` - Shared types (~100 LOC)
- `sim/world.ail` - World state (~200 LOC)
- `sim/step.ail` - Game loop (~150 LOC)

**Build system (modify):**
- `Makefile` - Add `sim` target
- `.github/workflows/` - CI integration

## Success Criteria

- [ ] `ailang check sim/*.ail` passes
- [ ] `ailang compile --emit-go` produces valid Go
- [ ] Generated code builds with `go build`
- [ ] All existing tests pass with generated code
- [ ] Performance within 10% of mock implementation

## Testing Strategy

**Type compatibility:**
- Generated types match engine expectations
- DrawCmd switch statements compile

**Functional correctness:**
- Same seed produces same world
- 100 ticks match mock output exactly

**Performance:**
- Benchmark step() function
- Compare against mock baseline

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| AILANG v0.5.0 delayed | High | Keep mock functional; track AILANG roadmap |
| Generated code slower than mock | Med | Profile; use `extern func` for hot paths |
| Type mismatch at boundary | High | Define protocol types first; verify early |
| Effect handling differs from expected | Med | Test with AILANG beta builds |

## References

- [consumer-contract-v0.5.md](../../ailang_resources/consumer-contract-v0.5.md) - AILANG contract
- [CLAUDE.md](../../CLAUDE.md) - Build commands and architecture
- [DEVELOPMENT.md](../../DEVELOPMENT.md) - Data flow documentation

## Future Work

- Hot reload of AILANG during development
- AILANG → WebAssembly for browser builds
- AILANG debug visualizer integration
- LSP integration for editor support

---

**Document created**: 2025-12-02
**Last updated**: 2025-12-02
