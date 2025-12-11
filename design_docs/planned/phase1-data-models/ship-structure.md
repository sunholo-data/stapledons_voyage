# Higgs Bubble Ship Structure

**Status**: Planned
**Target**: v0.4.0
**Priority**: P0 - Foundation for all ship gameplay
**Estimated**: 3-4 weeks (art-heavy)
**Dependencies**: View System, Isometric Engine

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | + | +1 | Open views to space constantly remind player of journey |
| Civilization Simulation | 0 | 0 | Infrastructure - doesn't directly affect civs |
| Philosophical Depth | + | +1 | Cathedral spaces, Archive shrine create contemplative spaces |
| Ship & Crew Life | +++ | +3 | THE core implementation of Pillar 4 - this IS the ship as home |
| Legacy Impact | + | +1 | The bubble civilization's fate shown in end-screen |
| Hard Sci-Fi Authenticity | ++ | +2 | 1g thrust gravity, Higgs bubble physics, realistic volume |
| **Net Score** | | **+8** | **Decision: Move forward - critical foundation** |

**Feature type:** Infrastructure + Gameplay (enables all ship-based features)

**Design Decisions Referenced:**
- [2025-12-08] Higgs Bubble Ship: 10-20+ Levels Around Spire
- [2025-12-08] Visual Aesthetic: French 70s Comic (Moebius/Métal Hurlant)
- [2025-12-08] Spire: Monolithic Superstructure with Archive Interface
- [2025-12-08] Ship Orientation: Vertical Thrust Axis with 1g Gravity
- [2025-12-08] Archive: Distributed Terminals Plus Robots
- [2025-12-08] Open Levels: Views Outward Through Bubble
- [2025-12-06] Bubble Society as Living Sim
- [2025-12-06] The Spire as Universal Constant

## Problem Statement

The game needs a fully realized ship interior that serves as the player's home for 100 subjective years. Currently we have:
- Working isometric projection engine
- Bridge interior design doc (single room)
- No overall ship structure defined

**Current State:**
- Bridge doc describes one 16x12 tile room
- No multi-level navigation
- No connection between ship areas
- Visual style was "clean functional sci-fi" - now superseded

**Impact:**
- Pillar 4 (Ship Is Home) cannot be realized without this
- All crew interaction, Archive dialogue, and ship life depends on this
- Sets visual identity for the entire game

## Goals

**Primary Goal:** Define the complete Higgs bubble ship as a 10-20+ level structure around a central spire, with distinct zones, navigation systems, and views outward to space.

**Success Metrics:**
- Player can navigate between all major zones
- Views outward show space with SR/GR effects from any open level
- Spire is visually prominent and mysterious from all locations
- Visual style is consistently French 70s comic aesthetic
- Ship feels like a small city, not a submarine

## Solution Design

### Overview

The ship is a vertical structure oriented along the thrust axis:

```
                    ▲ Direction of travel (0.9c+)
                    │
            ╭───────┴───────╮
           ╱   OBSERVATION   ╲      Level 20: Observation Deck
          ╱      DOME         ╲
         │ ═══════════════════ │    Level 18-19: Command
         │        BRIDGE       │
         │ ═══════════════════ │    Level 15-17: Upper Commons
         │       COMMONS       │
         │ ═══════════════════ │    Level 12-14: Residential Upper
         │     DWELLINGS       │
         │ ═══════════════════ │    Level 9-11: Gardens/Cathedral
         │   GARDEN CATHEDRAL  │
         │ ═══════════════════ │    Level 6-8: Residential Lower
         │      DWELLINGS      │
         │ ═══════════════════ │    Level 3-5: Industrial
         │     FABRICATION     │
         │ ═══════════════════ │    Level 1-2: Engineering
         │       ENGINES       │
          ╲                   ╱
           ╲     SPIRE       ╱      Central column through all levels
            ╰───────┬───────╯
                    │
                    ▼ Thrust (provides 1g gravity)

    ←─── BUBBLE BOUNDARY ───→   (100m radius, transparent)
```

### Ship Specifications

| Specification | Value | Notes |
|---------------|-------|-------|
| **Bubble radius** | 100 meters | Per bubble-ship-design |
| **Population** | ~100 people (start) | Can grow/shrink |
| **Levels** | 15-20 | Each ~5-8m vertical spacing |
| **Level diameter** | 30-80m | Varies by height (wider at equator) |
| **Spire diameter** | ~10m | Central column, inaccessible interior |
| **Gravity** | 1g | From constant thrust |

### Level Zones

#### Command Zone (Levels 18-20)

| Level | Name | Purpose |
|-------|------|---------|
| 20 | Observation Dome | Panoramic view, ceremonial space |
| 19 | Bridge | Ship control, navigation, crew stations |
| 18 | Command Support | Archive primary shrine, captain's quarters |

**Visual Character:** Most exposed to space, dramatic views, the "cathedral" of command. Observation dome is the highest point - you look "up" into space through the transparent bubble.

#### Upper Commons (Levels 15-17)

| Level | Name | Purpose |
|-------|------|---------|
| 17 | Market Plaza | Trade, social gathering, news |
| 16 | Assembly Hall | Town meetings, votes, celebrations |
| 15 | Recreation | Games, exercise, entertainment |

**Visual Character:** Open, bustling, where society happens. Multiple interconnected platforms with chaotic ramps and lifts.

#### Residential Upper (Levels 12-14)

| Level | Name | Purpose |
|-------|------|---------|
| 14 | Upper Dwellings | Family homes, private spaces |
| 13 | Mid Dwellings | Crew quarters, communal living |
| 12 | Commons Edge | Transition zone, small shops |

**Visual Character:** Small houses, winding paths, balconies with views outward. Intimate scale against cosmic backdrop.

#### Garden Cathedral (Levels 9-11)

| Level | Name | Purpose |
|-------|------|---------|
| 11 | Upper Gardens | Food production, orchards |
| 10 | Cathedral | Remembrance, Earth memorial, rituals |
| 9 | Lower Gardens | Hydroponics, water features |

**Visual Character:** The emotional heart of the ship. Organic growth, natural light (simulated), waterfalls. The cathedral is where crew remember Earth and process cosmic grief. "Sad but happy" - bittersweet.

#### Residential Lower (Levels 6-8)

| Level | Name | Purpose |
|-------|------|---------|
| 8 | Worker Housing | Engineering crew quarters |
| 7 | Industrial Edge | Transition, material storage |
| 6 | Archive Annex | Data storage, research labs |

**Visual Character:** More industrial feel, closer to the engines. Still has views outward but denser, more utilitarian.

#### Industrial Zone (Levels 3-5)

| Level | Name | Purpose |
|-------|------|---------|
| 5 | Fabrication | Proto-tech construction, workshops |
| 4 | Processing | Material recycling, refinement |
| 3 | Storage | Raw materials, emergency supplies |

**Visual Character:** Functional, mechanical. The organic-mechanical blend is most visible here - pipes that look grown, machines with organic curves.

#### Engineering (Levels 1-2)

| Level | Name | Purpose |
|-------|------|---------|
| 2 | Power Systems | Reactor interfaces, power distribution |
| 1 | Engine Core | Thrust control, spire base interface |

**Visual Character:** The deepest level, closest to the engines. Restricted access. The spire's base is here - glowing, humming, forbidden.

### The Spire

The Higgs Generator Spire runs through the center of all levels:

| Aspect | Description |
|--------|-------------|
| **Visual** | Monolithic, dark, subtly glowing. Material looks neither metal nor organic - alien. |
| **Sound** | Low constant hum, felt as much as heard |
| **Access** | Crew cannot enter. Surface is smooth, no doors. |
| **Interaction** | Archive terminals adjacent to spire on each level |
| **Mystery** | May be the same object across all universes in the recursion loop |

**Spire Visual by Zone:**
- **Command:** Spire surface shows faint patterns, Archive interprets as "navigation data"
- **Gardens:** Spire appears to have organic veins, may be alive
- **Industrial:** Spire surface has geometric patterns, interpreted as "engineering specs"
- **Engineering:** Spire base glows brightest, Archive warns of "proximity effects"

### Navigation Systems

Movement between levels uses multiple systems (arranged "chaotically" per Moebius aesthetic):

| System | Description | Speed |
|--------|-------------|-------|
| **Central Lifts** | Large platforms along spire, main transit | Fast |
| **Spiral Ramps** | Wrapped around levels, scenic routes | Slow |
| **Local Lifts** | Small platforms between adjacent levels | Medium |
| **Stairs** | Emergency/exercise routes | Slow |
| **Ladders** | Maintenance access | Slow |

**Navigation Philosophy:** No single "correct" route. Multiple paths between any two points. Getting lost is possible and interesting.

### Views Outward

From any open level, looking sideways (perpendicular to spire):

```
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│   INTERIOR              BUBBLE EDGE           SPACE          │
│   (isometric)           (parallax)            (SR/GR)        │
│                                                              │
│   ╔═══════════╗         ░░░░░░░░░░░         ✦    ✦  ✦       │
│   ║  LEVEL    ║         ░ GLOW    ░         ✦ Stars         │
│   ║  FLOOR    ║  ──▶    ░ (ISM    ░   ──▶   Planets 🪐      │
│   ║ SPIRE ║   ║         ░ impact) ░         SR aberration   │
│   ╚═══════════╝         ░░░░░░░░░░░         Doppler shift   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Layered rendering:**
1. **Interior:** Isometric level geometry, crew, objects
2. **Bubble edge:** Faint glow from ISM impacts (brighter forward)
3. **Space:** Stars/planets with SR effects based on velocity

### Visual Aesthetic: French 70s Comic

Reference artists:
- **Moebius (Jean Giraud):** Clean flowing lines, impossible architecture
- **Philippe Druillet:** Cathedral-scale structures, cosmic horror beauty
- **Métal Hurlant magazine:** Saturated colors, vast emptiness

**Key Visual Principles:**

| Principle | Application |
|-----------|-------------|
| **Organic-mechanical blend** | Pipes that look grown, machines with organic curves |
| **Saturated colors vs emptiness** | Rich interior colors against black space |
| **Tiny humans vs massive structures** | Figures dwarfed by architecture |
| **Clean flowing curves** | Avoid sharp industrial angles |
| **Cathedral-like awe** | Vertical spaces that evoke wonder |

**Color Palette (evolved from bridge doc):**

| Use | Color | Hex | Notes |
|-----|-------|-----|-------|
| Primary structure | Deep teal | #1a3d4d | Organic-mechanical base |
| Secondary | Warm amber | #d4a574 | Living spaces, warmth |
| Accent | Coral pink | #e07b6c | Organic highlights |
| Spire | Alien violet | #4a3b6e | Unknowable, other |
| Space | Deep black | #0a0a12 | Vast emptiness |
| Stars | Shifted spectrum | varies | SR Doppler effects |

### Archive Presence

The Archive AI is accessible throughout the ship:

| Presence Type | Location | Visual |
|---------------|----------|--------|
| **Primary Terminal** | Level 18 (Archive Shrine) | Large, ornate, ceremonial |
| **Level Terminals** | Adjacent to spire each level | Medium, integrated into spire |
| **Mobile Robots** | Roaming all levels | Small, spider-like, helpful |
| **Audio** | Everywhere | Voice from any terminal |

**Archive Visual Design:**
- Terminals have organic-tech aesthetic matching spire
- Robots are small but distinctly "Archive" - same material as spire
- Screen interfaces show text/imagery that sometimes "glitches" (hints at cross-universe data)

## Architecture

### AILANG Types

```ailang
module sim/ship

type ShipLevel = {
    id: int,
    name: string,
    zone: ShipZone,
    tiles: [LevelTile],
    entities: [LevelEntity],
    connections: [LevelConnection]
}

type ShipZone =
    | ZoneCommand
    | ZoneUpperCommons
    | ZoneResidentialUpper
    | ZoneGardenCathedral
    | ZoneResidentialLower
    | ZoneIndustrial
    | ZoneEngineering

type LevelTile = {
    pos: Coord,
    tileType: TileType,
    walkable: bool,
    interactable: Option[InteractableID]
}

type TileType =
    | TileFloor
    | TileSpireEdge
    | TileOpenEdge      -- View to space
    | TileWall
    | TileRamp
    | TileGarden
    | TileWater

type LevelConnection = {
    fromPos: Coord,
    toLevel: int,
    toPos: Coord,
    connectionType: ConnectionType
}

type ConnectionType =
    | ConnLift
    | ConnRamp
    | ConnStairs
    | ConnLadder

type ArchiveTerminal = {
    pos: Coord,
    level: int,
    terminalType: TerminalType
}

type TerminalType =
    | TerminalPrimary    -- Level 18 shrine
    | TerminalLevel      -- Standard level terminal
    | TerminalRobot      -- Mobile robot position
```

### Engine Integration

```go
// engine/ship/level_renderer.go

type LevelRenderer struct {
    isoRenderer   *IsoRenderer
    spaceRenderer *SpaceRenderer  // For outward views
    spireRenderer *SpireRenderer  // Central spire
}

func (r *LevelRenderer) Draw(screen *ebiten.Image, level ShipLevel, camera Camera) {
    // 1. Draw space background (visible through open edges)
    r.drawSpaceBackground(screen, camera)

    // 2. Draw spire (always visible, centered)
    r.spireRenderer.Draw(screen, level.Zone, camera)

    // 3. Draw level floor and structures
    r.isoRenderer.DrawTiles(screen, level.Tiles, camera)

    // 4. Draw entities (crew, objects, player)
    r.isoRenderer.DrawEntities(screen, level.Entities, camera)

    // 5. Draw bubble edge glow (overlay)
    r.drawBubbleEdge(screen, camera)
}
```

### Implementation Plan

**Phase 1: Core Structure** (~1 week)
- [ ] Define all ShipLevel types in AILANG
- [ ] Create level data for 3 test levels (Bridge, Gardens, Engineering)
- [ ] Implement level rendering with spire
- [ ] Implement basic navigation between levels

**Phase 2: Visual Style** (~1.5 weeks)
- [ ] Create Moebius-style asset prompts
- [ ] Generate floor tiles for each zone
- [ ] Generate spire segments
- [ ] Generate structural elements (ramps, lifts, walls)
- [ ] Implement space view from open edges

**Phase 3: Full Ship** (~1 week)
- [ ] Create all 15-20 levels
- [ ] Implement navigation mesh (lifts, ramps, stairs)
- [ ] Add Archive terminals per level
- [ ] Add level transition animations

**Phase 4: Polish** (~0.5 weeks)
- [ ] Ambient audio per zone
- [ ] Spire visual effects (glow, hum)
- [ ] Bubble edge glow rendering
- [ ] Performance optimization

### Files to Create

**AILANG:**
- `sim/ship.ail` - Ship types and level data (~400 LOC)
- `sim/ship_render.ail` - DrawCmd generation for levels (~300 LOC)
- `sim/navigation.ail` - Pathfinding between levels (~200 LOC)

**Go Engine:**
- `engine/ship/level_renderer.go` - Level rendering (~300 LOC)
- `engine/ship/spire_renderer.go` - Spire rendering (~200 LOC)
- `engine/ship/navigation.go` - Level transitions (~150 LOC)

**Assets:**
- `assets/sprites/ship/` - All ship tiles and entities
- `assets/audio/ship/` - Ambient audio per zone

## Success Criteria

- [ ] Player can navigate between all 15-20 levels
- [ ] Spire is visible and prominent from every level
- [ ] Open edges show space with SR/GR effects
- [ ] Visual style is consistently Moebius-inspired
- [ ] Each zone has distinct character
- [ ] Archive terminals accessible on every level
- [ ] Navigation feels "chaotic but navigable"
- [ ] Ship feels like a small city, not corridors
- [ ] 60 FPS on all levels

## Testing Strategy

**Visual tests:**
- Screenshot each level, verify Moebius aesthetic
- Verify spire visibility from all angles
- Verify space view SR effects from open edges

**Navigation tests:**
- Pathfind from any level to any other
- Verify all connections work bidirectionally
- Test lift/ramp animations

**Performance tests:**
- Measure FPS on most complex levels
- Verify no stutter during level transitions

## Non-Goals

**Not in this feature:**
- Crew AI/pathfinding (separate feature)
- Dialogue system (separate feature)
- Archive conversation content (separate feature)
- Detailed room interiors (furniture, etc.) - later
- Dynamic level changes over time - later

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Art generation doesn't match Moebius style | High | Create detailed style guide, iterate on prompts |
| 15-20 levels too much content | Med | Start with 5 key levels, expand incrementally |
| Navigation confusing | Med | Add map/minimap, waypoint system |
| Performance with large levels | Med | Level-of-detail, culling, streaming |

## References

- [docs/vision/design-decisions.md](../../../docs/vision/design-decisions.md) - All ship decisions logged 2025-12-08
- [02-bridge-interior.md](02-bridge-interior.md) - Bridge as one level within this structure
- [bubble-ship-design.md](../future/bubble-ship-design.md) - Original bubble physics
- Moebius artwork: Arzach, The Incal, Airtight Garage
- Druillet artwork: Lone Sloane, Salammbô

## Future Work

- Crew housing assignments and customization
- Dynamic population growth affecting level usage
- Faction territories within ship
- Ship damage/repair systems
- Level modifications over 100-year journey
- Generational changes to architecture

---

**Document created**: 2025-12-08
**Last updated**: 2025-12-08
