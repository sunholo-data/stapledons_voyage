# Bridge Rendering Improvements

## Status
- **Status**: Planned
- **Priority**: P2 (Nice to have for visual polish)
- **Estimated**: 2-3 days

## Problem Summary

The bridge view has two rendering issues identified during isometric demo development:

### Issue 1: Tile Tessellation (Square vs Isometric)

**Current State:**
- Bridge floor sprites are **1024x1024 squares** (RGB, no alpha)
- Engine expects **2:1 aspect ratio isometric diamonds** (64x32 with transparent corners)
- Stretching square sprites to fit isometric diamonds causes visual gaps/overlaps

**Why This Matters:**
- Tiles don't tessellate properly when viewed
- The isometric aesthetic is broken

### Issue 2: Animation State Not Passed

**Current State:**
- `IsoEntity` DrawCmd uses `OffsetX`/`OffsetY` to determine animation
- AILANG renders entities with `0.0, 0.0` offsets (tile-based positioning)
- Engine always plays "idle" animation, never "walk"

**Why This Matters:**
- Characters appear frozen even when moving
- Animation system exists and works, but protocol doesn't communicate movement state

## Game Vision Alignment

| Pillar | Alignment | Notes |
|--------|-----------|-------|
| Time Dilation Consequence | N/A | Infrastructure feature |
| Civilization Simulation | N/A | Visual polish only |
| Ship & Crew Life | Neutral | Improves bridge visuals |
| Hard Sci-Fi Authenticity | N/A | No physics implications |
| Legacy Impact | N/A | Infrastructure |
| Philosophical Depth | N/A | Infrastructure |

**Conclusion:** This is an infrastructure/visual polish feature that doesn't conflict with any pillars.

## Proposed Solutions

### Solution A: Hybrid Rendering (Recommended)

Use the existing **pre-rendered floor cache** (already in `bridge_view.go`) for the floor, and only use IsoEntity for dynamic elements (player, crew, consoles).

**Pros:**
- Already partially implemented (`prerenderFloorCache`)
- Best performance (floor rendered once)
- Works with current square sprites

**Cons:**
- Floor is not tile-based (can't interact with individual tiles)
- Mixing 2D floor with isometric entities

**Implementation:**
1. Skip `IsoTile` rendering for bridge floor in AILANG
2. Use Go `prerenderFloorCache` for floor visuals
3. Only emit `IsoEntity` for dynamic objects
4. Animation issue still needs fixing separately

### Solution B: Add Animation Hint to Protocol

Extend `IsoEntity` to include animation state from AILANG.

**Current IsoEntity:**
```ailang
IsoEntity(id: string, tile: Coord, offsetX: float, offsetY: float, height: int, spriteId: int, layer: int)
```

**Proposed IsoEntity:**
```ailang
IsoEntity(id: string, tile: Coord, offsetX: float, offsetY: float, height: int, spriteId: int, layer: int, animName: string)
```

**Pros:**
- Clean separation of concerns
- AILANG controls animation state
- Works for any movement style

**Cons:**
- Protocol change requires updating all call sites
- More data passed per entity

**Implementation:**
1. Update `sim/protocol.ail` to add `animName` field
2. Update `sim_gen` types
3. Update `engine/render/draw_iso.go` to use `animName`
4. Update AILANG rendering to pass "idle"/"walk" based on `MoveState`

### Solution C: Movement Detection in Engine

Have the engine track entity positions frame-to-frame and detect movement.

**Pros:**
- No protocol changes
- Works automatically

**Cons:**
- Engine has game logic (violates AILANG-first principle)
- Can't distinguish movement types (walk, run, etc.)

**Not recommended** - violates architecture principle that all game logic is in AILANG.

## Recommended Approach

1. **For tessellation:** Use Solution A (hybrid rendering)
   - Bridge already has `prerenderFloorCache`
   - Modify AILANG to not emit IsoTile for bridge (or make them invisible)
   - Use floor cache for visual floor

2. **For animations:** Use Solution B (protocol extension)
   - Add `animName: string` to IsoEntity
   - AILANG determines animation based on MoveState
   - Clean separation of concerns

## AILANG Changes

### Bridge Rendering (Solution A)

```ailang
-- Option 1: Skip floor tiles entirely, rely on Go floor cache
export pure func renderBridgeFloor(state: BridgeState) -> [DrawCmd] {
    []  -- Empty, floor rendered by Go prerenderFloorCache
}

-- Option 2: Emit transparent/invisible floor markers for collision only
-- (Not rendered, but used for tile lookups)
```

### Animation State (Solution B)

```ailang
-- Updated IsoEntity constructor in protocol.ail
type DrawCmd =
    | ...
    | IsoEntity(string, Coord, float, float, int, int, int, string)  -- Added animName

-- In bridge.ail, pass animation based on state
export pure func renderPlayer(state: BridgeState) -> [DrawCmd] {
    let anim = match state.moveState {
        MoveIdle => "idle",
        MoveWalking(_) => "walk",
        MoveTransitioning(_, _) => "walk"
    };
    [IsoEntity("player", state.playerPos, 0.0, 0.0, 0, spritePlayer(), layerPlayer(), anim)]
}
```

## Engine Changes

### draw_iso.go

```go
func (r *Renderer) drawIsoEntity(screen *ebiten.Image, c *sim_gen.DrawCmdIsoEntity, cam sim_gen.Camera, screenW, screenH int) {
    // ...

    // Use animation name from AILANG (new field)
    animName := c.AnimName
    if animName == "" {
        animName = "idle"  // Fallback for compatibility
    }

    if r.anims != nil && r.anims.HasAnimations(int(c.SpriteId)) {
        r.drawAnimatedEntity(screen, sprite, c, sx, sy, cam, screenW, screenH, animName)
        return
    }
    // ...
}
```

## Success Criteria

- [ ] Bridge floor renders without tessellation gaps
- [ ] Player/crew play "walk" animation when moving
- [ ] Player/crew play "idle" animation when stationary
- [ ] No performance regression
- [ ] AILANG-first architecture preserved

## Testing

1. Visual inspection of bridge floor for gaps
2. Move player with WASD, verify walk animation plays
3. Stop moving, verify idle animation plays
4. Verify crew members animate when patrolling

## Dependencies

- No blocking dependencies
- Animation system already implemented
- Floor cache already implemented
