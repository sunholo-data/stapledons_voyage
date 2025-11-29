# Sprint 003: Player Interaction

**Status:** Completed
**Goal:** Implement click-to-select tile interaction with visual feedback
**Estimated Effort:** 1-2 sessions
**AILANG Dependency:** None (mock implementation)
**Design Doc:** [player-interaction.md](../design_docs/planned/v0_2_0/player-interaction.md)

## Context

Sprints 001-002 established the engine infrastructure and camera system. The game now renders a centered world with proper coordinate transforms. This sprint adds the first player interaction: clicking to select tiles.

**Key Dependency:** Uses `ScreenToWorld` transform from Sprint 002 to convert mouse clicks to world/tile coordinates.

## Success Criteria

- [x] Click on tile highlights it with yellow overlay
- [x] Click elsewhere moves highlight to new tile
- [x] Click outside world clears selection
- [x] Highlight renders on correct tile (uses camera transform)
- [x] Selection persists across frames (until new click)
- [x] `make eval-mock` still passes

## Tasks

### Phase 1: Selection Types in Mock (P0)

Add Selection type to sim_gen mock.

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/types.go` | Add Selection interface and concrete types |
| 1.2 | `sim_gen/types.go` | Add Selection field to World struct |
| 1.3 | `sim_gen/funcs.go` | Update InitWorld to set NoSelection |
| 1.4 | Build | Verify `make game-mock` compiles |

**Selection types:**
```go
type Selection interface {
    isSelection()
}

type SelectionNone struct{}
func (SelectionNone) isSelection() {}

type SelectionTile struct {
    X, Y int
}
func (SelectionTile) isSelection() {}
```

**Estimated:** 0.5 session

### Phase 2: Click Detection (P0)

Detect mouse clicks and convert to tile coordinates.

| Task | File | Description |
|------|------|-------------|
| 2.1 | `engine/input/input.go` | Create input package with click detection |
| 2.2 | `engine/input/input.go` | IsJustPressed for left mouse button |
| 2.3 | `engine/input/input.go` | GetMousePosition returns screen coords |
| 2.4 | Test | Unit tests for input detection |

**Key functions:**
```go
func IsMouseJustPressed() bool
func GetMousePosition() (x, y int)
```

**Note:** Uses Ebiten's `inpututil.IsMouseButtonJustPressed` for edge detection.

**Estimated:** 0.5 session

### Phase 3: Selection Logic in Step (P0)

Process clicks and update selection in Step function.

| Task | File | Description |
|------|------|-------------|
| 3.1 | `sim_gen/funcs.go` | Add ClickedThisFrame to FrameInput |
| 3.2 | `sim_gen/funcs.go` | Process click in Step function |
| 3.3 | `sim_gen/funcs.go` | Convert screen→world→tile coords |
| 3.4 | `sim_gen/funcs.go` | Update World.Selection on valid click |
| 3.5 | Test | Unit tests for tile selection |

**Click processing:**
```go
func Step(world World, input FrameInput) (World, FrameOutput, error) {
    // ... existing logic ...

    // Process selection
    newSelection := world.Selection
    if input.ClickedThisFrame {
        // Screen → World (using camera inverse)
        // World → Tile (divide by TileSize)
        tileX := int(worldX / TileSize)
        tileY := int(worldY / TileSize)
        if inBounds(tileX, tileY, world.Planet) {
            newSelection = SelectionTile{X: tileX, Y: tileY}
        } else {
            newSelection = SelectionNone{}
        }
    }

    newWorld := World{
        // ... include newSelection ...
    }
}
```

**Estimated:** 0.5 session

### Phase 4: Selection Rendering (P0)

Render selection highlight as a DrawCmd.

| Task | File | Description |
|------|------|-------------|
| 4.1 | `sim_gen/funcs.go` | Generate highlight DrawCmdRect for selection |
| 4.2 | `sim_gen/funcs.go` | Set Z=1 so highlight draws on top |
| 4.3 | `engine/render/draw.go` | Verify color index 4 = yellow highlight |
| 4.4 | Test | Visual verification with `make run-mock` |

**Highlight rendering:**
```go
// In Step, after tile draw commands:
if sel, ok := newSelection.(SelectionTile); ok {
    highlightCmd := DrawCmdRect{
        X:     float64(sel.X) * TileSize,
        Y:     float64(sel.Y) * TileSize,
        W:     TileSize,
        H:     TileSize,
        Color: 4,  // Yellow highlight
        Z:     1,  // On top of tiles
    }
    drawCmds = append(drawCmds, highlightCmd)
}
```

**Estimated:** 0.5 session

### Phase 5: Integration & Verification (P1)

Connect input to game loop and verify everything works.

| Task | File | Description |
|------|------|-------------|
| 5.1 | `cmd/game/main.go` | Update CaptureInput to detect clicks |
| 5.2 | `cmd/game/main.go` | Pass click state to Step |
| 5.3 | Test | Run `make eval-mock` |
| 5.4 | Test | Manual click testing |
| 5.5 | Update | Mark sprint complete |

**Manual test cases:**
- Click on tile → yellow highlight appears
- Click different tile → highlight moves
- Click outside world → highlight disappears
- Hold mouse button → only selects on first frame

**Estimated:** 0.5 session

## Technical Details

### Coordinate Conversion Chain

```
Screen (mouse)     World (pixels)      Tile (grid)
┌─────────────┐   ┌─────────────┐    ┌─────────────┐
│ (320, 240)  │ → │ (256, 256)  │ → │  (32, 32)   │
│ mouse click │   │ camera.ScreenToWorld │ ÷ TileSize   │
└─────────────┘   └─────────────┘    └─────────────┘
```

### State Flow

```
FrameInput.ClickedThisFrame = true
    ↓
Step() processes click
    ↓
World.Selection = SelectionTile{X, Y}
    ↓
Step() generates DrawCmdRect with Color=4, Z=1
    ↓
Renderer draws yellow highlight on top
```

### Edge Cases

| Case | Behavior |
|------|----------|
| Click at (0,0) | Select tile (0,0) |
| Click at world edge | Select edge tile |
| Click outside world | Clear selection (SelectionNone) |
| Hold mouse button | Only register first frame |
| Click while zoomed | Use camera transform for accurate position |

## Dependencies

| Dependency | Status | Notes |
|------------|--------|-------|
| Sprint 001 | Complete | Mock sim_gen exists |
| Sprint 002 | Complete | Camera transforms available |
| ScreenToWorld | Complete | In engine/camera/transform.go |

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Off-by-one in tile coords | Add unit tests for boundary cases |
| Click detection flaky | Use Ebiten's inpututil (proven) |
| Z-fighting on highlight | Ensure Z=1 for highlight, Z=0 for tiles |
| Camera zoom breaks coords | Test with Zoom != 1.0 |

## Future Work (not this sprint)

- Hover highlight (show tile under cursor)
- Multi-select (shift+click)
- Right-click context menu
- Keyboard navigation (arrow keys)
- Selection actions (build, harvest)

## Notes

- This is pure Go work - no AILANG changes needed
- Selection state in mock will eventually come from AILANG
- Highlight color (index 4) already defined in biomeColors
- Keep selection logic simple - complexity should be in AILANG later
