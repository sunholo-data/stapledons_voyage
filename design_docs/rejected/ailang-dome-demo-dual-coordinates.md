# AILANG Dome Demo - Dual Coordinate System

---

**⚠️ REJECTED: 2025-12-20**

**Reason:** 3D dome dual-coordinate system abandoned for scene-based approach. Marked as "most important scene in the game" but implemented differently.

**Why rejected:**
- Player movement in ship-local meters is unnecessary (no spatial navigation)
- Dual coordinates still needed but for different purpose: deck scene rendering vs galactic positioning
- Scene-based approach achieves same goals (observation deck, crew gathering) without 3D complexity
- Quote preserved: "marrying internal player experience with accurate external physics" - but through layered 2D scenes + live starmap compositing

**Replacement:** See design decision "[2025-12-20] Interior Ship Experience: Scene-Based Navigation" in [docs/vision/design-decisions.md](../../docs/vision/design-decisions.md)

**What to preserve:** The "most important scene" concept - observation deck where crew witnesses cosmos. Now implemented as outward-facing deck scenes with windowed starmap views.

---

**Status**: Rejected
**Target**: v0.5.0
**Priority**: P0 - Critical (most important scene in the game)
**Estimated**: 3 days
**Dependencies**: bubble-ship-dome-system.md (architecture reference)

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | ++ | +2 | Core scene for experiencing time dilation visually |
| Civilization Simulation | ++ | +2 | Enables observing galaxy evolution from inside ship |
| Philosophical Depth | + | +1 | Contemplative space for crew to witness cosmic changes |
| Ship & Crew Life | ++ | +2 | Primary crew gathering space, observation deck |
| Legacy Impact | + | +1 | Where crew reflects on their journey's meaning |
| Hard Sci-Fi Authenticity | ++ | +2 | Proper SR/GR effects, dual coordinate separation |
| **Net Score** | | **+10** | **CRITICAL FEATURE** |

**User quote:** "This is the most important scene of the game - the marrying together of internal player movement with accurate physics of external objects outside the dome"

**Reference:** See [game-vision.md](../../docs/game-vision.md)

## Problem Statement

### What We Have

- `demo-engine-dome` (Go) - Working dome with player movement and sky sphere
- `sim/solar_demo.ail` - Working solar system simulation in AILANG
- `sim/dome_demo.ail` - **BROKEN** - Current AILANG implementation that mixes coordinate systems

### What's Wrong

The current AILANG implementation creates 4 separate Window3D commands every frame and doesn't properly separate:

1. **Ship-local coordinates (meters)** - Player walking around inside the 50m radius dome
2. **Galactic coordinates (AU/light-years)** - Planets, stars, celestial objects affected by ship velocity

**Current symptoms:**
- Window3D recreated every frame (logs spam "Loaded 3802 stars")
- No dome struts visible
- No proper separation between player movement and ship motion
- Floor not opaque
- Stars/planets not visible through dome

### What We Need

An AILANG implementation that matches demo-engine-dome's dual coordinate architecture:

- **Inside dome (ship-local)**: Dome structure, floor, props - all in meters, adjust when player moves
- **Outside dome (galactic)**: Solar system, stars - in AU, only adjust with ship velocity

**Key insight from bubble-ship-dome-system.md:** The sky sphere GEOMETRY follows the player (always surrounds viewer), but the sky sphere TEXTURE is determined by ship velocity/heading (not player position).

## Goals

**Primary Goal:** Implement dual coordinate system in AILANG for dome observation deck.

**Success Metrics:**
- ✅ Player can walk around inside ship (WASD) without affecting distant star positions
- ✅ Ship velocity affects sky appearance via SR/GR effects
- ✅ 3D parallax visible for interior structures (dome struts, floor props)
- ✅ Sky sphere appears at astronomical distance
- ✅ No logs spam (window created once, not every frame)
- ✅ Dome struts visible going up above player's head

## Solution Design

### Architecture Overview

**Coordinate Separation:**

```ailang
-- Player State (ship-local, meters)
type PlayerState = {
    posX: float,        -- Position inside ship (-50 to +50m)
    posY: float,        -- Height (1.5m = eye level)
    posZ: float,        -- Position inside ship
    yaw: float,         -- Player look direction (radians)
    pitch: float,       -- Vertical look angle
    -- ... movement params
}

-- Ship State (galactic, AU / light-years)
type ShipState = {
    posX: float,        -- Ship position in galaxy (light-years)
    posY: float,
    posZ: float,
    velocity: float,    -- Ship speed through space (0-0.99c)
    headingX: float,    -- Ship travel direction (unit vector)
    headingY: float,
    headingZ: float
}

-- Demo State combining both
type DomeState = {
    tick: int,
    player: PlayerState,      -- Ship-local coordinates
    ship: ShipState,          -- Galactic coordinates
    dome: DomeConfig,         -- Static dome params (radius, etc.)
    solarBodies: [SolarPlanet] -- From solar_demo.ail
}
```

### Rendering Strategy

**Based on demo-engine-dome architecture (lines 696-707):**

1. **Camera**: Set to player position (ship-local)
   ```ailang
   Camera3D(player.posX, player.posY, player.posZ, player.yaw, player.pitch, 75.0)
   ```

2. **Interior Geometry**: FIXED in ship-local coordinates (doesn't move with player)
   ```ailang
   -- Floor (Room3D) stays at origin
   Room3D(100.0, 100.0, 10.0, "", "", "", floorColor, wallColor, ceilColor, 1.0, 5)

   -- Props stay at fixed ship positions
   Prop3D("prop_center", 0.0, floorY, 0.0, ...)
   Prop3D("prop_r1_0", 10.0, floorY, 0.0, ...)
   ```

3. **Sky Sphere**: Geometry follows player, texture from ship params
   ```ailang
   -- ShipState3D sets ship parameters for sky sphere rendering
   ShipState3D(
       ship.posX, ship.posY, ship.posZ,           -- Ship position in galaxy
       ship.headingX, ship.headingY, ship.headingZ, -- Ship direction
       0.0, 1.0, 0.0,                             -- Up vector
       ship.velocity, 0.0                         -- SR velocity, GR phi
   )

   -- Window3D uses ship state to render sky with proper SR/GR
   -- Geometry: surrounds player at large radius (e.g., 500m)
   -- Texture: generated from ship velocity/heading
   Window3D("dome_sky", 0.0, 0.0, 0.0, 0.0, 1.0, 0.0,
            360.0, 180.0, 1, true, false, 0.0, 0.0, 20)
   ```

4. **Solar System Objects**: Positioned in galactic coords, rendered relative to player
   ```ailang
   -- From solar_demo.ail planets
   -- Convert galactic position to billboard position
   let dirX = (planet.posX - ship.posX) / dist
   let dirY = (planet.posY - ship.posY) / dist
   let dirZ = (planet.posZ - ship.posZ) / dist

   -- Render at scaled distance from player (map AU to visible meters)
   let renderDist = 80.0 + scaleFactor * orbitRadius
   Billboard3D(
       planet.textureName,
       player.posX + dirX * renderDist,
       player.posY + dirY * renderDist,
       player.posZ + dirZ * renderDist,
       planet.radius * sizeScale
   )
   ```

### DrawCmd Generation

**Current problem (sim/dome_demo.ail lines 208-213):**
```ailang
-- WRONG: Creates 4 windows every frame
let domeWindows = [
    Window3D("win_north", 0.0, 5.0, 50.0, ...),
    Window3D("win_south", 0.0, 5.0, -50.0, ...),
    Window3D("win_east", 50.0, 5.0, 0.0, ...),
    Window3D("win_west", -50.0, 5.0, 0.0, ...)
]
```

**Correct approach:**
```ailang
-- Single window for entire dome sky (hemispherical or full sphere)
-- Engine creates sky sphere ONCE, updates texture based on ShipState3D params
let skyWindow = Window3D(
    "dome_sky",
    0.0, 0.0, 0.0,      -- Position doesn't matter (follows player geometry)
    0.0, 1.0, 0.0,      -- Normal pointing up (full dome coverage)
    360.0, 180.0,       -- Full hemisphere or sphere
    1,                  -- Layer
    true,               -- Show stars
    false,              -- No grid
    0.0, 0.0,           -- No velocity offset (handled by ShipState3D)
    20                  -- Priority
)
```

**Dome struts:** Not currently in DrawCmd API - need to add or use Prop3D

**Options for dome struts:**
1. Add `DomeStruts3D` DrawCmd (engine generates meridian/ring geometry)
2. Use multiple `Prop3D` commands for individual struts
3. Use `Line3D` DrawCmd if available (not in current protocol.ail)

**Recommendation:** Use Prop3D for now (simple), request DomeStruts3D for future

### State Updates

**Player movement (WASD):** Only affects PlayerState
```ailang
pure func updatePlayerPosition(
    player: PlayerState,
    keys: [KeyEvent],
    domeRadius: float
) -> PlayerState {
    -- Calculate new position in ship-local space
    -- Clamp to dome boundary (maxDist = radius - 1m)
    -- Return updated player (ship state unchanged)
}
```

**Ship motion (future - V/H keys):** Only affects ShipState
```ailang
pure func updateShipVelocity(ship: ShipState, delta: float) -> ShipState
pure func updateShipHeading(ship: ShipState, newHeading: Vec3) -> ShipState
```

**Solar system animation:** Advances orbital phases (independent of both player and ship)
```ailang
pure func updatePlanetRotations(planets: [SolarPlanet]) -> [SolarPlanet]
```

## Implementation Plan

### Phase 1: Fix Coordinate Separation (Day 1)

**Tasks:**
- [ ] Add `ShipState` type to dome_demo.ail
- [ ] Update `DomeState` to include both `player: PlayerState` and `ship: ShipState`
- [ ] Initialize ship at origin with zero velocity: `{posX: 0, posY: 0, posZ: 0, velocity: 0.0, heading: (0,0,1)}`
- [ ] Add `ShipState3D` DrawCmd before Window3D in renderDomeDemo
- [ ] Replace 4 Window3D with single dome sky window
- [ ] Test: Player movement doesn't affect sky (no logs spam)

**Acceptance:**
- ✅ Logs show "Loaded 3802 stars" ONCE at startup, not every frame
- ✅ WASD movement works, sky stays fixed

### Phase 2: Fix Interior Rendering (Day 2)

**Tasks:**
- [ ] Verify Room3D has proper height (10.0m) for opaque floor/ceiling
- [ ] Verify Room3D colors are visible (not black/transparent)
- [ ] Add dome struts using Prop3D (8 meridians, 3 horizontal rings)
  - Meridians: 8 evenly spaced around circle (every 45°)
  - Rings: At heights 10m, 25m, 40m
  - Use thin vertical/horizontal cylinders
- [ ] Test: Floor opaque, struts visible above player head

**Acceptance:**
- ✅ Floor is solid gray (not transparent)
- ✅ Dome struts visible going up above player's head
- ✅ Props (colored cubes) visible at floor level

### Phase 3: Add Solar System Rendering (Day 3)

**Tasks:**
- [ ] Import solar system bodies from solar_demo.ail
- [ ] Convert planet positions (AU) to billboard positions (meters from player)
  - Direction: `(planet.pos - ship.pos) / distance`
  - Render distance: Map orbital radius to visible scale (80-460m)
- [ ] Add Billboard3D for each planet
- [ ] Test planet visibility and SR effects (when ship velocity > 0)
- [ ] Add HUD showing player pos (ship-local) and ship velocity

**Acceptance:**
- ✅ Planets visible through dome at scaled distances
- ✅ Planet positions don't change when player walks (only with ship motion)
- ✅ HUD clearly shows dual coordinate systems

### Phase 4: Test & Polish

**Tasks:**
- [ ] Screenshot verification at frames 30, 60, 90
- [ ] Verify no logs spam (stars loaded once)
- [ ] Verify dome structure (struts, floor, props)
- [ ] Verify solar system visible through dome
- [ ] Update design doc with actual implementation details
- [ ] Move design doc to implemented/v0_5_0/

**Acceptance:**
- ✅ All visual elements working
- ✅ No performance issues
- ✅ Clear separation between coordinate systems

## Files to Modify/Create

**Modified files:**
- `sim/dome_demo.ail` - Major rewrite
  - Add ShipState type
  - Fix coordinate separation in renderDomeDemo
  - Use single Window3D instead of 4
  - Add dome struts via Prop3D
  - Add solar system billboards
- `cmd/demo-game-dome/main.go` - No changes needed (already has effect handlers)

**No new files needed** - this is a rewrite of existing dome demo.

## Testing Strategy

### Visual Verification (Screenshots)

```bash
# Use sprint-executor screenshot helper
.claude/skills/sprint-executor/scripts/take_screenshot.sh \
  -c demo-game-dome -f 30 -o out/screenshots/dome-v2/initial.png

.claude/skills/sprint-executor/scripts/take_screenshot.sh \
  -c demo-game-dome -f 60 -o out/screenshots/dome-v2/mid.png

.claude/skills/sprint-executor/scripts/take_screenshot.sh \
  -c demo-game-dome -f 90 -o out/screenshots/dome-v2/final.png
```

**What to verify in screenshots:**
1. Floor is opaque gray (Room3D visible)
2. Dome struts visible going up above player head
3. Props (colored cubes) at floor level
4. Stars/planets visible through dome
5. No black screen or rendering artifacts

### Manual Testing

```bash
make sim          # Compile AILANG
make game         # Build executable
go run ./cmd/demo-game-dome
```

**Test cases:**
1. **Player movement** - WASD moves player, stars stay fixed
2. **Mouse look** - Camera rotates, doesn't affect sky rendering
3. **Shift run** - Movement speed increases
4. **Logs** - "Loaded 3802 stars" appears ONCE, not every frame
5. **Visual elements**:
   - Floor opaque
   - Dome struts visible
   - Props visible
   - Stars visible through dome
   - Planets visible (if ship velocity > 0 in future)

## Success Criteria

- [ ] Player can walk around inside 50m radius dome
- [ ] Dome struts visible going up above player's head
- [ ] Floor opaque with props for depth perception
- [ ] Stars/planets visible through dome
- [ ] Player movement (WASD) doesn't affect sky
- [ ] No logs spam (sky loaded once, not every frame)
- [ ] Screenshots show all visual elements correctly
- [ ] Code properly separates ship-local vs galactic coordinates

## AILANG Constraints

### Available Features (v0.5.0)
- ✅ Record types (PlayerState, ShipState, DomeState)
- ✅ Lists of records ([SolarPlanet])
- ✅ Pattern matching
- ✅ std/math (sin, cos, sqrt, tan)
- ✅ Recursion (for list operations)

### Potential Issues
- **Dome struts**: May need many Prop3D commands (8 meridians × 3 rings = 24 struts)
  - Workaround: Generate list of Prop3D DrawCmds
- **Billboard positioning**: Requires vector math (normalize, scale)
  - Workaround: Inline math operations
- **No loops**: Must use recursion for generating struts
  - Workaround: Helper functions with pattern matching

### AILANG Feedback Plan

Report any issues encountered:
```bash
ailang messages send user "Description of issue" \
  --title "Dome Demo: <specific issue>" \
  --from stapledons_voyage \
  --type bug \
  --github
```

## References

### Key Reference Implementations
- [demo-engine-dome/main.go](../../cmd/demo-engine-dome/main.go) - Working Go dome implementation
  - Lines 138-146: Ship vs player coordinate separation
  - Lines 654-718: Dome strut generation
  - Lines 696-707: Coordinate system updates
  - Lines 1094-1101: Rendering with dual coordinates
- [bubble-ship-dome-system.md](./bubble-ship-dome-system.md) - Dual coordinate architecture
- [sim/solar_demo.ail](../../sim/solar_demo.ail) - Solar system data
- [engine-capabilities.md](../reference/engine-capabilities.md) - Available DrawCmds

### DrawCmd Reference
- `Camera3D(x, y, z, yaw, pitch, fov)` - First-person camera
- `ShipState3D(posX, posY, posZ, fwdX, fwdY, fwdZ, upX, upY, upZ, velocity, grPhi)` - Ship parameters for sky rendering
- `Room3D(width, depth, height, floorTex, wallTex, ceilTex, floorColor, wallColor, ceilColor, uvScale, priority)` - Opaque floor/walls
- `Prop3D(id, x, y, z, sizeX, sizeY, sizeZ, texture, color, priority)` - 3D objects (struts, props)
- `Window3D(id, x, y, z, normX, normY, normZ, width, height, layer, showStars, showGrid, velX, velY, priority)` - Sky sphere window
- `Billboard3D(texture, x, y, z, size)` - Always-facing sprites (planets)

## Non-Goals

**Not in this demo:**
- Multiple decks/rooms - Single dome observation area only
- Crew NPCs - Focus on visual/physics demonstration
- Ship navigation controls (V/H keys) - Deferred to future iteration
- Autopilot/star charts - Game features, not demo features

## Future Work

### Ship Motion Controls (Next Iteration)
Add keyboard controls to modify ship state:
- V key: Cycle ship velocity (0.0, 0.2c, 0.5c, 0.8c, 0.95c)
- H key: Cycle ship heading (North, South, East, West, custom)
- Updates ship state, triggers SR/GR effects on sky sphere

### Multi-Room Ship
Extend to multiple connected rooms/decks:
- Engineering deck (machinery)
- Bridge deck (pilot controls)
- Habitat deck (crew quarters)
- Dome deck (observation, already implemented)
- Deck transitions (stairs, elevators)

### Game Integration
Integrate into main game loop:
- Save/load ship and player state
- Crew NPCs in observation deck
- Time dilation UI showing years passing outside
- Navigation system for selecting destinations

---

**Document created**: 2025-12-19
**Last updated**: 2025-12-19
**References**: bubble-ship-dome-system.md, demo-engine-dome/main.go
