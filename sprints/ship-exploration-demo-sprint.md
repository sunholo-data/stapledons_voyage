# Sprint: Ship Exploration Demo

**Design Doc:** [design_docs/planned/phase2-core-views/multi-level-ship-visualization.md](../design_docs/planned/phase2-core-views/multi-level-ship-visualization.md)
**Sprint Duration:** 2-3 days
**Priority:** P1 (Foundation for Ship Experience)
**Goal:** Standalone demo where player walks around isometric ship decks and moves between levels

## Executive Summary

Create `demo-ship-explore` combining:
- Existing isometric rendering (64x32 diamond tiles)
- AILANG-driven game state (player position, current deck, NPCs)
- AI-generated ship interior assets
- Deck transitions via stairs/elevator access points

## Prerequisites

| Dependency | Status | Notes |
|------------|--------|-------|
| Isometric engine | ✅ Done | `engine/render/iso.go`, `draw_iso.go` |
| IsoTile/IsoEntity DrawCmds | ✅ Done | `sim/protocol.ail` |
| iso_demo.ail | ✅ Done | Player movement, camera follow |
| ship_levels.ail | ✅ Done | DeckType, transitions |
| Asset manager skill | ✅ Done | AI image generation |
| Entity sprites (NPCs) | ✅ Done | IDs 100-105 |

## Day 1: Asset Generation & AILANG State ✅ COMPLETE

### Phase 1.1: Generate Ship Interior Tile Assets (~2 hours) ✅

Created 64x32 isometric floor tiles for ship decks:

- [x] **Bridge floor tile** - Blue LED circuit patterns (`iso_tiles/bridge_floor.png`)
- [x] **Engineering floor tile** - Industrial blue-teal patterns (`iso_tiles/engineering_floor.png`)
- [x] **Culture floor tile** - Green botanical patterns (`iso_tiles/culture_floor.png`)
- [x] **Habitat floor tile** - Blue residential patterns (`iso_tiles/habitat_floor.png`)
- [x] **Core floor tile** - Dark with red/orange warning markings (`iso_tiles/core_floor.png`)
- [x] **Access point tile** - Purple portal with checkered pattern (`iso_tiles/access_point.png`)

### Phase 1.2: Update Manifest (~30 min) ✅

- [x] Assets stored in `assets/sprites/iso_tiles/`
- [x] Added to manifest.json with IDs 30-35
- [x] Verified loading in demo screenshot

### Phase 1.3: AILANG Ship Demo State (~3 hours) ✅

Created `sim/ship_demo.ail` with all functions prefixed with 'sd' to avoid namespace collisions:

- [x] Defined `ShipDemoState` type with player position, deck, transitions, camera
- [x] 8x8 deck grid with access point at center (4,4)
- [x] `initShipDemo() -> ShipDemoState` - Start on Bridge deck
- [x] `stepShipDemo(state, input) -> ShipDemoState` - WASD movement, E for transitions
- [x] `renderShipDemo(state) -> [DrawCmd]` - Tiles, player, NPCs
- [x] `getShipCamera(state) -> Camera` - Smooth camera follow
- [x] NPCs defined per deck (colored circles as fallback)

## Day 2: Go Demo Integration & Polish ✅ MOSTLY COMPLETE

### Phase 2.1: Create demo-ship-explore (~2 hours) ✅

- [x] Created `cmd/demo-ship-explore/main.go`
- [x] Wire up AILANG handlers (Rand, Clock, Debug, AI)
- [x] Initialize state via `sim_gen.InitShipDemo()`
- [x] Game loop: capture input → step → render
- [x] Isometric rendering with sprite fallbacks

### Phase 2.2: Add HUD & Controls (~1 hour) ✅

- [x] Display current deck name and level
- [x] Show transition progress (when transitioning)
- [x] Display controls help text (WASD, E)
- [ ] Add F1 for debug overlay (optional polish)

### Phase 2.3: Deck Transition Visuals (~2 hours) ⏳ BASIC

- [x] Basic deck transition (teleport to new deck)
- [ ] Fade effects (optional polish)
- [ ] Slide/vertical offset (optional polish)

### Phase 2.4: Screenshot & Verification (~1 hour) ✅

- [x] Add --screenshot flag support
- [x] Verified screenshot of Bridge deck
- [x] Verified all tile assets loading correctly
- [ ] Screenshot all 5 decks (manual testing)

## Day 3: NPC Integration & Polish ⏳ PARTIAL

### Phase 3.1: Add NPCs to Decks (~2 hours) ✅

- [x] Defined NPC positions per deck in AILANG (`sdNpcsForDeck`)
- [x] NPCs rendered as IsoEntity on each deck
- [x] Using colored circle fallbacks (NPC sprites IDs 100-105 available but need proper sprite loading)
- [x] NPCs stay on their assigned deck

### Phase 3.2: Optional: NPC Movement (~2 hours)

- [ ] Add simple patrol patterns (from npc_ai.ail)
- [ ] NPCs walk between waypoints
- [ ] Player collision detection

### Phase 3.3: Final Polish (~1 hour)

- [x] Camera follow smoothness (0.15 lerp factor)
- [x] Transition duration (0.05 per frame, ~20 frames)
- [ ] Test edge cases (rapid deck switching)
- [ ] Final screenshot suite

## Files Created ✅

| File | Purpose | LOC |
|------|---------|-----|
| `sim/ship_demo.ail` | AILANG game state & logic | ~268 |
| `cmd/demo-ship-explore/main.go` | Go demo entry point | ~310 |
| `assets/sprites/iso_tiles/*.png` | 6 interior tiles | Generated |

## Files Modified ✅

| File | Change |
|------|--------|
| `assets/sprites/manifest.json` | Added ship tile IDs 30-35 |
| `sim_gen/ship_demo.go` | Regenerated from AILANG |

## Success Criteria

- [x] Player can walk around each of 5 decks using WASD
- [x] Access points allow moving between decks (E key at position 4,4)
- [x] Deck transitions work (basic teleport)
- [x] Each deck has distinct floor tile appearance
- [x] NPCs visible on each deck (colored circles as fallback)
- [x] 60 FPS maintained
- [x] All state managed in AILANG (not Go)
- [x] Screenshot verification of Bridge deck

## Sprite ID Allocation

| ID | Name | File |
|----|------|------|
| 30 | Bridge Floor | `ship_tiles/bridge_floor.png` |
| 31 | Engineering Floor | `ship_tiles/engineering_floor.png` |
| 32 | Culture Floor | `ship_tiles/culture_floor.png` |
| 33 | Habitat Floor | `ship_tiles/habitat_floor.png` |
| 34 | Core Floor | `ship_tiles/core_floor.png` |
| 35 | Access Point | `ship_tiles/access_point.png` |

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| AI generates wrong style tiles | Use --dry-run to preview prompts, iterate on style guide |
| AILANG codegen issues | Report via ailang-feedback, continue with working parts |
| Performance with 5 decks | Only render current deck + faded previews |
| Complex transition math | Reuse existing DeckTransition from game_views/ |

## Notes

- This demo proves the full pipeline: AILANG state → Go rendering → AI assets
- Foundation for full ship exploration (future: NPC dialogue, interactions)
- Intentionally small decks (8x8) to keep scope manageable
- Spire rendering is OPTIONAL for this sprint (nice-to-have)
