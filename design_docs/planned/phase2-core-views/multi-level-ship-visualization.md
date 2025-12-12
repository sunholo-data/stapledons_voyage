# Multi-Level Ship Visualization

**Status**: Planned
**Target**: v0.4.0
**Priority**: P1 (Foundation for Ship Exploration)
**Estimated**: 3-4 days
**Dependencies**: Isometric Depth & Parallax System, Viewport Compositing
**Enables**: Ship Exploration, Bridge Interior, Full Bubble Ship Experience

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Infrastructure feature |
| Civilization Simulation | N/A | 0 | Infrastructure feature |
| Philosophical Depth | + | +1 | Spire visibility reinforces mystery |
| Ship & Crew Life | ++ | +2 | Makes ship feel like 3D home |
| Legacy Impact | N/A | 0 | Infrastructure feature |
| Hard Sci-Fi Authenticity | + | +1 | Matches bubble ship physics |
| **Net Score** | | **+4** | **Decision: Move forward** |

**Feature type:** Engine/Infrastructure (with strong Gameplay enablement)

**Rationale:** The bubble ship is described as having nested layers (Core → Engineering → Habitat → Agri → Outer Shell) with the Higgs Spire running through all levels. For the ship to feel like a real 3D home, players need visual cues showing levels above and below, the spire as a constant anchor, and transparency effects that reinforce they're in a transparent bubble.

**Reference:** See [bubble-ship-design.md](../../input/bubble-ship-design.md) - "Multiple layers, recognizable silhouettes, rich parallax backgrounds"

## Problem Statement

The ship-exploration design requires 5+ decks, but current isometric rendering shows only one flat plane. There's no way to:
- Show levels above/below the current deck
- Display the central spire running through all levels
- Create visual "cut-away" effects revealing adjacent decks
- Help players understand vertical position in the ship

**Current State:**
- Single-deck isometric rendering works
- No concept of "deck above" or "deck below"
- No vertical reference point (spire)
- Deck transitions would feel like teleporting to unrelated spaces

**Impact:**
- Ship doesn't feel like a coherent 3D structure
- Players lose spatial awareness between decks
- The spire (a key narrative element) has no visual presence
- "Bubble ship" feeling is lost without vertical context

## Goals

**Primary Goal:** Enable visualization of multi-level ship structure with the spire as central anchor, showing adjacent deck hints and maintaining spatial coherence across deck transitions.

**Success Metrics:**
- Spire visible through translucent floors on all decks
- "Deck above" preview shows silhouettes of upper structures
- "Deck below" preview shows hint of lower level
- Smooth visual transition when changing decks
- Players can orient themselves relative to spire position
- 60 FPS with current deck + 2 adjacent deck previews

## Solution Design

### Overview

Introduce a **Multi-Level Visualization System** that:
1. Renders current deck as primary scene
2. Shows spire as always-visible background element
3. Previews adjacent decks (above/below) with reduced opacity
4. Provides smooth transitions between decks

```
┌─────────────────────────────────────────────────────────────┐
│                   DECK ABOVE (preview)                       │
│             ░░░ faint silhouettes ░░░                       │
│                        │                                     │
├────────────────────────┼────────────────────────────────────┤
│                        │                                     │
│   [Console]           │║│          [Console]                │
│                       │║│                                    │
│      Current          │║│  ← SPIRE (always visible)         │
│      Deck             │║│                                    │
│                       │║│                                    │
│                        │                                     │
├────────────────────────┼────────────────────────────────────┤
│                        │                                     │
│             ░░░ hint of deck below ░░░                      │
│                   DECK BELOW (preview)                       │
└─────────────────────────────────────────────────────────────┘
```

### Ship Vertical Structure

Based on [bubble-ship-design.md](../../input/bubble-ship-design.md):

```
                    OBSERVATION DECK (Top)
                           │
                    ┌──────┴──────┐
           Deck 1   │   BRIDGE    │  ← Command & Navigation
                    └──────┬──────┘
                           │
           Deck 2   ┌──────┴──────┐
                    │   HABITAT   │  ← Crew Quarters, Commons
                    └──────┬──────┘
                           │
           Deck 3   ┌──────┴──────┐
                    │   CULTURE   │  ← Archive, Labs, Gardens
                    └──────┬──────┘
                           │
           Deck 4   ┌──────┴──────┐
                    │ ENGINEERING │  ← Power, Life Support
                    └──────┬──────┘
                           │
                    ┌──────┴──────┐
                    │    CORE     │  ← Spire Base (restricted)
                    └─────────────┘

                    ═══════════════
                    HIGGS SPIRE runs through ALL decks
                    (central reference point)
```

### Architecture

**Key Concepts:**

1. **DeckStack**: Ordered collection of deck definitions
2. **SpireRenderer**: Central column visible through all decks
3. **AdjacentDeckPreview**: Reduced-opacity view of nearby decks
4. **DeckTransition**: Animation when moving between decks
5. **VerticalContext**: UI showing current position in ship

**Components:**

1. **DeckStackManager** (`engine/render/deck_stack.go`): Manages deck rendering order
2. **SpireRenderer** (`engine/render/spire.go`): Renders central spire at all depths
3. **DeckPreview** (`engine/render/deck_preview.go`): Ghost view of adjacent decks
4. **TransitionAnimator** (`engine/render/deck_transition.go`): Smooth deck changes

### AILANG Types

```ailang
module sim/ship_levels

import sim/protocol (Coord, DrawCmd, IsoTile)

-- Deck definition
type Deck = {
    id: int,
    name: string,
    level: int,              -- 0 = bottom (core), 4 = top (bridge)
    tiles: [TileData],
    width: int,
    height: int,
    spirePosition: Coord,    -- Where spire intersects this deck
    accessPoints: [AccessPoint]
}

-- Access point between decks (stairs, elevator, hatch)
type AccessPoint = {
    position: Coord,
    targetDeck: int,
    targetPosition: Coord,
    accessType: AccessType
}

type AccessType =
    | Stairs
    | Elevator
    | Hatch
    | Ladder

-- Ship vertical structure
type ShipStructure = {
    decks: [Deck],
    currentDeck: int,
    spireState: SpireState,
    transitionState: Option[DeckTransition]
}

-- Spire visual state
type SpireState = {
    glowIntensity: float,    -- 0.0-1.0, pulses slowly
    activeSegment: int,      -- Which deck segment is highlighted
    mysteryLevel: int        -- Increases as game progresses (visual changes)
}

-- Deck transition animation
type DeckTransition = {
    fromDeck: int,
    toDeck: int,
    progress: float,         -- 0.0 to 1.0
    direction: TransitionDir,
    accessPoint: AccessPoint
}

type TransitionDir =
    | TransitionUp
    | TransitionDown

-- Render commands for multi-level view
pure func renderShipView(ship: ShipStructure, camera: Camera) -> [DrawCmd] {
    let currentDeck = getDeck(ship, ship.currentDeck);

    -- Layer 0: Space background (from depth system)
    let spaceBg = renderSpaceBackground(camera);

    -- Layer 1: Spire (always visible, mid-parallax)
    let spireCmds = renderSpire(ship.spireState, currentDeck.spirePosition);

    -- Layer 1.5: Deck below preview (if exists)
    let belowCmds = match getDeckBelow(ship) {
        None => [],
        Some(deck) => renderDeckPreview(deck, 0.2, -1)  -- 20% opacity, below
    };

    -- Layer 2: Current deck (full opacity)
    let currentCmds = renderDeck(currentDeck, 1.0);

    -- Layer 2.5: Deck above preview (if exists)
    let aboveCmds = match getDeckAbove(ship) {
        None => [],
        Some(deck) => renderDeckPreview(deck, 0.15, 1)  -- 15% opacity, above
    };

    -- Layer 3: Transition overlay (if transitioning)
    let transitionCmds = match ship.transitionState {
        None => [],
        Some(t) => renderTransition(t, ship)
    };

    concat([spaceBg, spireCmds, belowCmds, currentCmds, aboveCmds, transitionCmds])
}

-- Render deck at reduced opacity for preview
pure func renderDeckPreview(deck: Deck, opacity: float, offset: int) -> [DrawCmd] {
    -- Render only structural elements (walls, major furniture)
    -- Skip details (small items, decorations)
    let structuralTiles = filterStructural(deck.tiles);
    let cmds = map(\t. renderTileWithOpacity(t, opacity), structuralTiles);

    -- Apply vertical offset for visual separation
    map(\cmd. offsetVertically(cmd, offset * 32), cmds)
}
```

### Go Engine Implementation

```go
// engine/render/deck_stack.go

type DeckStackManager struct {
    decks         []*DeckData
    currentDeck   int
    spireRenderer *SpireRenderer
    previewAlpha  float64  // Opacity for adjacent deck previews
}

type DeckData struct {
    ID           int
    Level        int
    Name         string
    TileMap      [][]TileType
    SpirePos     Coord
    AccessPoints []AccessPoint
}

func NewDeckStackManager(decks []*DeckData) *DeckStackManager {
    return &DeckStackManager{
        decks:        decks,
        currentDeck:  0,
        previewAlpha: 0.2,
        spireRenderer: NewSpireRenderer(),
    }
}

func (m *DeckStackManager) RenderToLayers(
    layers *DepthLayerManager,
    camera *ParallaxCamera,
) {
    // Layer 1 (MidBackground): Spire
    m.spireRenderer.Draw(
        layers.GetBuffer(LayerMidBackground),
        m.decks[m.currentDeck].SpirePos,
        camera.ForLayer(LayerMidBackground),
    )

    // Layer 1.5: Deck below (if exists)
    if m.currentDeck > 0 {
        belowDeck := m.decks[m.currentDeck-1]
        m.renderDeckPreview(
            layers.GetBuffer(LayerMidBackground),
            belowDeck,
            m.previewAlpha,
            -1, // Below current
            camera,
        )
    }

    // Layer 2: Current deck (full opacity)
    m.renderDeck(
        layers.GetBuffer(LayerScene),
        m.decks[m.currentDeck],
        1.0,
        camera,
    )

    // Layer 2.5: Deck above (if exists)
    if m.currentDeck < len(m.decks)-1 {
        aboveDeck := m.decks[m.currentDeck+1]
        m.renderDeckPreview(
            layers.GetBuffer(LayerScene),
            aboveDeck,
            m.previewAlpha * 0.75, // Slightly more transparent
            1, // Above current
            camera,
        )
    }
}
```

```go
// engine/render/spire.go

type SpireRenderer struct {
    segments     []*ebiten.Image  // Sprite for each deck segment
    glowShader   *ebiten.Shader
    pulsePhase   float64
}

func NewSpireRenderer() *SpireRenderer {
    return &SpireRenderer{
        pulsePhase: 0,
    }
}

func (r *SpireRenderer) Update(dt float64) {
    // Slow pulse effect
    r.pulsePhase += dt * 0.5
    if r.pulsePhase > 2*math.Pi {
        r.pulsePhase -= 2 * math.Pi
    }
}

func (r *SpireRenderer) Draw(
    target *ebiten.Image,
    spirePos Coord,
    cam Transform,
) {
    // Calculate glow intensity (subtle pulse)
    glow := 0.7 + 0.3*math.Sin(r.pulsePhase)

    // Draw spire segment for current view
    screenX, screenY := cam.WorldToScreen(
        float64(spirePos.X)*TileWidth,
        float64(spirePos.Y)*TileHeight,
    )

    opts := &ebiten.DrawImageOptions{}
    opts.GeoM.Translate(screenX-SpireWidth/2, screenY-SpireHeight)
    opts.ColorScale.ScaleAlpha(float32(glow))

    // Use appropriate segment sprite based on deck
    // (different visual styles per deck region)
    target.DrawImage(r.segments[0], opts)
}

// SpireWidth and SpireHeight for the visual element
const (
    SpireWidth  = 32
    SpireHeight = 128  // Tall, visible through floors
)
```

```go
// engine/render/deck_transition.go

type DeckTransitionAnimator struct {
    active      bool
    fromDeck    int
    toDeck      int
    progress    float64  // 0.0 to 1.0
    direction   int      // -1 = down, +1 = up
    duration    float64  // seconds
    easing      EasingFunc
}

func (a *DeckTransitionAnimator) Start(from, to int, duration float64) {
    a.active = true
    a.fromDeck = from
    a.toDeck = to
    a.progress = 0
    a.duration = duration
    if to > from {
        a.direction = 1  // Going up
    } else {
        a.direction = -1 // Going down
    }
}

func (a *DeckTransitionAnimator) Update(dt float64) bool {
    if !a.active {
        return false
    }

    a.progress += dt / a.duration
    if a.progress >= 1.0 {
        a.progress = 1.0
        a.active = false
        return true // Transition complete
    }
    return false
}

// GetRenderParams returns opacity and offset for transition effect
func (a *DeckTransitionAnimator) GetRenderParams() (fromAlpha, toAlpha, offset float64) {
    // Ease the transition
    t := a.easing(a.progress)

    // Old deck fades out, new deck fades in
    fromAlpha = 1.0 - t
    toAlpha = t

    // Vertical slide effect
    offset = float64(a.direction) * (1.0 - t) * 100.0

    return
}
```

### Spire Visual Design

The spire is the emotional and narrative anchor of the ship. Its visual should:

1. **Always visible**: Through translucent floors, visible from any deck
2. **Mysterious**: Slight glow, subtle animation, otherworldly feel
3. **Central reference**: Helps with orientation (spire is always "center")
4. **Evolving**: Visual changes as game progresses and mysteries unfold

```
Spire Visual Segments (per deck):

    BRIDGE (Top)
    ┌─────────┐
    │ ▓▓▓▓▓▓▓ │  ← Navigation lattice (most visible to crew)
    │ ░░███░░ │
    └────┬────┘
         │
    HABITAT
    ┌────┴────┐
    │ ░░███░░ │  ← Data conduits (subtle pulse)
    │ ░░███░░ │
    └────┬────┘
         │
    CULTURE
    ┌────┴────┐
    │ ▓▓███▓▓ │  ← Archive interface (blue glow)
    │ ░░███░░ │
    └────┬────┘
         │
    ENGINEERING
    ┌────┴────┐
    │ ██████▓ │  ← Power feeds (warm glow)
    │ ██████▓ │
    └────┬────┘
         │
    CORE (Bottom)
    ┌────┴────┐
    │ ▓▓▓▓▓▓▓ │  ← Higgs generator (intense, restricted)
    │ ███████ │
    └─────────┘
```

### Implementation Plan

**Phase 1: Deck Stack Structure** (~4 hours)
- [ ] Create `DeckData` type with tile map and metadata
- [ ] Create `DeckStackManager` to hold all decks
- [ ] Define 5-deck test structure matching bubble ship design
- [ ] Add spire position to each deck

**Phase 2: Spire Renderer** (~4 hours)
- [ ] Create `SpireRenderer` with segment sprites
- [ ] Implement subtle pulse animation
- [ ] Draw spire to MidBackground layer
- [ ] Test: Spire visible through transparent floors

**Phase 3: Adjacent Deck Preview** (~4 hours)
- [ ] Implement `renderDeckPreview()` with opacity
- [ ] Filter to structural elements only (skip details)
- [ ] Apply vertical offset for visual separation
- [ ] Test: See ghost of deck above/below

**Phase 4: Deck Transitions** (~3 hours)
- [ ] Create `DeckTransitionAnimator`
- [ ] Implement fade + slide effect
- [ ] Connect to access point interactions
- [ ] Test: Smooth transition via stairs/elevator

**Phase 5: Integration & Polish** (~3 hours)
- [ ] Integrate with depth layer system
- [ ] Add deck indicator UI (which deck am I on?)
- [ ] Performance testing with all layers
- [ ] Documentation and examples

### Files to Modify/Create

**New files:**
- `engine/render/deck_stack.go` - Deck management (~150 LOC)
- `engine/render/spire.go` - Spire rendering (~100 LOC)
- `engine/render/deck_preview.go` - Adjacent deck previews (~80 LOC)
- `engine/render/deck_transition.go` - Transition animations (~100 LOC)
- `sim/ship_levels.ail` - AILANG deck types (~80 LOC)

**Modified files:**
- `sim/protocol.ail` - Add Deck, AccessPoint types (~40 LOC)
- `engine/render/draw.go` - Wire up deck stack rendering (~30 LOC)

## Examples

### Example 1: Bridge with Spire and Engineering Preview

**Goal:** Standing on bridge, see spire and hint of engineering deck below.

```ailang
-- On Bridge (Deck 4)
pure func renderBridgeWithContext(state: ShipState) -> [DrawCmd] {
    -- Space background (through dome)
    let spaceBg = renderSpaceBackground(state.camera);

    -- Spire (mid-parallax, always visible)
    let spire = renderSpire(state.spireState, {x: 8, y: 6});

    -- Engineering deck preview (below, 20% opacity)
    let engPreview = renderDeckPreview(state.decks[3], 0.2, -1);

    -- Bridge deck (full opacity)
    let bridge = renderDeck(state.decks[4], 1.0);

    -- Dome viewport with space view
    let dome = renderDome(state.domeView);

    concat([spaceBg, spire, engPreview, bridge, dome])
}
```

**Visual Result:**
```
┌─────────────────────────────────────────────────────────────┐
│                    OBSERVATION DOME                          │
│              ╭─────────────────────╮                        │
│             ╱  🪐  ✦  ✦  ✦    ✦   ╲                       │
│            │                       │                        │
│             ╲   Space View        ╱                         │
│              ╰─────────────────────╯                        │
│                                                             │
│    [HELM]          │║│           [NAV]                     │
│                    │║│                                      │
│                    │║│  ← SPIRE (glowing, central)         │
│                    │║│                                      │
│    [COMMS]         │║│          [STATUS]                   │
│                                                             │
│ ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │
│ ░░ Faint reactor glow (engineering preview below) ░░░░░░░ │
└─────────────────────────────────────────────────────────────┘
```

### Example 2: Deck Transition (Stairs)

**Goal:** Player uses stairs, smooth transition between decks.

```ailang
-- Player interacts with stairs on Habitat deck going to Bridge
pure func handleStairsInteraction(ship: ShipStructure, accessPoint: AccessPoint) -> ShipStructure {
    let transition = DeckTransition({
        fromDeck: ship.currentDeck,
        toDeck: accessPoint.targetDeck,
        progress: 0.0,
        direction: TransitionUp,
        accessPoint: accessPoint
    });

    { ship | transitionState: Some(transition) }
}

-- Render during transition
pure func renderTransition(t: DeckTransition, ship: ShipStructure) -> [DrawCmd] {
    let fromDeck = getDeck(ship, t.fromDeck);
    let toDeck = getDeck(ship, t.toDeck);

    -- Eased progress
    let ease = smoothstep(0.0, 1.0, t.progress);

    -- Old deck fades out with slide
    let fromCmds = renderDeckWithAlpha(fromDeck, 1.0 - ease);
    let fromOffset = if t.direction == TransitionUp then -ease * 100.0 else ease * 100.0;

    -- New deck fades in with slide
    let toCmds = renderDeckWithAlpha(toDeck, ease);
    let toOffset = if t.direction == TransitionUp then (1.0 - ease) * 100.0 else -(1.0 - ease) * 100.0;

    -- Spire remains constant (anchor point)
    let spire = renderSpire(ship.spireState, toDeck.spirePosition);

    concat([spire, offsetCmds(fromCmds, fromOffset), offsetCmds(toCmds, toOffset)])
}
```

### Example 3: Spire Mystery Progression

**Goal:** Spire visual changes as player uncovers mysteries.

```ailang
-- Spire appearance changes based on mystery level
pure func spireAppearance(mysteryLevel: int) -> SpireVisual {
    match mysteryLevel {
        0 => SpireVisual({ glow: 0.3, color: DimBlue, pulse: Slow }),
        1 => SpireVisual({ glow: 0.5, color: Blue, pulse: Medium }),    -- First clue found
        2 => SpireVisual({ glow: 0.6, color: Cyan, pulse: Medium }),    -- Archive mentions it
        3 => SpireVisual({ glow: 0.8, color: White, pulse: Fast }),     -- Major revelation
        _ => SpireVisual({ glow: 1.0, color: Gold, pulse: Intense })    -- Endgame
    }
}
```

## Success Criteria

- [ ] Spire visible through translucent floors on all 5 decks
- [ ] Deck above/below previews show at correct opacity (15-20%)
- [ ] Deck transitions are smooth (fade + slide, ~0.5s)
- [ ] Player can orient by spire position
- [ ] Access points (stairs, elevators) trigger correct transitions
- [ ] Performance: 60 FPS with current deck + 2 previews + spire
- [ ] Spire visual state updates based on game progression

## Testing Strategy

**Unit tests:**
- DeckStackManager returns correct adjacent decks
- SpireRenderer pulse calculation
- Transition animator progress tracking

**Integration tests:**
- Render all 5 decks, verify layer ordering
- Transition from deck 0 to 4, verify smooth animation
- Spire visibility through transparent tiles

**Manual testing:**
- Visual: Spire feels like central anchor
- Visual: Deck previews are subtle, not distracting
- Feel: Transitions feel smooth and oriented
- Performance: FPS counter during deck changes

**Test Scenarios:**
```bash
# Multi-deck demo
./bin/demo-ship --decks 5 --show-spire

# Transition test
./bin/demo-ship --transition-test

# Full ship exploration
./bin/demo-ship --exploration-mode
```

## Non-Goals

**Not in this feature:**
- **Full deck tile maps** - Decks use placeholder layouts, real layouts in ship-exploration.md
- **Player movement** - Handled by ship-exploration.md
- **Crew AI on other decks** - Future feature
- **Real-time deck events** - What happens on other decks while you're away
- **Spire interaction** - Late-game mystery content

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Too many layers = slow | Med | Limit previews to immediate adjacent decks |
| Previews too visible/distracting | Med | Tune opacity, filter to structural only |
| Spire obscures gameplay | Low | Place spire on open tiles, not walkways |
| Transitions feel jarring | Med | Tune easing curves, maintain spire anchor |

## References

- [isometric-depth-parallax.md](../phase1-data-models/isometric-depth-parallax.md) - Depth layer system (prerequisite)
- [viewport-compositing.md](../phase1-data-models/viewport-compositing.md) - Dome/window viewports
- [bubble-ship-design.md](../../input/bubble-ship-design.md) - Ship structure and spire
- [bubble-ship-layout.md](../../planned/future/bubble-ship-layout.md) - Deck purposes
- [ship-exploration.md](./ship-exploration.md) - Player movement and deck interaction
- [02-bridge-interior.md](./02-bridge-interior.md) - Bridge deck design

## Future Work

- **Crew on Other Decks** - See crew silhouettes on adjacent decks
- **Deck Events** - Visual hints of activity (lights, movement) on other decks
- **Spire Interaction** - Late-game mechanics involving the spire
- **Dynamic Deck States** - Damage, modifications visible in previews
- **Vertical Sound Design** - Audio cues from other decks (footsteps above, machinery below)

---

**Document created**: 2025-12-12
**Last updated**: 2025-12-12
