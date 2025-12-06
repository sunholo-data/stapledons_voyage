# Ship Exploration Mode

**Version:** 0.5.1
**Status:** Planned
**Priority:** P0 (Core Gameplay Loop)
**Complexity:** High
**AILANG Workarounds:** Recursive room traversal, entity lookup by position
**Depends On:** v0.5.0 UI Modes Framework, v0.4.0 NPC Movement

## Related Documents

- [UI Modes Architecture](../v0_5_0/ui-modes.md) - Mode framework
- [NPC Movement](../v0_4_0/npc-movement.md) - Crew entity movement
- [Dialogue System](../v0_5_3/dialogue-system.md) - Triggered from crew interaction
- [Game Vision](../../../docs/game-vision.md) - Ship as story space

## Problem Statement

The ship is the player's "home" throughout the 100-year voyage. It needs to feel lived-in, with a crew that ages, forms relationships, and creates emergent stories. Currently there's no ship interior, no player avatar, and no way to interact with the crew between journeys.

**Current State:**
- No ship interior tilemap
- No player character on ship
- No room/deck system
- No crew interaction points
- NPCs exist but only in abstract world

**What's Needed:**
- Multi-deck ship interior layout
- Player movement through ship
- Crew NPCs positioned in rooms
- Interaction hotspots (systems, consoles, crew)
- Transition triggers to other modes
- Ambient life (crew wandering, working, socializing)

---

## Design Overview

### Ship Layout Philosophy

The ship should feel like a **generational home**, not just a vehicle:

- **Functional spaces** - Bridge, engineering, life support, archives
- **Living spaces** - Quarters, common areas, gardens
- **Sacred spaces** - Memorial wall, observation deck
- **Hidden spaces** - Maintenance tunnels, storage (discoverable)

### Deck Structure

```
Deck 1 (Top)     - Bridge, Observatory, Communications
Deck 2           - Crew Quarters, Common Room, Medical
Deck 3           - Archives, Laboratory, Workshop
Deck 4           - Gardens, Recreation, Nursery
Deck 5 (Bottom)  - Engineering, Drive Core, Cargo
```

### Visual Style

- **Isometric or top-down** - Consistent with planet surfaces
- **Lived-in aesthetic** - Personal items, wear marks, modifications
- **Generational markers** - Art, memorials, graffiti from past crews
- **Functional clarity** - Each room's purpose obvious at a glance

---

## Detailed Specification

### Ship State

```ailang
module sim/ship

type Ship = {
    decks: [Deck],
    currentDeck: int,
    playerPos: Coord,
    playerFacing: Direction,
    crewPositions: [(CrewID, Coord, int)],  -- crew, position, deck
    roomStates: [RoomState],
    systemStates: [SystemState],
    ambientEvents: [AmbientEvent]
}

type Deck = {
    id: int,
    name: string,
    tiles: [Tile],
    width: int,
    height: int,
    rooms: [Room],
    connections: [DeckConnection]  -- stairs, elevators
}

type Room = {
    id: RoomID,
    name: string,
    bounds: Rect,
    roomType: RoomType,
    interactables: [Interactable],
    ambientSlots: [Coord]  -- where crew can idle
}

type RoomType =
    | Bridge
    | Quarters(CrewID)        -- assigned to specific crew
    | CommonRoom
    | Engineering
    | Archives
    | Observatory
    | Medical
    | Gardens
    | Laboratory
    | Cargo
    | Corridor
```

### Room Interactions

```ailang
type Interactable = {
    id: InteractableID,
    pos: Coord,
    interactType: InteractType,
    label: string,
    available: bool
}

type InteractType =
    | CrewMember(CrewID)           -- Talk to crew
    | ShipSystem(SystemID)          -- Check/repair system
    | Console(ConsoleType)          -- Bridge controls, archives
    | Object(ObjectType)            -- Artifacts, mementos
    | DeckAccess(int, Coord)        -- Move to another deck

type ConsoleType =
    | NavigationConsole             -- Opens Galaxy Map
    | CommunicationsConsole         -- Check messages from civs
    | ArchiveTerminal               -- Browse knowledge/artifacts
    | StatusPanel                   -- Ship systems overview
    | MemorialWall                  -- Crew who died

type ObjectType =
    | Artifact(ArtifactID)          -- Collected from civs
    | CrewMemento(CrewID)           -- Personal item
    | ShipLog(int)                  -- Historical entry
    | Photograph(PhotoID)           -- Key moments captured
```

### Player Movement

```ailang
type ShipExplorationState = {
    playerPos: Coord,
    playerFacing: Direction,
    currentDeck: int,
    moveState: MoveState,
    hoveredEntity: Maybe(InteractableID),
    selectedEntity: Maybe(InteractableID),
    menuOpen: Maybe(ContextMenu)
}

type MoveState =
    | Idle
    | Walking(Direction)
    | Transitioning(int, Coord)  -- deck change animation

-- Process ship exploration input
pure func processShipInput(state: ShipExplorationState, ship: Ship, input: FrameInput) -> ShipExplorationState {
    let afterMove = processMovement(state, ship, input);
    let afterHover = processHover(afterMove, ship, input);
    let afterInteract = processInteraction(afterHover, ship, input);
    afterInteract
}

-- Handle WASD/arrow movement
pure func processMovement(state: ShipExplorationState, ship: Ship, input: FrameInput) -> ShipExplorationState {
    let dir = inputToDirection(input);
    match dir {
        None => { state | moveState: Idle },
        Some(d) => {
            let newPos = moveInDirection(state.playerPos, d);
            if isWalkable(newPos, getCurrentDeck(ship, state.currentDeck)) then
                { state | playerPos: newPos, playerFacing: d, moveState: Walking(d) }
            else
                { state | playerFacing: d, moveState: Idle }
        }
    }
}

-- Check what player is hovering over
pure func processHover(state: ShipExplorationState, ship: Ship, input: FrameInput) -> ShipExplorationState {
    let mousePos = screenToTile(input.mouseX, input.mouseY);
    let deck = getCurrentDeck(ship, state.currentDeck);
    let entity = findInteractableAt(deck, mousePos);
    { state | hoveredEntity: entity }
}

-- Handle click/interact
pure func processInteraction(state: ShipExplorationState, ship: Ship, input: FrameInput) -> ShipExplorationState {
    if input.clicked then
        match state.hoveredEntity {
            None => state,
            Some(id) => { state | selectedEntity: Some(id) }
        }
    else
        state
}
```

### Interaction Resolution

When player selects an interactable, the game transitions to appropriate mode:

```ailang
-- Determine what happens when interacting
pure func resolveInteraction(world: World, interactable: Interactable) -> World {
    match interactable.interactType {
        CrewMember(crewID) => {
            -- Transition to dialogue with this crew member
            let dialogueState = initCrewDialogue(crewID, world);
            transitionTo(world, ModeDialogue(dialogueState))
        },
        ShipSystem(sysID) => {
            -- Open system status panel
            let panelState = initSystemPanel(sysID, world);
            { world | mode: ModeShipExploration({ world.mode | menuOpen: Some(SystemPanel(panelState)) }) }
        },
        Console(NavigationConsole) => {
            -- Transition to galaxy map
            let mapState = initGalaxyMap(world);
            transitionTo(world, ModeGalaxyMap(mapState))
        },
        Console(ArchiveTerminal) => {
            -- Open archive browser
            let archiveState = initArchiveBrowser(world);
            { world | mode: ModeShipExploration({ world.mode | menuOpen: Some(ArchivePanel(archiveState)) }) }
        },
        Console(MemorialWall) => {
            -- Open memorial view
            let memorialState = initMemorial(world);
            { world | mode: ModeShipExploration({ world.mode | menuOpen: Some(MemorialPanel(memorialState)) }) }
        },
        DeckAccess(targetDeck, targetPos) => {
            -- Transition to new deck
            { world | mode: ModeShipExploration({ world.mode |
                currentDeck: targetDeck,
                playerPos: targetPos,
                moveState: Transitioning(targetDeck, targetPos) }) }
        },
        _ => world
    }
}
```

### Crew Ambient Behavior

Crew NPCs should feel alive, not just standing around:

```ailang
type CrewActivity =
    | Idle(Coord)                    -- Standing/sitting
    | Walking(Coord, Coord)          -- Moving between points
    | Working(RoomID)                -- At their station
    | Socializing(CrewID)            -- Talking to another crew
    | Resting(RoomID)                -- In quarters
    | Eating                         -- Common room
    | Exercising                     -- Recreation
    | Gardening                      -- Gardens deck

-- Update crew positions and activities
pure func updateCrewAmbient(ship: Ship, crew: [Crew], tick: int) -> [(CrewID, Coord, int, CrewActivity)] {
    map(\c. updateSingleCrew(c, ship, tick), crew)
}

pure func updateSingleCrew(crew: Crew, ship: Ship, tick: int) -> (CrewID, Coord, int, CrewActivity) {
    -- Simple schedule-based behavior
    let hour = (tick / 60) % 24;  -- Assuming 60 ticks = 1 hour
    let activity = match hour {
        h if h >= 22 || h < 6 => Resting(crew.quartersRoom),
        h if h >= 6 && h < 8 => Eating,
        h if h >= 8 && h < 12 => Working(crew.workStation),
        h if h >= 12 && h < 13 => Eating,
        h if h >= 13 && h < 18 => Working(crew.workStation),
        _ => chooseSocialActivity(crew, tick)
    };
    let pos = activityToPosition(activity, ship);
    let deck = activityToDeck(activity, ship);
    (crew.id, pos, deck, activity)
}
```

---

## Visual Design

### Tileset Requirements

| Tile Category | Examples | Count Est. |
|---------------|----------|------------|
| Floor | Metal, carpet, garden soil, glass | 8-12 |
| Wall | Hull, interior, window, door | 10-15 |
| Furniture | Beds, chairs, tables, consoles | 20-30 |
| Decoration | Plants, art, personal items | 15-20 |
| Systems | Panels, pipes, vents | 10-15 |
| Transitions | Stairs, elevators, hatches | 5-8 |

### Crew Sprites

| State | Frames | Animation |
|-------|--------|-----------|
| Idle | 2 | Breathing loop |
| Walking | 4 | Walk cycle per direction |
| Working | 2 | Typing/operating |
| Talking | 2 | Gesture animation |

### Player Character

- **Distinct from crew** - Different silhouette or highlight
- **Directional sprites** - 4 or 8 directions
- **Interaction indicator** - Glow when near interactable

### Room Indicators

- **Room name labels** - Appear when entering
- **System status icons** - Green/yellow/red health
- **Crew count badges** - How many in each room

---

## Go/Engine Integration

### Ship Tilemap Rendering

```go
// engine/render/ship.go

type ShipRenderer struct {
    tilesets    map[int]*ebiten.Image  // Per-deck tilesets
    crewSprites *SpriteSheet
    playerSprite *SpriteSheet
    uiElements  *UISheet
}

func (r *ShipRenderer) RenderDeck(screen *ebiten.Image, deck Deck, state ShipExplorationState) {
    // Render floor tiles
    for y := 0; y < deck.Height; y++ {
        for x := 0; x < deck.Width; x++ {
            tile := deck.Tiles[y*deck.Width+x]
            r.drawTile(screen, tile, x, y)
        }
    }

    // Render furniture/objects (sorted by Y for depth)
    objects := collectRenderableObjects(deck)
    sort.Slice(objects, func(i, j int) bool {
        return objects[i].Y < objects[j].Y
    })
    for _, obj := range objects {
        r.drawObject(screen, obj)
    }

    // Render crew on this deck
    for _, crew := range state.crewOnDeck {
        r.drawCrew(screen, crew)
    }

    // Render player
    r.drawPlayer(screen, state.playerPos, state.playerFacing)

    // Render interaction highlights
    if state.hoveredEntity != nil {
        r.drawHighlight(screen, state.hoveredEntity)
    }

    // Render room labels
    r.drawRoomLabels(screen, deck.Rooms)
}
```

### Input Handling

```go
// engine/input/ship.go

func CaptureShipInput() FrameInput {
    var input FrameInput

    // Movement keys
    if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
        input.MoveY = -1
    }
    if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
        input.MoveY = 1
    }
    if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
        input.MoveX = -1
    }
    if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
        input.MoveX = 1
    }

    // Interaction
    input.Clicked = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
    input.MouseX, input.MouseY = ebiten.CursorPosition()

    // Quick keys
    input.OpenMap = inpututil.IsKeyJustPressed(ebiten.KeyM)
    input.OpenLog = inpututil.IsKeyJustPressed(ebiten.KeyL)
    input.Pause = inpututil.IsKeyJustPressed(ebiten.KeyEscape)

    return input
}
```

---

## Implementation Plan

### Phase 1: Basic Ship Layout

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/ship.go` | Ship, Deck, Room types |
| 1.2 | `sim_gen/ship.go` | Single deck test layout (8x8 rooms) |
| 1.3 | `engine/render/ship.go` | Basic floor tile rendering |
| 1.4 | Test | See ship interior render |

### Phase 2: Player Movement

| Task | File | Description |
|------|------|-------------|
| 2.1 | `sim_gen/types.go` | ShipExplorationState type |
| 2.2 | `sim_gen/funcs.go` | Player movement logic |
| 2.3 | `engine/input/ship.go` | WASD input capture |
| 2.4 | `engine/render/ship.go` | Player sprite rendering |
| 2.5 | Test | Move around ship |

### Phase 3: Room System

| Task | File | Description |
|------|------|-------------|
| 3.1 | `sim_gen/ship.go` | Room bounds and types |
| 3.2 | `sim_gen/ship.go` | Collision detection |
| 3.3 | `engine/render/ship.go` | Room name labels |
| 3.4 | Test | Can't walk through walls |

### Phase 4: Interactables

| Task | File | Description |
|------|------|-------------|
| 4.1 | `sim_gen/ship.go` | Interactable type |
| 4.2 | `sim_gen/funcs.go` | Hover detection |
| 4.3 | `sim_gen/funcs.go` | Click handling |
| 4.4 | `engine/render/ship.go` | Highlight rendering |
| 4.5 | Test | Hover/click on objects |

### Phase 5: Crew Placement

| Task | File | Description |
|------|------|-------------|
| 5.1 | `sim_gen/funcs.go` | Crew position tracking |
| 5.2 | `sim_gen/funcs.go` | Basic schedule system |
| 5.3 | `engine/render/ship.go` | Crew sprite rendering |
| 5.4 | Test | See crew in rooms |

### Phase 6: Multi-Deck

| Task | File | Description |
|------|------|-------------|
| 6.1 | `sim_gen/ship.go` | Multiple deck definitions |
| 6.2 | `sim_gen/funcs.go` | Deck transition logic |
| 6.3 | `engine/render/ship.go` | Deck change animation |
| 6.4 | Test | Move between decks |

### Phase 7: Mode Transitions

| Task | File | Description |
|------|------|-------------|
| 7.1 | `sim_gen/funcs.go` | Navigation console → Galaxy Map |
| 7.2 | `sim_gen/funcs.go` | Crew click → Dialogue |
| 7.3 | Test | Full mode transitions |

---

## Testing Strategy

### Manual Testing

```bash
make run-mock
# 1. Game starts in ship exploration mode
# 2. WASD moves player around deck
# 3. See room boundaries respected
# 4. Hover highlights interactables
# 5. Click crew → dialogue opens
# 6. Click navigation → galaxy map opens
# 7. Use stairs → deck changes
```

### Automated Testing

```go
func TestPlayerMovement(t *testing.T)
func TestWallCollision(t *testing.T)
func TestInteractableHover(t *testing.T)
func TestDeckTransition(t *testing.T)
func TestCrewSchedule(t *testing.T)
func TestModeTransitions(t *testing.T)
```

### Headless Scenarios

```go
// Scenario: Player walks to engineering
scenario := Scenario{
    Name: "walk_to_engineering",
    Steps: []Step{
        {Input: KeyDown(KeyS), Frames: 30},  // Walk south
        {Input: KeyDown(KeyD), Frames: 20},  // Walk east
        {Assert: PlayerInRoom("Engineering")},
    },
}
```

---

## AILANG Constraints

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No mutable state | Can't update player position in place | Functional state update |
| Recursion depth | Large ship (100+ tiles per deck) | Chunk-based processing |
| List O(n) | Finding crew by position | Maintain sorted position list |
| No RNG | Crew behavior predictable | Deterministic schedule + tick seed |

---

## Performance Considerations

### Tile Count Limits

| Deck Size | Tiles | Rendering Cost |
|-----------|-------|----------------|
| 16x16 | 256 | Low |
| 32x32 | 1024 | Medium |
| 64x64 | 4096 | High - need culling |

**Recommendation:** 32x32 per deck, 5 decks = 5120 total tiles. Use viewport culling.

### Crew Updates

- Max 20-30 crew members
- Update positions once per tick (not per frame)
- Only animate visible crew

---

## Success Criteria

### Core Functionality
- [ ] Ship interior renders
- [ ] Player moves with WASD
- [ ] Walls block movement
- [ ] Can hover/select interactables
- [ ] Crew visible in rooms

### Navigation
- [ ] Can move between decks
- [ ] Can access galaxy map from bridge
- [ ] Can initiate crew dialogue

### Polish
- [ ] Room labels appear on entry
- [ ] Crew animate when idle
- [ ] Player faces movement direction
- [ ] Interaction highlights clear

### Performance
- [ ] 60 FPS with full deck visible
- [ ] No lag on deck transitions
- [ ] Smooth player movement

---

## Future Extensions

| Feature | Description |
|---------|-------------|
| Ship customization | Choose room layouts at game start |
| Room upgrades | Install better systems |
| Damage system | Rooms can be damaged/repaired |
| Crew pathfinding | Crew navigate around obstacles |
| Day/night cycle | Lighting changes with ship time |
| Personal quarters | Crew decorate their spaces |
| Ghost crew | See memories of deceased crew in their spaces |
