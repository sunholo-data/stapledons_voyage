# Bubble Ship Dome System - Dual Coordinate Spaces

---

**⚠️ REJECTED: 2025-12-20**

**Reason:** This 3D first-person dome approach has been abandoned in favor of scene-based 2D/2.5D navigation with AI-generated images.

**Why rejected:**
- **3D asset complexity:** Too difficult to create 3D assets for interior spaces
- **Misaligned with game focus:** Game is about space visualization and conversations, not spatial navigation
- **AI asset constraints:** AI excels at generating 2D images, struggles with 3D dimensions
- **Over-engineering:** Interior experience is for conversations and visual contrast, not 3D exploration

**Replacement:** See design decision "[2025-12-20] Interior Ship Experience: Scene-Based Navigation" in [docs/vision/design-decisions.md](../../docs/vision/design-decisions.md)

**What worked:** Exterior space visualization (starmaps, SR/GR effects) - these are kept and enhanced in new approach

---

**Status**: Rejected
**Target**: v0.5.0
**Priority**: P1 - High (blocks dome observation deck features)
**Estimated**: 1 day
**Dependencies**: demo-engine-lod (space travel), demo-game-interior (player movement)

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Rendering tech, not gameplay |
| Civilization Simulation | + | +1 | Enables observing distant civilizations from observation deck |
| Philosophical Depth | N/A | 0 | Infrastructure |
| Ship & Crew Life | + | +1 | Enables crew looking out at universe from bubble ship |
| Legacy Impact | N/A | 0 | Infrastructure |
| Hard Sci-Fi Authenticity | + | +1 | Proper SR/GR effects for realistic space observation |
| **Net Score** | | **+3** | **Decision: Move forward** |

**Feature type:** Engine/Infrastructure
- Enables observation deck gameplay features
- Required for realistic space observation from inside bubble ship

**Reference:** See [game-vision.md](../../../docs/game-vision.md)

## Problem Statement

The 100m radius bubble ship needs to reconcile two incompatible rendering systems:

**Current State:**
- [demo-engine-lod](cmd/demo-engine-lod/main.go): Works well for space travel (ship velocity affects starfield with SR/GR effects)
- [demo-game-interior](cmd/demo-game-interior/main.go): Works well for player movement (WASD through 3D rooms)
- [demo-engine-dome](cmd/demo-engine-dome/main.go): **BROKEN** - player movement (WASD) affects BOTH interior AND exterior
  - Lines 696-707: Sky sphere, platform, and struts all follow camera position
  - Player walking 1m inside ship shouldn't move distant stars
  - Ship traveling at 0.8c shouldn't move the floor under your feet

**The Core Conflict:**
- **Interior space** (ship-local): Player can walk ~100m to the dome edge, see 3D parallax of rooms/structures
- **Exterior space** (galactic): Stars are virtually infinite distance away, only affected by ship velocity

Currently `demo-engine-dome` treats player movement as moving the entire universe, which breaks immersion.

**Impact:**
- Blocks observation deck features (crew looking at stars from inside bubble)
- Breaks hard sci-fi authenticity (walking shouldn't shift distant stars)
- Prevents proper SR/GR effects (need ship velocity separate from player position)

## Goals

**Primary Goal:** Separate ship-local coordinates (player movement) from galactic coordinates (ship velocity) in the bubble dome rendering system.

**Success Metrics:**
- Player can walk around inside ship without affecting distant star positions
- Ship velocity (0-0.99c) affects sky sphere appearance via SR/GR effects
- 3D parallax visible for interior structures up to ~100m dome boundary
- Sky sphere appears at "astronomical distance" (unaffected by player position)
- demo-engine-dome works as well as demo-engine-lod + demo-game-interior combined

## Solution Design

### Overview

Implement **dual coordinate systems** in the dome renderer:

1. **Ship-local space** - Camera position for player movement (WASD)
   - Origin at ship center
   - Player walks around inside ship (-50m to +50m in each axis for 100m radius)
   - Interior geometry (struts, platform, rooms) rendered in ship-local space
   - Camera position = player position in ship coordinates

2. **Galactic space** - Sky sphere orientation based on ship velocity
   - Sky sphere always centered at camera position (follows player)
   - Sky sphere texture determined by ship velocity and heading
   - SR/GR effects applied to sky sphere texture generation
   - Sky sphere radius >> dome radius (e.g., 500m vs 5m dome)

**Key insight:** The sky sphere GEOMETRY follows the player (always surrounds viewer), but the sky sphere TEXTURE is determined by ship velocity/heading (not player position).

### Architecture

**Coordinate Spaces:**

```
Ship-Local Space (player movement):
  camPos = playerPos   // (camX, camY, camZ) in ship coordinates

Interior Rendering:
  platform.position = (0, -2, 0)         // Fixed in ship space
  struts.position = (0, 0, 0)            // Fixed in ship space
  camera.position = camPos               // Player position

Galactic Space (ship velocity):
  shipVelocity = 0.0 to 0.99c            // Ship speed through space
  shipHeading = (dirX, dirY, dirZ)       // Ship travel direction

Sky Sphere Rendering:
  skySphere.geometry.center = camPos     // Follow player (always surrounds)
  skySphere.texture = generateSkyTex(    // Based on ship velocity
    shipVelocity,
    shipHeading,
    grPhi
  )
```

**Components:**
1. **Player Movement System**: WASD input → ship-local camPos (unchanged from demo-game-interior)
2. **Interior Renderer**: Renders struts, platform, rooms at fixed ship-local positions
3. **Sky Sphere Renderer**: Geometry follows player, texture determined by ship velocity
4. **LOD Star System**: 3D stars positioned in ship-local space for nearby objects (<5 ly)

### Implementation Plan

**Phase 1: Separate Variables** (~2 hours)
- [ ] Add `shipVelocity` and `shipHeading` fields to Game struct (separate from player camX/Y/Z)
- [ ] Add keyboard controls: V key cycles ship velocity (0.0, 0.2, 0.5, 0.8c)
- [ ] Add keyboard controls: H key cycles ship heading (north, south, east, west)
- [ ] Remove lines 698-707 that make sky sphere/struts follow camera position

**Phase 2: Fix Coordinate Systems** (~3 hours)
- [ ] Keep `updateStarTexture()` using ship velocity/heading (NOT camera look direction)
- [ ] Change sky sphere: geometry follows player (line 699), texture uses ship params
- [ ] Keep struts/platform at FIXED ship-local positions (remove lines 702-707)
- [ ] Update LOD stars: position in galactic coords, but render relative to camera

**Phase 3: Test & Polish** (~1 hour)
- [ ] Test: Walk around (WASD) - stars don't move
- [ ] Test: Change ship velocity (V) - sky sphere texture updates with SR effects
- [ ] Test: Change ship heading (H) - different stars visible through dome
- [ ] Add HUD showing both player position (ship-local) and ship velocity (galactic)
- [ ] Update controls documentation in file header

### Files to Modify/Create

**New files:**
- None (all changes to existing demo)

**Modified files:**
- `cmd/demo-engine-dome/main.go` - Main changes (~50 LOC modified)
  - Add shipVelocity, shipHeading fields to Game struct
  - Add V/H key handlers for ship velocity/heading controls
  - Fix Update() to keep interior geometry fixed in ship-local space
  - Fix updateStarTexture() to use ship params, not camera look direction
  - Update HUD to show both coordinate systems

## Examples

### Example 1: Player Walks Forward (WASD)

**Before (BROKEN):**
```go
// Lines 696-707 in demo-engine-dome/main.go
g.scene.SetCameraPosition(g.camX, g.camY, g.camZ)
g.skySphere.SetPosition(g.camX, g.camY, g.camZ)  // Sky moves with player!

g.platform.SetLocalPosition(g.camX, g.camY-2.0, g.camZ)  // Floor moves!
for _, strut := range g.struts {
    strut.SetLocalPosition(g.camX, g.camY, g.camZ)       // Struts move!
}
```
Result: Player walks 1m → entire universe shifts 1m (wrong!)

**After (FIXED):**
```go
// Camera position in ship-local space
g.scene.SetCameraPosition(g.camX, g.camY, g.camZ)

// Sky sphere geometry follows player (always surrounds viewer)
g.skySphere.SetPosition(g.camX, g.camY, g.camZ)

// Interior geometry FIXED in ship-local coordinates
// (platform and struts don't move - player moves relative to them)
```
Result: Player walks around ship interior, stars stay fixed (correct!)

### Example 2: Ship Accelerates to 0.8c (V key)

**Before (BROKEN):**
```go
// updateStarTexture() uses camera LOOK direction for ship velocity
lookDirX := math.Sin(g.yaw) * math.Cos(g.pitch)  // Where player is looking!
params := render.ViewParams{
    ViewDirX: lookDirX,  // Player look direction used for SR effects
    Velocity: g.velocity,
}
```
Result: Looking around changes SR effects (wrong - ship direction ≠ view direction)

**After (FIXED):**
```go
// updateStarTexture() uses SHIP heading (independent of player look)
params := render.ViewParams{
    ViewDirX: g.shipHeading.X,  // Ship travel direction
    ViewDirY: g.shipHeading.Y,
    ViewDirZ: g.shipHeading.Z,
    Velocity: g.shipVelocity,   // Ship speed through space
}
```
Result: Ship velocity determines SR effects, player can look anywhere (correct!)

## Success Criteria

- [ ] Player movement (WASD) does not affect sky sphere star positions
  - **Test:** Walk 10m forward, verify stars don't shift
- [ ] Ship velocity (V key) affects SR effects on sky sphere
  - **Test:** Press V to 0.8c, verify aberration/Doppler effects visible
- [ ] Ship heading (H key) rotates which stars are visible
  - **Test:** Cycle through N/S/E/W, verify different constellations
- [ ] Interior geometry stays fixed in ship-local space
  - **Test:** Walk around, verify platform/struts appear at consistent positions
- [ ] 3D parallax visible for nearby objects (<5 ly)
  - **Test:** Walk around, verify nearby stars show parallax shift
- [ ] HUD shows both coordinate systems clearly
  - **Test:** Verify player position (ship-local) and ship velocity (galactic) both displayed

## Testing Strategy

**Unit tests:**
- Not applicable (demo/visualization code)

**Integration tests:**
- Not applicable (demo code)

**Manual testing:**
- Run `go run ./cmd/demo-engine-dome` and verify:
  1. Walk around with WASD - platform/struts stay fixed, stars don't move
  2. Press V to cycle ship velocity - SR effects change (aberration, Doppler)
  3. Press H to cycle ship heading - different stars visible
  4. Look around with mouse - doesn't affect SR effects or star positions
  5. Press R to reset - returns to initial state correctly
  6. HUD shows clear separation between player pos and ship velocity

## Non-Goals

**Not in this feature:**
- Full navigation system (autopilot, star charts) - Deferred to navigation feature
- Crew NPCs in observation deck - Deferred to crew simulation
- Multiple decks/rooms - This demo focuses on single dome observation area
- AILANG integration - This is a Go engine demo, AILANG integration comes later

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| LOD stars still tied to camera position | Medium | Review LOD positioning code in lines 266-333, ensure galactic coords |
| SR/GR shader performance with dual coords | Low | Shaders already separate, just need param changes |
| Confusing UX with two coordinate systems | Medium | Clear HUD labels, separate key bindings (WASD vs V/H) |

## References

- [demo-engine-lod/main.go](../../cmd/demo-engine-lod/main.go) - Working space travel with SR/GR
- [demo-game-interior/main.go](../../cmd/demo-game-interior/main.go) - Working player movement
- [demo-engine-dome/main.go](../../cmd/demo-engine-dome/main.go) - Current broken implementation
- [engine/render/space_view.go](../../engine/render/space_view.go) - Sky sphere texture generation

## Future Work

### AILANG Integration API

When porting to AILANG, use **existing engine APIs** (see [engine-capabilities.md](../../design_docs/reference/engine-capabilities.md)) with this dual coordinate pattern:

**Key Concept: Dual Coordinate State**

```ailang
type DomeState = {
  -- Player position INSIDE ship (meters, ship-local)
  playerPosX: float,
  playerPosY: float,
  playerPosZ: float,

  -- Ship motion in galaxy (light-years, galactic)
  shipPosX: float,
  shipPosY: float,
  shipPosZ: float,
  shipVelocity: float,     -- 0.0 to 0.99c
  shipHeadingX: float,
  shipHeadingY: float,
  shipHeadingZ: float,

  -- Rest of game state...
}
```

**How to Use Existing Engine DrawCmds:**

```ailang
-- Sky sphere: geometry follows PLAYER, texture uses SHIP params
DrawCmd.GalaxyBg(
  state.playerPosX, state.playerPosY, state.playerPosZ,  -- Center at player
  state.shipPosX, state.shipPosY, state.shipPosZ,        -- Texture from ship pos
  state.shipVelocity                                      -- SR effects from ship velocity
)

-- Interior geometry: FIXED in ship-local coordinates (not following player)
DrawCmd.Sprite(spriteID, 0, -2, 0)      -- Floor at origin
DrawCmd.Sprite(propID, 10, 0, 15)       -- Prop at fixed ship position

-- 3D objects: positioned relative to ship, rendered relative to player
let dirX = (objGalacticX - state.shipPosX) / dist  -- Direction from SHIP
let dirY = (objGalacticY - state.shipPosY) / dist
let dirZ = (objGalacticZ - state.shipPosZ) / dist
let renderDist = 80 + (orbitDist - 0.04) / (0.90 - 0.04) * 380  -- Map to 80-460m
DrawCmd.Billboard3D(
  textureID,
  state.playerPosX + dirX * renderDist,  -- Offset from PLAYER
  state.playerPosY + dirY * renderDist,
  state.playerPosZ + dirZ * renderDist,
  size
)
```

**Responsibilities:**

| Layer | Owns | Uses |
|-------|------|------|
| AILANG | Player pos (meters), Ship pos (light-years), Orbital mechanics, DrawCmd generation | Existing DrawCmd types, existing engine effects |
| Engine | Rendering DrawCmds, SR/GR shaders, Texture loading, Camera transforms | No game state or logic |

**All engine capabilities already exist** - this pattern just separates player coordinates from ship coordinates when generating DrawCmds.

### Other Future Work

- **Multi-room interiors**: Extend ship-local space to multiple connected rooms/decks
- **Navigation UI**: Star charts, autopilot, destination selection
- **Crew simulation**: NPCs in observation deck reacting to view
- **Dynamic ship rotation**: Ship can rotate while maintaining player position inside

---

**Document created**: 2025-12-19
**Last updated**: 2025-12-19
