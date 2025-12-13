# Simple Isometric Walk Demo

## Status
- **Status**: Planned
- **Priority**: P1 (Validates core isometric system)
- **Estimated**: 1 day

## Goal

Create a **minimal, clean isometric demo** that:
1. Uses proper isometric assets (64x32 tiles, 128x48 entities)
2. Tests the screen-aligned WASD movement
3. Bypasses all bridge_view.go complexity
4. Directly wires AILANG → Renderer (no intermediate views)

## Why Start Fresh

The current `bridge_view.go` has accumulated complexity:
- Pre-rendered floor cache
- Dome renderer integration
- 3D planet rendering
- Square (1024x1024) sprites that don't tessellate

A fresh demo uses:
- **Only** proper isometric assets (iso_tiles/, iso_entities/)
- **Direct** AILANG → DrawCmd → Renderer flow
- **Simple** demo template (no view abstraction)

## Available Assets

Already in the codebase with correct dimensions:

| Asset | Sprite ID | Size | Type |
|-------|-----------|------|------|
| water.png | 1 | 64x32 | Tile |
| forest.png | 2 | 64x32 | Tile |
| desert.png | 3 | 64x32 | Tile |
| mountain.png | 4 | 64x32 | Tile |
| player.png | 105 | 128x48 (4×32x48 frames) | Entity |
| npc_red.png | 100 | 128x48 (4×32x48 frames) | Entity |
| npc_green.png | 101 | 128x48 (4×32x48 frames) | Entity |
| npc_blue.png | 102 | 128x48 (4×32x48 frames) | Entity |

## Implementation Plan

### 1. New AILANG Module: `sim/iso_demo.ail`

Minimal state and functions:

```ailang
module sim/iso_demo

import sim/protocol (Coord, DrawCmd, IsoTile, IsoEntity, FrameInput, KeyEvent)
import std/option (Option, Some, None)

-- Minimal state for isometric demo
export type IsoWalkState = {
    playerX: int,
    playerY: int,
    gridWidth: int,
    gridHeight: int
}

-- Initialize state
export pure func initIsoDemo() -> IsoWalkState {
    { playerX: 4, playerY: 4, gridWidth: 8, gridHeight: 8 }
}

-- Isometric movement (screen-aligned WASD)
type IsoMove = { dx: int, dy: int }
pure func isoMoveUp() -> IsoMove { { dx: -1, dy: -1 } }
pure func isoMoveDown() -> IsoMove { { dx: 1, dy: 1 } }
pure func isoMoveLeft() -> IsoMove { { dx: -1, dy: 1 } }
pure func isoMoveRight() -> IsoMove { { dx: 1, dy: -1 } }

-- Key codes (Ebiten)
pure func keyW() -> int { 22 }
pure func keyA() -> int { 0 }
pure func keyS() -> int { 18 }
pure func keyD() -> int { 3 }

-- Process input
export pure func stepIsoDemo(state: IsoWalkState, input: FrameInput) -> IsoWalkState {
    let move = getIsoMovement(input);
    match move {
        Some(m) => tryMove(state, m),
        None => state
    }
}

-- Render floor tiles + player
export pure func renderIsoDemo(state: IsoWalkState) -> [DrawCmd] {
    concat(renderFloor(state), renderPlayer(state))
}

-- Helper: render 8x8 grid with alternating tiles
pure func renderFloor(state: IsoWalkState) -> [DrawCmd] { ... }

-- Helper: render player entity
pure func renderPlayer(state: IsoWalkState) -> [DrawCmd] {
    [IsoEntity("player", { x: state.playerX, y: state.playerY }, 0.0, 0.0, 0, 105, 3)]
}
```

### 2. New Demo: `cmd/demo-iso-walk/main.go`

Based on demo template, directly wires AILANG:

```go
type DemoGame struct {
    renderer   *render.Renderer
    isoState   *sim_gen.IsoWalkState
    frameCount int
}

func (g *DemoGame) Update() error {
    // Capture input
    input := render.CaptureInputWithCamera(...)

    // Step AILANG
    g.isoState = sim_gen.StepIsoDemo(g.isoState, input)

    return nil
}

func (g *DemoGame) Draw(screen *ebiten.Image) {
    // Get DrawCmds from AILANG
    cmds := sim_gen.RenderIsoDemo(g.isoState)

    // Render directly
    out := sim_gen.FrameOutput{
        Draw:   cmds,
        Camera: &sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0},
    }
    g.renderer.RenderFrame(screen, out)

    // HUD
    g.drawHUD(screen)
}
```

### 3. Files to Create

| File | Purpose |
|------|---------|
| `sim/iso_demo.ail` | Minimal AILANG state/step/render |
| `cmd/demo-iso-walk/main.go` | Demo entry point |

### 4. Files to NOT Touch

- `engine/view/bridge_view.go` - Leave as-is (may delete later)
- `sim/bridge.ail` - Leave as-is (complex, not needed for demo)
- Bridge assets (1024x1024) - Not used

## Success Criteria

- [ ] Player sprite renders at correct isometric position
- [ ] Tiles tessellate properly (64x32 diamonds)
- [ ] W moves player toward screen top
- [ ] S moves player toward screen bottom
- [ ] A moves player toward screen left
- [ ] D moves player toward screen right
- [ ] Player stays within grid bounds
- [ ] Animation frames cycle when moving (stretch goal)

## Camera Setup

For an 8x8 grid centered on screen:
- Camera at world (0, 0)
- Grid tiles from (0,0) to (7,7)
- Player starts at (4, 4)
- Screen center offset handled by renderer

## Testing

```bash
# Build and run
make sim
go build -o bin/demo-iso-walk ./cmd/demo-iso-walk
bin/demo-iso-walk

# Screenshot for CI
bin/demo-iso-walk --screenshot 60 --output out/iso-walk.png
```
