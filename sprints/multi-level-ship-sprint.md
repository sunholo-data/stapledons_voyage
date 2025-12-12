# Multi-Level Ship Visualization Sprint

**Design Doc:** [design_docs/planned/phase2-core-views/multi-level-ship-visualization.md](../design_docs/planned/phase2-core-views/multi-level-ship-visualization.md)
**Status:** Completed
**Created:** 2025-12-12
**Completed:** 2025-12-12
**Estimated LOC:** ~770
**Actual LOC:** 1140

## Overview

Implement the 5-deck ship structure with Higgs Spire as central anchor. Players navigate between decks (Core → Engineering → Culture → Habitat → Bridge) with smooth transitions and preview system.

## Dependencies

- [x] Depth & Parallax System (engine/depth/, engine/render/depth_layers.go)
- [x] Viewport Compositing (engine/render/viewport_*.go)

## Day 1: AILANG Ship Level Types (~160 LOC actual)

**Goal:** Define deck structure and level types in AILANG

### Tasks
- [x] Create `sim/ship_levels.ail` with:
  - [x] `DeckType` ADT: Core, Engineering, Culture, Habitat, Bridge
  - [x] `DeckInfo` record: name, description, colorScheme, yOffset
  - [x] `ShipLevels` record: currentDeck, transitionState, spireGlow
  - [x] `get_deck_info(deck: DeckType) -> DeckInfo` function (renamed to avoid codegen collision)
  - [x] `init_ship_levels() -> ShipLevels` function
- [x] Update `sim/world.ail` to include ShipLevels in World state
- [x] Run `make sim` to verify codegen

### Files
- `sim/ship_levels.ail` (NEW - 160 LOC)
- `sim/world.ail` (MODIFIED)
- `sim/step.ail` (MODIFIED - imports & init_world)

## Day 2: Deck Stack Renderer + Higgs Spire (~419 LOC actual)

**Goal:** Engine components for rendering deck layers and the central spire

### Tasks
- [x] Create `engine/render/deck_stack.go`:
  - [x] `DeckStackRenderer` struct with buffers per deck
  - [x] `RenderDeckStack(currentDeck, transitionProgress, targetDeck, renderDeck)` method
  - [x] Parallax offset based on deck distance from current
  - [x] Integration with DepthLayerManager
- [x] Create `engine/render/spire.go`:
  - [x] `HiggsSpire` struct for the central visual anchor
  - [x] `Render(screen, currentDeck, transitionProgress, targetDeck)` method
  - [x] Spire segments that light up based on current deck
  - [x] Color interpolation and glow effects

### Files
- `engine/render/deck_stack.go` (NEW - 197 LOC)
- `engine/render/spire.go` (NEW - 222 LOC)

## Day 3: Deck Preview System (~286 LOC actual)

**Goal:** Show adjacent deck previews at screen edges

### Tasks
- [x] Create `engine/render/deck_preview.go`:
  - [x] `DeckPreview` struct for edge previews
  - [x] `RenderPreviews(screen, currentDeck, transitionProgress, targetDeck)` method
  - [x] Semi-transparent peek at adjacent decks
  - [x] Edge fade using gradient rendering

### Files
- `engine/render/deck_preview.go` (NEW - 286 LOC)

## Day 4: Deck Transitions (~275 LOC actual)

**Goal:** Smooth animated transitions between decks

### Tasks
- [x] Create `engine/render/deck_transition.go`:
  - [x] `DeckTransition` struct with animation state
  - [x] `StartTransition(from, to, duration)` method
  - [x] `Update(deltaTime)` for animation progress
  - [x] Fade + slide composite effect
  - [x] Cubic ease-in-out easing
- [x] AILANG transition support (already in ship_levels.ail):
  - [x] `TransitionState` ADT: TransitionIdle, Transitioning(from, to, progress)
  - [x] `start_deck_transition(levels, target)` function
  - [x] `update_transition(levels, dt)` function

### Files
- `engine/render/deck_transition.go` (NEW - 275 LOC)

## Day 5: Integration & Polish

**Goal:** Wire everything together and polish

### Tasks
- [x] Run `make build` to verify everything compiles
- [ ] Full game loop integration (deferred to future sprint)

### Files
- Build verification complete

## Acceptance Criteria

From design doc:
- [x] 5-deck structure defined with Higgs Spire as anchor
- [x] DeckStackRenderer supports parallax rendering of multiple decks
- [x] Adjacent decks visible as previews at screen edges (DeckPreview)
- [x] Smooth transitions between decks (DeckTransition with fade + slide)
- [x] Spire segments indicate current deck position (HiggsSpire)
- [x] Works with existing depth layer system

## Lessons Learned

1. AILANG function names export with CamelCase conversion - avoid naming functions same as types (deck_info → DeckInfo collision)
2. Use helper functions (makeDeckInfo) to avoid inline record literals in match arms
3. Record updates inside match arms don't parse correctly - use helper functions

## Notes

- Full game loop wiring deferred - all engine components ready for integration
- Leverages existing depth layer system for deck compositing
- Spire uses vector graphics with color interpolation for glow effects
