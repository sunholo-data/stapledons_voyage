# Bridge Interior View

> **NOTE (2025-12-10):** This doc contains some Go code sections that need revision.
> The AILANG types (BridgeState, renderBridge, etc.) are correct.
> The Go code sections (ObservationDome struct, BridgeView struct) should be
> refactored to be stateless renderers that take DrawCmds from AILANG.
> See [view-layer-ailang-migration.md](view-layer-ailang-migration.md) for the migration plan.

**Status:** Planned
**Priority:** P0 (Core Player Experience)
**Complexity:** High
**Depends On:** ~~View System (90% done)~~ View Layer Migration, Isometric Engine (done)
**Enables:** Ship Exploration, Crew Dialogue, Galaxy Map Access
**Sprint:** [sprints/bridge-interior-sprint.md](../../../sprints/bridge-interior-sprint.md)

## Problem Statement

The bridge is the player's primary interface with the game. Currently we have:
- Working exterior space view with planets and stars
- Working isometric projection engine
- No ship interior

We need the bridge to be the **command center** where players:
- View the cosmos through the observation dome (space exterior as backdrop)
- Interact with crew at their stations
- Access consoles (navigation, communications, archives)
- Feel "at home" in their ship

This is the **first isometric interior** and sets the visual standard for all ship spaces.

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Ship Is Home | +++ | Bridge is the heart of the ship |
| Hard Sci-Fi Authenticity | ++ | Realistic command center layout |
| Crew Are Everything | ++ | Crew stations establish relationships |
| Time Has Emotional Weight | + | Crew aging visible at stations |

## Architecture Overview

### Three-Layer Composition

Per ~~[01-view-system.md](./01-view-system.md)~~ [view-layer-ailang-migration.md](./view-layer-ailang-migration.md), the bridge view composes three layers:

```
┌─────────────────────────────────────────────────────────────┐
│                        UI LAYER                              │  Z: 100-199
│    Ship Time | Location | Mode Indicators | Menus            │
├─────────────────────────────────────────────────────────────┤
│                     CONTENT LAYER                            │  Z: 10-99
│              ┌─────────────────────────┐                     │
│              │    OBSERVATION DOME     │                     │
│              │   (Space View Inset)    │                     │  Dome: Z: 90
│              │    🪐 Planet/Stars      │                     │
│              └─────────────────────────┘                     │
│                                                              │
│     [Helm]     [Comms]    [Status]    [Nav]                 │  Consoles: Z: 30
│       👤         👤          👤         👤                   │  Crew: Z: 40
│    ═══════════════════════════════════════════════          │  Floor: Z: 10
│                     [Captain's Chair]                        │
│                          👤                                  │
├─────────────────────────────────────────────────────────────┤
│                   BACKGROUND LAYER                           │  Z: 0-9
│     Starfield (subtle, visible around dome edges)            │
└─────────────────────────────────────────────────────────────┘
```

### The Observation Dome

The bridge's defining feature is a large **observation dome** that shows the exterior space view:

- **Not a flat window** - A curved dome at the front of the bridge
- **Renders SpaceView content** - Uses existing planet/star rendering
- **Clipped to dome shape** - Circular or elliptical mask
- **Perspective-correct** - 3D planets appear to curve with dome

This creates the visual connection between interior and exterior.

```go
// Dome rendering approach
type ObservationDome struct {
    bounds   Rect          // Screen bounds of dome area
    mask     *ebiten.Image // Circular mask for clipping
    spaceView *SpaceView   // Reuses existing space rendering
}

func (d *ObservationDome) Draw(screen *ebiten.Image) {
    // 1. Render space view to offscreen buffer
    d.spaceView.Draw(d.buffer)

    // 2. Apply dome mask (circular clip)
    d.buffer.DrawImage(d.mask, compositeOp)

    // 3. Draw masked result to screen at dome position
    screen.DrawImage(d.buffer, d.opts)
}
```

### Bridge Layout

The bridge is a 16x12 tile isometric space:

```
       [OBSERVATION DOME - 8 tiles wide]
              ╭────────────╮
             ╱              ╲
            ╱                ╲
     ┌─────────────────────────────┐
     │  HELM      COMMS    STATUS  │  Row 2: Console stations
     │   ⬡         ⬡         ⬡     │
     │                             │
     │     ───────────────────     │  Row 5: Central walkway
     │                             │
     │                 ⬡           │  Row 7: Captain's console
     │              CAPTAIN        │
     │                             │
     │  ⬡ NAV                 ⬡    │  Row 9: Side stations
     │                    SCIENCE  │
     │                             │
     │     ⬢         ⬢         ⬢   │  Row 11: Access hatches
     │   LIFT    CORRIDOR   STORE  │
     └─────────────────────────────┘
```

**Legend:**
- ⬡ = Console station (interactable)
- ⬢ = Access point (deck transition)

## Asset Requirements

### Tile Manifest (IDs 1000-1099: Bridge Tiles)

| ID | Name | File | Description |
|----|------|------|-------------|
| 1000 | bridge_floor | `bridge/floor.png` | Metal deck plating |
| 1001 | bridge_floor_glow | `bridge/floor_glow.png` | Floor with status lights |
| 1002 | bridge_console_base | `bridge/console_base.png` | Console station floor |
| 1003 | bridge_walkway | `bridge/walkway.png` | Central walkway tiles |
| 1004 | bridge_dome_edge | `bridge/dome_edge.png` | Edge of observation dome |
| 1005 | bridge_wall | `bridge/wall.png` | Interior wall segments |
| 1006 | bridge_hatch | `bridge/hatch.png` | Deck access hatch |
| 1007 | bridge_captain_area | `bridge/captain.png` | Captain's area flooring |

### Console Assets (IDs 1100-1149: Bridge Consoles)

| ID | Name | Size | Description |
|----|------|------|-------------|
| 1100 | console_helm | 64x48 | Pilot station with controls |
| 1101 | console_comms | 64x48 | Communications array |
| 1102 | console_status | 64x48 | Ship systems display |
| 1103 | console_nav | 64x48 | Galaxy map terminal |
| 1104 | console_science | 64x48 | Sensor/research station |
| 1105 | console_captain | 64x64 | Captain's chair with displays |

### Crew Sprites (IDs 1200-1249: Bridge Crew)

| ID | Name | Size | Animations |
|----|------|------|------------|
| 1200 | crew_pilot | 32x48 | idle, working, talking |
| 1201 | crew_comms | 32x48 | idle, working, talking |
| 1202 | crew_engineer | 32x48 | idle, working, talking |
| 1203 | crew_scientist | 32x48 | idle, working, talking |
| 1204 | crew_captain | 32x48 | idle, commanding, talking |
| 1205 | player_avatar | 32x48 | idle, walk_N/S/E/W |

### Visual Style Guide

**Aesthetic:** Clean, functional, lived-in sci-fi

**Color Palette:**
- Primary: Deep space blue (#1a2634)
- Secondary: Metallic silver (#8b95a1)
- Accent: Console amber (#f5a623)
- Alert: Status red (#ff4757)
- Highlight: Interactive cyan (#00d9ff)

**Isometric Specifics:**
- Tile size: 64x32 (standard isometric)
- Entity height: 48px (3:2 ratio to tile height)
- Console height: 48-64px (elevated from floor)
- Dome perspective: Fish-eye slight distortion

**Material Textures:**
- Deck: Brushed metal with subtle wear
- Consoles: Glass displays with holographic glow
- Walls: Padded panels with status lights
- Dome: Transparent with structural ribs

## AI Art Generation Pipeline

Assets will be generated using the `asset-manager` skill with Gemini Imagen.

### Generation Prompts

**Bridge Floor Tile:**
```
Isometric 64x32 pixel art tile, sci-fi spaceship bridge floor,
brushed metal deck plating with subtle blue ambient glow,
worn footpath marks, minimalist design, game asset,
transparent background, no shadows, clean edges
```

**Console Station:**
```
Isometric 64x48 pixel art, sci-fi spaceship console station,
holographic amber display screens, curved design,
control panels with buttons and sliders, subtle glow effects,
dark metal base, game asset, transparent background
```

**Crew Member (Idle):**
```
Isometric 32x48 pixel art character sprite, sci-fi ship crew member,
dark blue uniform with rank insignia, standing at attention,
facing camera (south direction), clean anime-inspired style,
transparent background, game asset
```

**Observation Dome Interior:**
```
Isometric view looking up into circular observation dome,
space visible through transparent dome material,
structural ribs radiating from center, subtle starlight,
sci-fi aesthetic, 128x96 game asset, transparent edges
```

### Asset Generation Workflow

```bash
# 1. Generate base tiles
.claude/skills/asset-manager/scripts/generate_asset.sh \
  --type tile \
  --name bridge_floor \
  --id 1000 \
  --prompt "Isometric 64x32 pixel art tile, sci-fi spaceship bridge floor..."

# 2. Generate consoles
.claude/skills/asset-manager/scripts/generate_asset.sh \
  --type entity \
  --name console_helm \
  --id 1100 \
  --width 64 --height 48 \
  --prompt "Isometric 64x48 pixel art, sci-fi spaceship helm console..."

# 3. Generate crew sprites
.claude/skills/asset-manager/scripts/generate_asset.sh \
  --type entity \
  --name crew_pilot \
  --id 1200 \
  --width 128 --height 48 \
  --frames 4 \
  --animations "idle:0:2:2,working:2:2:4" \
  --prompt "Isometric pixel art sprite sheet, sci-fi pilot character..."
```

## AILANG Types

### Bridge State

```ailang
module sim/bridge

import sim/protocol (Coord, DrawCmd, IsoTile, IsoEntity)

type BridgeState = {
    playerPos: Coord,
    playerFacing: Direction,
    crewPositions: [CrewPosition],
    consoleStates: [ConsoleState],
    hoveredInteractable: Option[InteractableID],
    domeView: DomeViewState
}

type CrewPosition = {
    crewId: CrewID,
    station: BridgeStation,
    pos: Coord,
    activity: CrewActivity
}

type BridgeStation =
    | StationHelm
    | StationComms
    | StationStatus
    | StationNav
    | StationScience
    | StationCaptain
    | StationNone        -- Crew is moving/idle

type ConsoleState = {
    station: BridgeStation,
    pos: Coord,
    active: bool,
    hasAlert: bool
}

type DomeViewState = {
    targetPlanet: Option[PlanetID],    -- What planet is visible
    velocity: float,                    -- Ship velocity for SR effects
    viewAngle: float                    -- Camera angle
}
```

### Bridge Rendering

```ailang
-- Generate DrawCmds for bridge view
pure func renderBridge(state: BridgeState) -> [DrawCmd] {
    let floorCmds = renderBridgeFloor();
    let consoleCmds = renderConsoles(state.consoleStates);
    let domeCmds = renderObservationDome(state.domeView);
    let crewCmds = renderBridgeCrew(state.crewPositions);
    let playerCmd = renderPlayer(state.playerPos, state.playerFacing);
    let highlightCmds = renderInteractionHighlight(state.hoveredInteractable);

    -- Concatenate in render order (floor first, player last)
    concat([floorCmds, domeCmds, consoleCmds, crewCmds, [playerCmd], highlightCmds])
}

-- Bridge floor is 16x12 tiles
pure func renderBridgeFloor() -> [DrawCmd] {
    renderFloorTiles(bridgeLayout, 0, [])
}

pure func renderFloorTiles(layout: [int], idx: int, acc: [DrawCmd]) -> [DrawCmd] {
    match layout {
        [] => acc,
        tileId :: rest => {
            let x = idx % 16;
            let y = idx / 16;
            let cmd = IsoTile({ x: x, y: y }, 0, tileId, 10, 0);
            renderFloorTiles(rest, idx + 1, acc ++ [cmd])
        }
    }
}

-- Render crew at their stations
pure func renderBridgeCrew(crew: [CrewPosition]) -> [DrawCmd] {
    map(renderCrewMember, crew)
}

pure func renderCrewMember(cp: CrewPosition) -> DrawCmd {
    let spriteId = crewSpriteId(cp.crewId);
    IsoEntity(
        crewIdToString(cp.crewId),
        cp.pos,
        0.0, 0.0,  -- No sub-tile offset when at station
        0,          -- Floor height
        spriteId,
        40          -- Layer above floor, below UI
    )
}
```

### Bridge Input Handling

```ailang
pure func processBridgeInput(
    state: BridgeState,
    input: FrameInput
) -> BridgeInputResult {
    -- Check for movement
    let afterMove = processPlayerMovement(state, input);

    -- Check for hover
    let hovered = findInteractableAt(afterMove, input.tileMouseX, input.tileMouseY);
    let afterHover = { afterMove | hoveredInteractable: hovered };

    -- Check for click
    if input.clickedThisFrame then
        match hovered {
            None => BridgeResult(afterHover),
            Some(id) => resolveInteraction(afterHover, id)
        }
    else
        BridgeResult(afterHover)
}

type BridgeInputResult =
    | BridgeResult(BridgeState)              -- Stay on bridge
    | TransitionToDialogue(CrewID)           -- Talk to crew
    | TransitionToGalaxyMap                  -- Nav console clicked
    | TransitionToDeck(int)                  -- Hatch clicked
```

## Go Engine Integration

### Observation Dome Rendering

The dome requires special handling to composite the space view:

```go
// engine/render/dome.go

type DomeRenderer struct {
    spaceRenderer *SpaceRenderer
    maskShader    *ebiten.Shader
    domeBuffer    *ebiten.Image
}

func (d *DomeRenderer) Draw(screen *ebiten.Image, domeState DomeViewState, bounds Rect) {
    // 1. Render space scene to buffer
    d.spaceRenderer.SetVelocity(domeState.Velocity)
    d.spaceRenderer.SetTargetPlanet(domeState.TargetPlanet)
    d.spaceRenderer.Draw(d.domeBuffer)

    // 2. Apply fish-eye dome effect + circular mask
    opts := &ebiten.DrawImageOptions{}
    opts.GeoM.Translate(bounds.X, bounds.Y)
    screen.DrawImage(d.domeBuffer, opts)
}
```

### Bridge View Mode

```go
// engine/render/bridge_view.go

type BridgeView struct {
    isoRenderer  *IsoRenderer
    domeRenderer *DomeRenderer
    uiRenderer   *UIRenderer
}

func (v *BridgeView) Draw(screen *ebiten.Image, state BridgeState) {
    // 1. Draw starfield background (subtle, around dome edges)
    v.drawBackgroundStars(screen)

    // 2. Draw isometric bridge floor and walls
    v.isoRenderer.Draw(screen, state.FloorCmds)

    // 3. Draw observation dome with space view
    v.domeRenderer.Draw(screen, state.DomeView, domeBounds)

    // 4. Draw consoles (in front of dome)
    v.isoRenderer.Draw(screen, state.ConsoleCmds)

    // 5. Draw crew and player
    v.isoRenderer.Draw(screen, state.CrewCmds)

    // 6. Draw UI overlay
    v.uiRenderer.Draw(screen, state.UICmds)
}
```

## Implementation Plan

### Phase 1: Bridge Layout (2 days)

| Task | Description |
|------|-------------|
| 1.1 | Create `sim/bridge.ail` with BridgeState types |
| 1.2 | Define 16x12 bridge layout as tile array |
| 1.3 | Implement `renderBridgeFloor()` function |
| 1.4 | Add bridge mode to game modes enum |
| 1.5 | Create placeholder tile assets (colored diamonds) |
| 1.6 | Test: See isometric bridge floor rendering |

### Phase 2: Observation Dome (2 days)

| Task | Description |
|------|-------------|
| 2.1 | Create `engine/render/dome.go` |
| 2.2 | Implement space view rendering to offscreen buffer |
| 2.3 | Create circular mask shader for dome effect |
| 2.4 | Add DomeViewState to AILANG protocol |
| 2.5 | Wire dome rendering into bridge view |
| 2.6 | Test: See planets through dome on bridge |

### Phase 3: Console Stations (1 day)

| Task | Description |
|------|-------------|
| 3.1 | Define console positions in bridge layout |
| 3.2 | Implement `renderConsoles()` function |
| 3.3 | Add hover highlight for consoles |
| 3.4 | Create placeholder console sprites |
| 3.5 | Test: See consoles, hover highlights work |

### Phase 4: Crew Placement (1 day)

| Task | Description |
|------|-------------|
| 4.1 | Define crew station assignments |
| 4.2 | Implement `renderBridgeCrew()` function |
| 4.3 | Add crew idle animations |
| 4.4 | Create placeholder crew sprites |
| 4.5 | Test: See crew at their stations |

### Phase 5: Player Movement (1 day)

| Task | Description |
|------|-------------|
| 5.1 | Add player position to BridgeState |
| 5.2 | Implement WASD movement processing |
| 5.3 | Add collision detection with consoles |
| 5.4 | Implement player sprite rendering |
| 5.5 | Test: Walk around bridge with WASD |

### Phase 6: Asset Generation (2 days)

| Task | Description |
|------|-------------|
| 6.1 | Generate bridge floor tiles with AI |
| 6.2 | Generate console sprites with AI |
| 6.3 | Generate crew sprites with AI |
| 6.4 | Generate dome edge/frame sprites |
| 6.5 | Update sprite manifest |
| 6.6 | Replace placeholders with generated assets |

### Phase 7: Polish & Integration (1 day)

| Task | Description |
|------|-------------|
| 7.1 | Add room entry/exit labels |
| 7.2 | Add console interaction (triggers mode change) |
| 7.3 | Add transition animation to/from bridge |
| 7.4 | Performance testing (60 FPS) |
| 7.5 | Create `demo-bridge` command |

## Demo Command

```bash
# Run bridge demo
./bin/demo-bridge

# With specific planet in dome
./bin/demo-bridge --planet earth --velocity 0.3

# With crew at stations
./bin/demo-bridge --crew 5
```

## Success Criteria

### Core Functionality
- [ ] Bridge renders as 16x12 isometric tile grid
- [ ] Observation dome shows planet/star view
- [ ] Player can move with WASD
- [ ] Crew visible at their stations
- [ ] Consoles are interactable (hover highlight)

### Visual Quality
- [ ] All assets are AI-generated (not placeholders)
- [ ] Consistent visual style across all elements
- [ ] Dome feels like looking into space
- [ ] Crew animations are smooth

### Performance
- [ ] 60 FPS with full bridge + dome rendering
- [ ] No stutter during player movement
- [ ] Dome space view updates smoothly

### Integration
- [ ] Can click Nav console to open Galaxy Map
- [ ] Can click crew to start dialogue
- [ ] Can transition to other decks
- [ ] Bridge is default mode after arrival sequence

## Dependencies

**Requires:**
- View Layer Migration (view-layer-ailang-migration.md)
- Isometric Engine - Done
- Space Background rendering - Done
- Planet 3D rendering - Done

**Enables:**
- Full Ship Exploration (ship-exploration.md)
- Crew Dialogue System
- Galaxy Map access
- Game main loop

## Visual Reference

```
       ╔══════════════════════════════════════════════╗
       ║              OBSERVATION DOME                 ║
       ║         ╭─────────────────────╮              ║
       ║        ╱   🪐                  ╲             ║
       ║       ╱     Earth approaching   ╲            ║
       ║      │    ✦  ✦    ✦   ✦    ✦    │           ║
       ║       ╲                        ╱             ║
       ║        ╰─────────────────────╯               ║
       ╠══════════════════════════════════════════════╣
       ║   ▓▓▓        ▓▓▓        ▓▓▓        ▓▓▓      ║
       ║   HELM      COMMS     STATUS       NAV      ║
       ║    👤         👤         👤         👤       ║
       ║                                              ║
       ║     ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░           ║
       ║                                              ║
       ║                   ▓▓▓                        ║
       ║                 CAPTAIN                      ║
       ║                    👤                        ║
       ║                                              ║
       ║   ⬢ DECK2      ░░░░░░░░░      ⬢ STORAGE    ║
       ╚══════════════════════════════════════════════╝

       Ship Time: 47y 3mo 12d          v=0.31c  γ=1.05
```

## Next Steps After This

1. **ship-exploration.md** - Full multi-deck ship (uses bridge as template)
2. **dialogue-system.md** - Crew conversations triggered from bridge
3. **galaxy-map.md** - Navigation console opens this view
4. **arrival-sequence.md** - Ends by transitioning to bridge
