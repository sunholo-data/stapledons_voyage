# Architecture Rules Reference

Detailed rules for Stapledon's Voyage three-layer architecture.

## Layer Responsibilities

### 1. Source Layer (`sim/*.ail`)

**Purpose:** Define game logic in AILANG

**Contains:**
- Type definitions (World, Entity, Tile, NPC, etc.)
- Pure functions (init_world, step)
- Protocol types (FrameInput, FrameOutput, DrawCmd)

**Rules:**
- No side effects
- No IO operations
- Deterministic given same inputs
- Will be compiled to Go via `ailang compile --emit-go`

### 2. Simulation Layer (`sim_gen/*.go`)

**Purpose:** Generated Go code from AILANG (or mock until AILANG ships)

**Contains (when mock):**
- All type definitions matching protocol.ail
- InitWorld function
- Step function
- Game logic (NPC AI, actions, world updates)
- DrawCmd generation

**Rules:**
- NO manual editing when using AILANG compiler
- Currently mock = hand-written Go mimicking AILANG output
- NO rendering imports (no ebiten, no image packages)
- NO file IO
- Must export: `InitWorld`, `Step`, `World`, `FrameInput`, `FrameOutput`, `DrawCmd`

**Allowed imports:**
```go
import (
    "math"      // OK - pure math
    "sort"      // OK - pure algorithms
)
```

**Forbidden imports:**
```go
import (
    "github.com/hajimehoshi/ebiten/v2"  // NO - rendering
    "image"                              // NO - rendering
    "os"                                 // NO - IO
    "net"                                // NO - IO
)
```

### 3. Engine Layer (`engine/*.go`)

**Purpose:** IO bridging between simulation and Ebiten

**Contains:**
- `engine/render/` - Input capture, frame rendering
- `engine/assets/` - Sprite, font, sound loading
- `engine/display/` - Window config, fullscreen
- `engine/scenario/` - Evaluation/testing harness
- `engine/bench/` - Benchmarks

**Rules:**
- NO game logic
- NO decision-making code
- NO World manipulation (except storing current World)
- Just bridges FrameInput → sim_gen.Step → FrameOutput → pixels

**What belongs here:**
```go
// YES - Input capture
func CaptureInput() sim_gen.FrameInput { ... }

// YES - Rendering
func RenderFrame(screen *ebiten.Image, output sim_gen.FrameOutput) { ... }

// YES - Asset loading
func LoadSprite(name string) *ebiten.Image { ... }
```

**What does NOT belong:**
```go
// NO - Game logic
func UpdateNPC(npc *sim_gen.NPC) { ... }

// NO - World manipulation
func SpawnEntity(world *sim_gen.World) { ... }

// NO - Decision making
if player.Health < 10 { ... }
```

### 4. Entry Layer (`cmd/*.go`)

**Purpose:** Wire everything together

**Contains:**
- `cmd/game/main.go` - Main executable
- `cmd/eval/main.go` - Benchmark runner

**Rules:**
- Minimal code (wiring only)
- Should be < 200 lines per file
- Creates Game struct, runs Ebiten loop
- Calls CaptureInput → Step → RenderFrame

## File Placement Guide

| Type of Code | Correct Location |
|--------------|------------------|
| NPC behavior | `sim_gen/` (mock) or `sim/*.ail` |
| Building logic | `sim_gen/` or `sim/*.ail` |
| World generation | `sim_gen/` or `sim/*.ail` |
| Input handling | `engine/render/input.go` |
| Sprite rendering | `engine/render/draw.go` |
| Asset loading | `engine/assets/` |
| Window config | `engine/display/` |
| Main game loop | `cmd/game/main.go` |

## Common Violations

### 1. Game Logic in Engine

**Wrong:**
```go
// engine/render/draw.go
func (g *Game) Update() {
    if g.world.Player.Health < 10 {
        g.world.Player.State = "fleeing"  // Game logic!
    }
}
```

**Right:**
```go
// sim_gen/player.go
func updatePlayerState(player Player) Player {
    if player.Health < 10 {
        return Player{...State: "fleeing"...}
    }
    return player
}
```

### 2. Rendering in Simulation

**Wrong:**
```go
// sim_gen/render.go
import "github.com/hajimehoshi/ebiten/v2"

func RenderWorld(screen *ebiten.Image) { ... }
```

**Right:**
```go
// sim_gen/step.go
func Step(...) FrameOutput {
    return FrameOutput{
        DrawCmds: []DrawCmd{...},  // Data only, no rendering
    }
}
```

### 3. Large cmd/ Files

**Wrong:**
```go
// cmd/game/main.go (500+ lines)
func main() {
    // Lots of inline logic
}
```

**Right:**
```go
// cmd/game/main.go (<100 lines)
func main() {
    game := engine.NewGame()
    ebiten.RunGame(game)
}
```

## Checking Boundaries

Use marker comments when code must cross boundaries:

```go
// allowed: World read for rendering
drawCmds := output.DrawCmds
```

These comments tell the boundary checker to skip false positives.

## Migration Path

When AILANG compiler ships:

1. Write/update `sim/*.ail` files
2. Run `ailang compile --emit-go -o sim_gen/`
3. Generated code replaces mock `sim_gen/*.go`
4. Engine layer unchanged
5. Game works with real AILANG
