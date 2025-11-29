# Player Interaction System

**Version:** 0.2.0
**Status:** Planned
**Priority:** P1 (Medium)
**Complexity:** Medium
**AILANG Workarounds:** None expected

## Related Documents

- [Engine Layer Design](../../implemented/v0_1_0/engine-layer.md) - Input capture
- [Camera & Viewport](../v0_3_0/camera-viewport.md) - Screen-to-world coordinate conversion
- [Architecture Overview](../../implemented/v0_1_0/architecture.md) - Data flow context

## Problem Statement

The game currently has no player interaction. Users cannot click, select, or interact with the world in any way.

**Current State:**
- Mouse position captured but not used
- Keyboard events captured but not processed
- No visual feedback for hover or selection
- No player entity or cursor

**What's Needed:**
- Click on tiles to select them
- Visual highlight for selected tile
- Hover feedback (optional)
- Foundation for future player actions (move, build, etc.)

## Design

### Interaction Model

**Click-to-Select:**
1. User clicks on screen
2. Screen coordinates → world coordinates (requires camera)
3. World coordinates → tile coordinates
4. Selected tile stored in World state
5. Render highlights selected tile

### AILANG Implementation

#### Types (sim/protocol.ail)

```ailang
-- Selection state
type Selection =
    | NoSelection
    | TileSelected(int, int)    -- x, y tile coordinates

-- Add to FrameInput (already has mouse)
type FrameInput = {
    mouse: MouseState,
    keys: [KeyEvent]
}
```

#### World State (sim/world.ail)

```ailang
-- Add selection to World
type World = {
    tick: int,
    planet: PlanetState,
    npcs: [NPC],
    selection: Selection        -- NEW: current selection
}
```

#### Input Processing (sim/input.ail - new file)

```ailang
module sim/input

type Selection = NoSelection | TileSelected(int, int)
type Coord = { x: int, y: int }

-- Convert screen position to tile coordinates
-- Note: Requires tile size and camera offset (passed from engine via FrameInput)
pure func screenToTile(screenX: float, screenY: float, tileSize: int) -> Coord {
    {
        x: floatToInt(screenX) / tileSize,
        y: floatToInt(screenY) / tileSize
    }
}

-- Check if left mouse button is pressed
pure func isLeftClick(buttons: [int]) -> bool {
    match buttons {
        [] => false,
        b :: rest => if b == 0 then true else isLeftClick(rest)
    }
}

-- Process click to update selection
pure func processClick(input: FrameInput, tileSize: int, worldW: int, worldH: int) -> Selection {
    if isLeftClick(input.mouse.buttons) then {
        let tile = screenToTile(input.mouse.x, input.mouse.y, tileSize);
        if tile.x >= 0 && tile.x < worldW && tile.y >= 0 && tile.y < worldH then
            TileSelected(tile.x, tile.y)
        else
            NoSelection
    } else
        NoSelection
}
```

#### Step Integration (sim/step.ail)

```ailang
-- Update step to process selection
export func step(world: World, input: FrameInput) -> (World, FrameOutput) {
    let newTick = world.tick + 1;

    -- Process selection (only on click, not hold)
    let newSelection = processClick(input, 8, world.planet.width, world.planet.height);
    let selection = match newSelection {
        NoSelection => world.selection,  -- Keep previous if no new click
        _ => newSelection                 -- Update on click
    };

    -- Generate draw commands including selection highlight
    let tileCmds = tilesToDraw(world.planet.tiles, world.planet.width, 0);
    let selectionCmds = selectionToDraw(selection);
    let drawCmds = append(tileCmds, selectionCmds);

    let newWorld = {
        tick: newTick,
        planet: world.planet,
        npcs: world.npcs,
        selection: selection
    };
    let output = { draw: drawCmds, sounds: [], debug: [] };
    (newWorld, output)
}

-- Draw selection highlight
pure func selectionToDraw(sel: Selection) -> [DrawCmd] {
    match sel {
        NoSelection => [],
        TileSelected(x, y) => [
            -- Draw highlight rect on top of tile (z=1)
            Rect(intToFloat(x * 8), intToFloat(y * 8), 8.0, 8.0, 4, 1)
        ]
    }
}
```

### Go/Engine Integration

#### Mock sim_gen Updates

Add Selection type to mock:

```go
// sim_gen/types.go
type Selection interface {
    isSelection()
}

type SelectionNoSelection struct{}
func (SelectionNoSelection) isSelection() {}

type SelectionTileSelected struct {
    X, Y int
}
func (SelectionTileSelected) isSelection() {}
```

Update World:

```go
type World struct {
    Tick      int
    Planet    PlanetState
    NPCs      []NPC
    Selection Selection  // NEW
}
```

#### Highlight Color

Add selection highlight color to render:

```go
// engine/render/draw.go
var biomeColors = []color.RGBA{
    {0, 100, 200, 255},   // 0: Water
    {34, 139, 34, 255},   // 1: Forest
    {210, 180, 140, 255}, // 2: Desert
    {139, 90, 43, 255},   // 3: Mountain
    {255, 255, 0, 128},   // 4: Selection highlight (yellow, semi-transparent)
}
```

### Click Detection (Engine-side alternative)

If AILANG click detection is complex, engine can detect clicks and pass flag:

```go
// engine/render/input.go
type FrameInput struct {
    Mouse MouseState
    Keys  []KeyEvent
    // Could add:
    // ClickedThisFrame bool
}

func CaptureInput() sim_gen.FrameInput {
    // Detect just-pressed (not held)
    clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
    // ...
}
```

## Visual Feedback

### Selection Highlight

| State | Visual |
|-------|--------|
| No selection | Normal tile rendering |
| Tile selected | Yellow semi-transparent overlay |
| Hover (future) | Subtle border or tint |

### Highlight Rendering Options

**Option A: Separate DrawCmd (recommended)**
- AILANG emits highlight as Rect with special color
- Engine renders with alpha blending

**Option B: Engine-side rendering**
- Engine checks World.Selection
- Draws highlight directly
- Breaks "thin engine" principle

**Decision:** Option A - AILANG controls what to highlight

## Future Extensions

| Feature | Description |
|---------|-------------|
| Multi-select | Shift+click to select multiple tiles |
| Drag select | Click and drag to select region |
| Right-click menu | Context actions on selection |
| Keyboard select | Arrow keys to move selection |
| Actions | Build, harvest, inspect selected tile |

## Implementation Plan

### Phase 1: Basic Click Selection

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/types.go` | Add Selection type to mock |
| 1.2 | `sim_gen/funcs.go` | Update InitWorld, Step with selection |
| 1.3 | `engine/render/draw.go` | Add highlight color |
| 1.4 | Test | Click tiles, see highlight |

### Phase 2: Click Detection Refinement

| Task | File | Description |
|------|------|-------------|
| 2.1 | `engine/render/input.go` | Detect just-pressed vs held |
| 2.2 | `sim_gen/funcs.go` | Only update selection on new click |
| 2.3 | Test | Click once to select, click elsewhere to move |

### Phase 3: AILANG Integration (when compiler ready)

| Task | File | Description |
|------|------|-------------|
| 3.1 | `sim/world.ail` | Add Selection to World |
| 3.2 | `sim/input.ail` | Click processing functions |
| 3.3 | `sim/step.ail` | Integrate selection into step |

## Testing Strategy

### Manual Testing

```bash
make run-mock
# Click on different tiles → should highlight
# Click outside world → no highlight
# Click same tile → stays highlighted
# Click different tile → highlight moves
```

### Automated Testing

```go
func TestScreenToTile(t *testing.T)
func TestSelectionHighlight(t *testing.T)
func TestClickOutOfBounds(t *testing.T)
```

### Edge Cases

- [ ] Click at world edge → clamp or reject
- [ ] Click at (0,0) → should work
- [ ] Rapid clicking → no issues
- [ ] Click while scrolling (future) → correct tile

## AILANG Constraints

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No click events | Must detect from button state | Compare current vs previous frame (or engine detects) |
| List for buttons | O(n) to check if clicked | Small list (3 buttons max), acceptable |

## Success Criteria

### Core Functionality
- [ ] Click on tile highlights it
- [ ] Click elsewhere moves highlight
- [ ] Click outside world clears selection
- [ ] Highlight renders on correct tile

### Visual Quality
- [ ] Highlight clearly visible
- [ ] No z-fighting with tile
- [ ] Smooth (no flicker)

### Performance
- [ ] No lag on click
- [ ] Selection check is O(1)

### Integration
- [ ] Selection state in World
- [ ] Works with current mock sim_gen
- [ ] Ready for AILANG migration
