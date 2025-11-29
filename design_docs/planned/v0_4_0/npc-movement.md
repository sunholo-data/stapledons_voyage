# NPC Movement System

**Version:** 0.4.0
**Status:** Planned
**Priority:** P1 (Medium)
**Complexity:** Medium
**AILANG Workarounds:** Recursion depth limits, no RNG until v0.5.1
**Depends On:** None (can run in parallel with Player Actions)

## Related Documents

- [Architecture Overview](../../implemented/v0_1_0/architecture.md) - Data flow context
- [Engine Layer Design](../../implemented/v0_1_0/engine-layer.md) - Rendering
- [Player Interaction](../v0_2_0/player-interaction.md) - May interact with NPCs later

## Problem Statement

The game world is static. NPCs exist in the World state but don't move or behave. The world needs life through moving entities.

**Current State:**
- World.NPCs is an empty list `[]`
- NPC type exists but unused
- No movement or AI logic
- No NPC rendering

**What's Needed:**
- NPCs spawn in the world
- NPCs move around the grid
- Basic movement patterns (random walk, patrol)
- Visual representation of NPCs
- Foundation for complex AI (pathfinding, goals)

## Design

### Movement Model

**Tick-based Movement:**
1. Each frame, each NPC decides whether to move
2. Movement is grid-based (8x8 tiles)
3. NPCs can move in 4 directions (N, S, E, W)
4. Movement blocked by world bounds (and later obstacles)
5. Smooth visual interpolation handled by engine (future)

### Movement Patterns (Phase 1)

| Pattern | Behavior |
|---------|----------|
| Static | NPC stays in place |
| RandomWalk | Move random direction every N ticks |
| Patrol | Move along fixed path, loop |

### AILANG Implementation

#### Types (sim/npc.ail)

```ailang
module sim/npc

-- Direction for movement
type Direction = North | South | East | West

-- Movement behavior pattern
type MovementPattern =
    | PatternStatic
    | PatternRandomWalk(int)     -- move every N ticks
    | PatternPatrol([Direction]) -- follow path

-- NPC entity
type NPC = {
    id: int,
    x: int,                      -- tile X position
    y: int,                      -- tile Y position
    sprite: int,                 -- sprite index for rendering
    pattern: MovementPattern,
    patrolIndex: int,            -- current index in patrol path
    moveCounter: int             -- ticks until next move
}

-- Movement result
type MoveResult =
    | Moved(int, int)            -- new x, y
    | Blocked                    -- couldn't move
```

#### Core Movement Functions (sim/npc.ail)

```ailang
-- Get direction offset
pure func directionOffset(dir: Direction) -> (int, int) {
    match dir {
        North => (0, -1),
        South => (0, 1),
        East => (1, 0),
        West => (-1, 0)
    }
}

-- Check if position is valid (in bounds)
pure func isValidPosition(x: int, y: int, width: int, height: int) -> bool {
    x >= 0 && x < width && y >= 0 && y < height
}

-- Attempt to move NPC in direction
pure func tryMove(npc: NPC, dir: Direction, world: World) -> MoveResult {
    let offset = directionOffset(dir);
    let newX = npc.x + fst(offset);
    let newY = npc.y + snd(offset);
    if isValidPosition(newX, newY, world.planet.width, world.planet.height) then
        Moved(newX, newY)
    else
        Blocked
}

-- Update single NPC for one tick
pure func updateNPC(npc: NPC, world: World, tick: int) -> NPC {
    match npc.pattern {
        PatternStatic => npc,
        PatternRandomWalk(interval) => updateRandomWalk(npc, world, tick, interval),
        PatternPatrol(path) => updatePatrol(npc, world, path)
    }
}

-- Random walk: move every N ticks in pseudo-random direction
pure func updateRandomWalk(npc: NPC, world: World, tick: int, interval: int) -> NPC {
    if npc.moveCounter <= 0 then {
        -- Time to move! Pick direction based on tick + id (deterministic "random")
        let dirIndex = (tick + npc.id) % 4;
        let dir = indexToDirection(dirIndex);
        let result = tryMove(npc, dir, world);
        match result {
            Moved(newX, newY) => {
                id: npc.id,
                x: newX,
                y: newY,
                sprite: npc.sprite,
                pattern: npc.pattern,
                patrolIndex: 0,
                moveCounter: interval
            },
            Blocked => {
                id: npc.id,
                x: npc.x,
                y: npc.y,
                sprite: npc.sprite,
                pattern: npc.pattern,
                patrolIndex: 0,
                moveCounter: interval
            }
        }
    } else {
        { id: npc.id, x: npc.x, y: npc.y, sprite: npc.sprite, pattern: npc.pattern, patrolIndex: npc.patrolIndex, moveCounter: npc.moveCounter - 1 }
    }
}

-- Convert index to direction
pure func indexToDirection(idx: int) -> Direction {
    match idx {
        0 => North,
        1 => South,
        2 => East,
        _ => West
    }
}
```

#### NPC List Processing (sim/npc.ail)

```ailang
-- Update all NPCs (recursive, with depth consideration)
pure func updateAllNPCs(npcs: [NPC], world: World, tick: int) -> [NPC] {
    match npcs {
        [] => [],
        npc :: rest => updateNPC(npc, world, tick) :: updateAllNPCs(rest, world, tick)
    }
}

-- Generate draw commands for all NPCs
pure func npcsToDraw(npcs: [NPC]) -> [DrawCmd] {
    match npcs {
        [] => [],
        npc :: rest => npcToDraw(npc) :: npcsToDraw(rest)
    }
}

-- Single NPC draw command
pure func npcToDraw(npc: NPC) -> DrawCmd {
    -- Draw NPC as colored rect (sprite index determines color)
    -- Position in world coords (tile * 8)
    Rect(intToFloat(npc.x * 8), intToFloat(npc.y * 8), 8.0, 8.0, npc.sprite + 10, 2)
    -- Note: Z=2 to render on top of tiles (Z=0) and selection (Z=1)
}
```

#### Step Integration (sim/step.ail)

```ailang
-- Update step to process NPCs
export func step(world: World, input: FrameInput) -> (World, FrameOutput) {
    let newTick = world.tick + 1;

    -- Process selection (existing)
    let selection = processSelection(input, world);

    -- Update all NPCs
    let updatedNPCs = updateAllNPCs(world.npcs, world, newTick);

    -- Generate draw commands
    let tileCmds = tilesToDraw(world.planet.tiles, world.planet.width, 0);
    let selectionCmds = selectionToDraw(selection);
    let npcCmds = npcsToDraw(updatedNPCs);
    let drawCmds = concat(concat(tileCmds, selectionCmds), npcCmds);

    let newWorld = {
        tick: newTick,
        planet: world.planet,
        npcs: updatedNPCs,
        selection: selection
    };
    (newWorld, { draw: drawCmds, sounds: [], debug: [], camera: computeCamera(world) })
}
```

#### World Initialization (sim/world.ail)

```ailang
-- Create initial NPCs for testing
pure func createTestNPCs() -> [NPC] {
    [
        { id: 1, x: 10, y: 10, sprite: 0, pattern: PatternRandomWalk(30), patrolIndex: 0, moveCounter: 30 },
        { id: 2, x: 20, y: 15, sprite: 1, pattern: PatternRandomWalk(45), patrolIndex: 0, moveCounter: 45 },
        { id: 3, x: 30, y: 20, sprite: 2, pattern: PatternStatic, patrolIndex: 0, moveCounter: 0 }
    ]
}

export func init_world(seed: int) -> World {
    -- ... existing tile generation ...
    {
        tick: 0,
        planet: planet,
        npcs: createTestNPCs(),
        selection: NoSelection
    }
}
```

### Go/Engine Integration

#### Mock sim_gen Updates

Update NPC type:

```go
// sim_gen/types.go

// Direction for NPC movement
type Direction int
const (
    North Direction = iota
    South
    East
    West
)

// MovementPattern defines how NPC moves
type MovementPattern interface {
    isMovementPattern()
}

type PatternStatic struct{}
func (PatternStatic) isMovementPattern() {}

type PatternRandomWalk struct {
    Interval int  // ticks between moves
}
func (PatternRandomWalk) isMovementPattern() {}

type PatternPatrol struct {
    Path []Direction
}
func (PatternPatrol) isMovementPattern() {}

// NPC entity
type NPC struct {
    ID          int
    X           int
    Y           int
    Sprite      int
    Pattern     MovementPattern
    PatrolIndex int
    MoveCounter int
}
```

#### Mock NPC Logic (sim_gen/funcs.go)

```go
// updateNPC processes a single NPC for one tick
func updateNPC(npc NPC, world World, tick int) NPC {
    switch p := npc.Pattern.(type) {
    case PatternStatic:
        return npc
    case PatternRandomWalk:
        return updateRandomWalk(npc, world, tick, p.Interval)
    case PatternPatrol:
        return updatePatrol(npc, world, p.Path)
    default:
        return npc
    }
}

func updateRandomWalk(npc NPC, world World, tick int, interval int) NPC {
    if npc.MoveCounter <= 0 {
        // Time to move
        dirIndex := (tick + npc.ID) % 4
        dx, dy := directionOffset(Direction(dirIndex))
        newX, newY := npc.X+dx, npc.Y+dy

        if isValidPosition(newX, newY, world.Planet.Width, world.Planet.Height) {
            npc.X, npc.Y = newX, newY
        }
        npc.MoveCounter = interval
    } else {
        npc.MoveCounter--
    }
    return npc
}

func directionOffset(dir Direction) (int, int) {
    switch dir {
    case North: return 0, -1
    case South: return 0, 1
    case East:  return 1, 0
    case West:  return -1, 0
    default:    return 0, 0
    }
}

func isValidPosition(x, y, width, height int) bool {
    return x >= 0 && x < width && y >= 0 && y < height
}
```

#### NPC Rendering (engine/render/draw.go)

```go
// Add NPC colors (sprite index + 10 to avoid biome collision)
var npcColors = []color.RGBA{
    {255, 0, 0, 255},     // NPC 0: Red
    {0, 255, 0, 255},     // NPC 1: Green
    {0, 0, 255, 255},     // NPC 2: Blue
    {255, 255, 0, 255},   // NPC 3: Yellow
    {255, 0, 255, 255},   // NPC 4: Magenta
}

// getColor now handles NPC colors
func getColor(colorIndex int) color.RGBA {
    if colorIndex >= 10 {
        npcIndex := colorIndex - 10
        if npcIndex < len(npcColors) {
            return npcColors[npcIndex]
        }
    }
    if colorIndex < len(biomeColors) {
        return biomeColors[colorIndex]
    }
    return color.RGBA{255, 255, 255, 255}
}
```

### Visual Design

#### NPC Appearance

| Sprite Index | Color | Represents |
|--------------|-------|------------|
| 0 (color 10) | Red | Worker |
| 1 (color 11) | Green | Scout |
| 2 (color 12) | Blue | Builder |

#### Rendering Layers

| Z | Content |
|---|---------|
| 0 | Terrain tiles |
| 1 | Selection highlight |
| 2 | NPCs |
| 3 | UI elements (future) |

## Implementation Plan

### Phase 1: NPC Types & Static Rendering

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/types.go` | Add Direction, MovementPattern, update NPC |
| 1.2 | `sim_gen/funcs.go` | Add test NPCs in InitWorld |
| 1.3 | `sim_gen/funcs.go` | Generate NPC DrawCmds in Step |
| 1.4 | `engine/render/draw.go` | Handle NPC colors (index >= 10) |
| 1.5 | Test | NPCs render at fixed positions |

### Phase 2: Random Walk Movement

| Task | File | Description |
|------|------|-------------|
| 2.1 | `sim_gen/funcs.go` | Implement updateNPC |
| 2.2 | `sim_gen/funcs.go` | Implement updateRandomWalk |
| 2.3 | `sim_gen/funcs.go` | Process all NPCs in Step |
| 2.4 | Test | NPCs move around randomly |

### Phase 3: Boundary Checking

| Task | File | Description |
|------|------|-------------|
| 3.1 | `sim_gen/funcs.go` | Add isValidPosition check |
| 3.2 | `sim_gen/funcs.go` | Prevent out-of-bounds movement |
| 3.3 | Test | NPCs stop at world edges |

### Phase 4: Polish & Patterns

| Task | File | Description |
|------|------|-------------|
| 4.1 | `sim_gen/funcs.go` | Add PatternPatrol support |
| 4.2 | Test | Patrol NPC follows path |
| 4.3 | Visual | Ensure smooth appearance |

## Testing Strategy

### Manual Testing

```bash
make run-mock
# 1. Launch game → see 3 colored squares (NPCs)
# 2. Wait → 2 NPCs start moving randomly
# 3. Watch NPC approach edge → stops at boundary
# 4. One NPC should stay static (test static pattern)
```

### Automated Testing

```go
func TestNPCInitialization(t *testing.T)
func TestRandomWalkMovement(t *testing.T)
func TestBoundaryCollision(t *testing.T)
func TestStaticNPCDoesntMove(t *testing.T)
func TestMultipleNPCsIndependent(t *testing.T)
```

### Edge Cases

- [ ] NPC at (0,0) → can move South/East only
- [ ] NPC at corner → limited movement
- [ ] All NPCs at same tile → should render all
- [ ] MoveCounter rollover → resets correctly
- [ ] Empty NPC list → no crashes

## AILANG Constraints

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No RNG | Can't do true random | Use tick+id for deterministic "random" |
| Recursion depth | Can't recurse through 100+ NPCs | Keep NPC count < 50, or batch process |
| List O(n) | Finding specific NPC is slow | Use list indices, consider ID map |
| No mutable state | Must rebuild NPC structs | Functional update pattern |

### RNG Note

When AILANG v0.5.1 adds RNG effect:
- `rand_int(4)` for random direction
- `AILANG_SEED` for reproducible tests
- Remove tick+id hack

## Performance Considerations

### NPC Count Limits

| Count | Recursion Depth | Expected Performance |
|-------|-----------------|---------------------|
| 10 | 10 | Excellent |
| 50 | 50 | Good |
| 100 | 100 | May approach limits |
| 500+ | 500+ | Likely problematic |

**Recommendation:** Start with 3-10 NPCs, profile before scaling.

### Optimization Opportunities

1. **Spatial partitioning** (future) - Only update visible NPCs
2. **Movement throttling** - Not all NPCs move every frame
3. **Batch updates** - Process NPCs in groups

## Future Extensions

| Feature | Description |
|---------|-------------|
| Pathfinding | A* to target destination |
| Collision | NPCs can't overlap |
| Goals | NPCs seek resources, buildings |
| Interaction | Click NPC to select |
| Animation | Sprite animation during movement |
| Smooth movement | Visual interpolation between tiles |

## Success Criteria

### Core Functionality
- [ ] NPCs appear in world
- [ ] Random walk NPCs move periodically
- [ ] Static NPCs stay in place
- [ ] NPCs respect world boundaries

### Visual Quality
- [ ] NPCs clearly visible on tiles
- [ ] NPCs render on top of terrain
- [ ] Movement appears regular (not jittery)

### Performance
- [ ] No lag with 10 NPCs
- [ ] Frame rate stable

### Integration
- [ ] NPCs part of World state
- [ ] NPC state persists across frames
- [ ] Ready for AILANG migration
