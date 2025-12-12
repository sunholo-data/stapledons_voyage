# Sprint: Multi-Level Ship Visualization

**Design Doc:** [design_docs/planned/phase2-core-views/multi-level-ship-visualization.md](../design_docs/planned/phase2-core-views/multi-level-ship-visualization.md)
**Duration:** 3-4 days
**Priority:** P1 (Foundation for Ship Exploration)
**Dependencies:**
- Isometric Depth & Parallax System (must complete first)
- Viewport Compositing (must complete first)

## Goal

Enable visualization of multi-level ship structure with the spire as central anchor, showing adjacent deck hints and maintaining spatial coherence across deck transitions.

## Pre-Sprint Checklist

- [ ] Verify Depth & Parallax sprint complete
- [ ] Verify Viewport Compositing sprint complete
- [ ] DepthLayerManager working with 4 layers
- [ ] SpireRenderer pattern from parallax sprint working
- [ ] Check for unread AILANG messages: `ailang messages list --unread`

## Day 1: Deck Stack Structure (~4 hours)

### Tasks
- [ ] Create AILANG deck types in `sim/ship_levels.ail`
  ```ailang
  type Deck = {
      id: int,
      name: string,
      level: int,
      spirePosition: Coord,
      accessPoints: [AccessPoint]
  }

  type AccessType = Stairs | Elevator | Hatch | Ladder

  type ShipStructure = {
      decks: [Deck],
      currentDeck: int,
      spireState: SpireState
  }
  ```
- [ ] Create `engine/render/deck_stack.go` with DeckStackManager
- [ ] Define 5-deck test structure matching bubble ship design:
  - Deck 0: Core (restricted)
  - Deck 1: Engineering
  - Deck 2: Culture (Archive, Labs)
  - Deck 3: Habitat (Crew Quarters)
  - Deck 4: Bridge
- [ ] Add spire position to each deck (central tile)

### Files to Create
- `sim/ship_levels.ail` (~80 LOC) - NEW
- `engine/render/deck_stack.go` (~150 LOC) - NEW

### Verification
```bash
ailang check sim/ship_levels.ail
make sim
go build ./engine/render/...
```

## Day 2: Spire Renderer (~4 hours)

### Tasks
- [ ] Create `engine/render/spire.go` with SpireRenderer
- [ ] Implement subtle pulse animation
  ```go
  func (r *SpireRenderer) Update(dt float64) {
      r.pulsePhase += dt * 0.5
      glow := 0.7 + 0.3*math.Sin(r.pulsePhase)
  }
  ```
- [ ] Create spire segment sprites (one per deck region)
  - Bridge: Navigation lattice (blue-white)
  - Habitat: Data conduits (subtle pulse)
  - Culture: Archive interface (blue glow)
  - Engineering: Power feeds (warm glow)
  - Core: Higgs generator (intense)
- [ ] Draw spire to MidBackground layer
- [ ] Test: Spire visible through transparent floors

### Files to Create
- `engine/render/spire.go` (~100 LOC) - NEW
- `assets/sprites/spire/` - Segment sprites

### Verification
```bash
./bin/game --test-spire
# Visual: Spire visible, pulsing gently, visible through glass floor
```

## Day 3: Adjacent Deck Preview (~4 hours)

### Tasks
- [ ] Create `engine/render/deck_preview.go`
- [ ] Implement renderDeckPreview() with opacity parameter
- [ ] Filter to structural elements only (walls, major furniture)
- [ ] Apply vertical offset for visual separation
- [ ] Render deck below at 20% opacity
- [ ] Render deck above at 15% opacity
- [ ] Test: Standing on Habitat, see hint of Engineering below and Bridge above

### Files to Create
- `engine/render/deck_preview.go` (~80 LOC) - NEW

### Verification
```bash
./bin/game --test-deck-preview --deck 3
# Visual: Current deck clear, faint silhouettes above/below
```

## Day 4: Transitions & Integration (~4 hours)

### Tasks
- [ ] Create `engine/render/deck_transition.go`
- [ ] Implement DeckTransitionAnimator
  - Fade old deck out (alpha 1.0 → 0.0)
  - Fade new deck in (alpha 0.0 → 1.0)
  - Vertical slide effect (100px offset)
  - Duration: ~0.5 seconds
- [ ] Connect to access point interactions (stairs, elevator)
- [ ] Keep spire constant during transition (anchor point)
- [ ] Add deck indicator UI (optional - show which deck)
- [ ] Performance testing with all layers
- [ ] Documentation

### Files to Create
- `engine/render/deck_transition.go` (~100 LOC) - NEW

### Verification
```bash
./bin/game --test-transition
# Press 1-5 to switch decks
# Visual: Smooth fade + slide, spire stays constant
```

## Success Criteria

- [ ] Spire visible through translucent floors on all 5 decks
- [ ] Deck above/below previews show at correct opacity (15-20%)
- [ ] Deck transitions are smooth (fade + slide, ~0.5s)
- [ ] Player can orient by spire position
- [ ] Access points (stairs, elevators) trigger correct transitions
- [ ] Performance: 60 FPS with current deck + 2 previews + spire
- [ ] Spire visual changes work (mysteryLevel affects appearance)

## AILANG Feedback Checkpoint

After sprint, report:
- [ ] List operations performance (filtering structural tiles)
- [ ] Record update performance (transition state updates)
- [ ] Any issues with nested Option types

## Handoff

This sprint enables:
- **Bridge Interior** - Full context: dome + space + spire + deck below
- **Ship Exploration** - Can navigate between decks with spatial awareness
- **Crew Quarters** - Window system + deck context working

## Visual Reference

```
Bridge Deck (Deck 4) with full context:

┌─────────────────────────────────────────────────────────────┐
│                    OBSERVATION DOME                          │
│              ╭─────────────────────╮                        │
│             ╱  🪐  ✦  ✦  ✦    ✦   ╲                       │
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

---

**Sprint created:** 2025-12-12
**Status:** Ready for execution (after Depth & Viewport sprints)
