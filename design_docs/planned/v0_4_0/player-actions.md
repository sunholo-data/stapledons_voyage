# Player Actions System

**Version:** 0.4.0
**Status:** Planned
**Priority:** P1 (Medium)
**Complexity:** Medium
**AILANG Workarounds:** None expected
**Depends On:** Player Interaction (v0.2.0) - must be completed first

## Related Documents

- [Player Interaction](../v0_2_0/player-interaction.md) - Selection system this builds on
- [Engine Layer Design](../../implemented/v0_1_0/engine-layer.md) - Input capture
- [Architecture Overview](../../implemented/v0_1_0/architecture.md) - Data flow context

## Problem Statement

The player can now select tiles (Sprint 003 complete), but cannot DO anything with the selection. The game needs actions the player can take on selected tiles.

**Current State:**
- Tiles can be selected via click
- Selection is highlighted visually
- No actions available on selection
- No info panel or context menu

**What's Needed:**
- Inspect action: View tile details (biome, resources, occupants)
- Build action: Place structures on tiles
- Action feedback: Visual/audio confirmation
- Foundation for more complex actions (harvest, assign NPCs, etc.)

## Design

### Action Model

**Action Flow:**
1. Player selects a tile (existing)
2. Player triggers action (keyboard shortcut or UI button)
3. Action validated against tile state
4. World state updated
5. Visual/audio feedback provided
6. UI updated with result

### Action Types (Phase 1)

| Action | Trigger | Effect |
|--------|---------|--------|
| Inspect | `I` key or click | Show tile info in debug output |
| Build | `B` key | Place structure on empty tile |
| Clear | `X` key | Remove structure from tile |

### AILANG Implementation

#### Types (sim/actions.ail)

```ailang
module sim/actions

-- Action the player wants to perform
type PlayerAction =
    | ActionNone
    | ActionInspect
    | ActionBuild(StructureType)
    | ActionClear

-- Structure types that can be built
type StructureType =
    | StructureHouse
    | StructureFarm
    | StructureRoad

-- Result of attempting an action
type ActionResult =
    | ActionSuccess(string)       -- Success message
    | ActionFailed(string)        -- Failure reason
    | ActionNoOp                  -- No action taken
```

#### Tile Enhancement (sim/world.ail)

```ailang
-- Enhanced tile with structure support
type Tile = {
    biome: int,
    structure: Structure,
    resources: int              -- Future: resource amount
}

type Structure =
    | NoStructure
    | HasStructure(StructureType)
```

#### Action Processing (sim/actions.ail)

```ailang
-- Process action on current selection
pure func processAction(
    action: PlayerAction,
    selection: Selection,
    world: World
) -> (World, ActionResult) {
    match selection {
        NoSelection => (world, ActionNoOp),
        TileSelected(x, y) => applyAction(action, x, y, world)
    }
}

-- Apply action to specific tile
pure func applyAction(
    action: PlayerAction,
    x: int, y: int,
    world: World
) -> (World, ActionResult) {
    match action {
        ActionNone => (world, ActionNoOp),
        ActionInspect => (world, inspectTile(x, y, world)),
        ActionBuild(structType) => buildOnTile(x, y, structType, world),
        ActionClear => clearTile(x, y, world)
    }
}

-- Inspect returns info about tile (no world change)
pure func inspectTile(x: int, y: int, world: World) -> ActionResult {
    let idx = y * world.planet.width + x;
    let tile = getTile(world.planet.tiles, idx);
    let biome = biomeToString(tile.biome);
    let structure = structureToString(tile.structure);
    ActionSuccess(concat("Tile ", concat(intToString(x), concat(",", concat(intToString(y), concat(": ", concat(biome, concat(" - ", structure))))))))
}

-- Build structure on empty tile
pure func buildOnTile(
    x: int, y: int,
    structType: StructureType,
    world: World
) -> (World, ActionResult) {
    let idx = y * world.planet.width + x;
    let tile = getTile(world.planet.tiles, idx);
    match tile.structure {
        HasStructure(_) => (world, ActionFailed("Tile already has structure")),
        NoStructure => {
            let newTile = { biome: tile.biome, structure: HasStructure(structType), resources: tile.resources };
            let newTiles = setTile(world.planet.tiles, idx, newTile);
            let newPlanet = { width: world.planet.width, height: world.planet.height, tiles: newTiles };
            let newWorld = { tick: world.tick, planet: newPlanet, npcs: world.npcs, selection: world.selection };
            (newWorld, ActionSuccess("Structure built!"))
        }
    }
}
```

#### Input Extension (sim/protocol.ail)

```ailang
-- Extend FrameInput with action triggers
type FrameInput = {
    mouse: MouseState,
    keys: [KeyEvent],
    clickedThisFrame: bool,
    worldMouseX: float,
    worldMouseY: float,
    actionRequested: PlayerAction   -- NEW: from keyboard
}
```

### Go/Engine Integration

#### Mock sim_gen Updates

Add action types:

```go
// sim_gen/types.go

// PlayerAction represents an action the player wants to perform
type PlayerAction interface {
    isPlayerAction()
}

type ActionNone struct{}
func (ActionNone) isPlayerAction() {}

type ActionInspect struct{}
func (ActionInspect) isPlayerAction() {}

type ActionBuild struct {
    StructureType StructureType
}
func (ActionBuild) isPlayerAction() {}

type ActionClear struct{}
func (ActionClear) isPlayerAction() {}

// StructureType for buildings
type StructureType int
const (
    StructureHouse StructureType = iota
    StructureFarm
    StructureRoad
)

// Structure on a tile
type Structure interface {
    isStructure()
}

type NoStructure struct{}
func (NoStructure) isStructure() {}

type HasStructure struct {
    Type StructureType
}
func (HasStructure) isStructure() {}

// ActionResult from attempting an action
type ActionResult interface {
    isActionResult()
}

type ActionSuccess struct {
    Message string
}
func (ActionSuccess) isActionResult() {}

type ActionFailed struct {
    Reason string
}
func (ActionFailed) isActionResult() {}

type ActionNoOp struct{}
func (ActionNoOp) isActionResult() {}
```

Update Tile:

```go
type Tile struct {
    Biome     int
    Structure Structure
    Resources int
}
```

Update FrameInput:

```go
type FrameInput struct {
    Mouse            MouseState
    Keys             []KeyEvent
    ClickedThisFrame bool
    WorldMouseX      float64
    WorldMouseY      float64
    ActionRequested  PlayerAction  // NEW
}
```

#### Action Detection (engine/render/input.go)

```go
func CaptureInputWithCamera(cam sim_gen.Camera, screenW, screenH int) sim_gen.FrameInput {
    // ... existing code ...

    // Detect action keys
    var action sim_gen.PlayerAction = sim_gen.ActionNone{}
    if inpututil.IsKeyJustPressed(ebiten.KeyI) {
        action = sim_gen.ActionInspect{}
    } else if inpututil.IsKeyJustPressed(ebiten.KeyB) {
        action = sim_gen.ActionBuild{StructureType: sim_gen.StructureHouse}
    } else if inpututil.IsKeyJustPressed(ebiten.KeyX) {
        action = sim_gen.ActionClear{}
    }

    return sim_gen.FrameInput{
        // ... existing fields ...
        ActionRequested: action,
    }
}
```

#### Structure Rendering (engine/render/draw.go)

```go
// Add structure colors
var structureColors = []color.RGBA{
    {139, 69, 19, 255},   // House: brown
    {50, 205, 50, 255},   // Farm: lime green
    {128, 128, 128, 255}, // Road: gray
}

// DrawStructure renders a structure sprite or rect
func DrawStructure(screen *ebiten.Image, cmd sim_gen.DrawCmdRect, structure sim_gen.StructureType) {
    // Future: Use sprites, for now use colored overlay
}
```

### Visual Feedback

#### Action Results Display

| Result | Visual | Audio |
|--------|--------|-------|
| Success | Flash green, show message | Positive chime |
| Failed | Flash red, show message | Error sound |
| NoOp | No change | Silent |

#### Structure Rendering

| Structure | Appearance |
|-----------|------------|
| House | Brown square (8x8) |
| Farm | Green square with texture |
| Road | Gray line |

### Debug Output

Action results should appear in FrameOutput.Debug:

```go
type FrameOutput struct {
    Draw   []DrawCmd
    Sounds []int
    Debug  []string  // Action messages appear here
    Camera Camera
}
```

## Implementation Plan

### Phase 1: Inspect Action

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/types.go` | Add PlayerAction, ActionResult types |
| 1.2 | `sim_gen/protocol.go` | Add ActionRequested to FrameInput |
| 1.3 | `engine/render/input.go` | Detect I key press |
| 1.4 | `sim_gen/funcs.go` | Process inspect, output to Debug |
| 1.5 | Test | Press I on tile, see debug output |

### Phase 2: Build Action

| Task | File | Description |
|------|------|-------------|
| 2.1 | `sim_gen/types.go` | Add Structure types, update Tile |
| 2.2 | `sim_gen/funcs.go` | Process build action |
| 2.3 | `engine/render/input.go` | Detect B key press |
| 2.4 | `engine/render/draw.go` | Render structures |
| 2.5 | Test | Build on empty tile, verify render |

### Phase 3: Clear Action

| Task | File | Description |
|------|------|-------------|
| 3.1 | `sim_gen/funcs.go` | Process clear action |
| 3.2 | `engine/render/input.go` | Detect X key press |
| 3.3 | Test | Clear structure, verify removal |

### Phase 4: Feedback & Polish

| Task | File | Description |
|------|------|-------------|
| 4.1 | `sim_gen/funcs.go` | Add result messages |
| 4.2 | `engine/render/draw.go` | Flash effect on action |
| 4.3 | Audio | Add action sounds (if audio ready) |
| 4.4 | Test | Full action flow |

## Testing Strategy

### Manual Testing

```bash
make run-mock
# 1. Click tile to select
# 2. Press I → debug shows tile info
# 3. Press B → structure built (if empty)
# 4. Press B again → "already has structure" message
# 5. Press X → structure removed
# 6. Press I → shows "no structure"
```

### Automated Testing

```go
func TestInspectAction(t *testing.T)
func TestBuildOnEmptyTile(t *testing.T)
func TestBuildOnOccupiedTile(t *testing.T)
func TestClearStructure(t *testing.T)
func TestActionWithNoSelection(t *testing.T)
```

### Edge Cases

- [ ] Action with no selection → NoOp
- [ ] Build on occupied tile → ActionFailed
- [ ] Clear on empty tile → ActionFailed or NoOp
- [ ] Rapid key presses → Only first action processed
- [ ] Action during camera pan → Works correctly

## AILANG Constraints

| Limitation | Impact | Workaround |
|------------|--------|------------|
| List access O(n) | Tile lookup slow for large worlds | Use index formula, consider array type in v0.5.0 |
| No string interpolation | Verbose message building | Use concat chains |
| No mutable state | Each action returns new World | Functional updates (already our pattern) |

## Future Extensions

| Feature | Description |
|---------|-------------|
| Build menu | UI to select structure type |
| Resource costs | Structures require resources |
| Tech tree | Unlock structures via research |
| Undo | Reverse last action |
| Action queue | Queue multiple actions |

## Success Criteria

### Core Functionality
- [ ] Inspect shows tile info in debug
- [ ] Build places structure on empty tile
- [ ] Build fails on occupied tile with message
- [ ] Clear removes structure
- [ ] Actions only work with selection

### Visual Feedback
- [ ] Structures render differently than base tile
- [ ] Action results visible (debug or overlay)

### Performance
- [ ] Actions process in single frame
- [ ] No lag on key press

### Integration
- [ ] Works with existing selection system
- [ ] Debug output shows action results
- [ ] Ready for AILANG migration
