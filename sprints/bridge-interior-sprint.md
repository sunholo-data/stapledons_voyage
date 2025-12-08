# Sprint: Bridge Interior

**Sprint ID:** bridge-interior-v1
**Design Doc:** [design_docs/planned/next/02-bridge-interior.md](../design_docs/planned/next/02-bridge-interior.md)
**Duration:** 8 working sessions (compressed from 10 days due to existing infrastructure)
**Priority:** P0 - Core Player Experience

## Goal

Implement the bridge as the first ship interior view, with:
- 16x12 isometric tile grid
- Observation dome showing space/planet exterior
- 5 crew NPCs at stations + captain
- Player movement with WASD
- Console interactions (hover/click)
- AI-generated visual assets

## Prerequisites

**Already Complete:**
- [x] Isometric projection engine (`engine/render/iso.go`)
- [x] IsoTile/IsoEntity DrawCmd rendering (`engine/render/draw_iso.go`)
- [x] View system with ViewBridge type (`engine/view/view.go`)
- [x] SpaceView with planet rendering (`engine/view/space_view.go`)
- [x] Sprite manifest system (`assets/sprites/manifest.json`)
- [x] Animation support for entities

**AILANG Status:**
- [x] `sim/protocol.ail` compiles (IsoTile, IsoEntity types ready)
- [ ] Need new `sim/bridge.ail` module

## Session 1: AILANG Bridge Types

**Focus:** Define all bridge-related types in AILANG

### Tasks
- [ ] Create `sim/bridge.ail` module
- [ ] Define `BridgeState` record type
- [ ] Define `BridgeStation` ADT (Helm, Comms, Status, Nav, Science, Captain)
- [ ] Define `CrewPosition` record with station assignment
- [ ] Define `ConsoleState` for interactive consoles
- [ ] Define `DomeViewState` for observation dome
- [ ] Run `ailang check sim/bridge.ail` - verify compiles
- [ ] Export types via `sim/protocol.ail`

### Files to Create/Modify
```
sim/bridge.ail          # NEW - Bridge state types
sim/protocol.ail        # MODIFY - Export bridge types
```

### Acceptance Criteria
- [ ] `ailang check sim/bridge.ail` passes
- [ ] `ailang check sim/protocol.ail` passes with bridge imports

---

## Session 2: Bridge Layout Data

**Focus:** Define the 16x12 bridge tile layout and console positions

### Tasks
- [ ] Create bridge floor layout as tile ID array (16x12 = 192 tiles)
- [ ] Define console positions (5 stations + captain chair)
- [ ] Define crew station assignments
- [ ] Define walkable vs blocked tiles
- [ ] Add `initBridge()` function to create initial BridgeState
- [ ] Run `ailang check`

### Bridge Layout Plan
```
Row 0-1:  Dome edge tiles (1004)
Row 2:    Console stations (1002) - Helm, Comms, Status
Row 3-4:  Floor (1000) with walkway (1003)
Row 5:    Central walkway (1003)
Row 6-7:  Floor (1000) with captain area (1007)
Row 8-9:  Side stations - Nav, Science
Row 10-11: Access hatches (1006) and walls
```

### Files to Modify
```
sim/bridge.ail          # Add layout data, initBridge()
```

### Acceptance Criteria
- [ ] `initBridge()` returns valid BridgeState
- [ ] All 192 tiles have assigned IDs
- [ ] Console positions defined with correct tile coords

---

## Session 3: Bridge Rendering

**Focus:** Implement AILANG functions to generate DrawCmds for bridge

### Tasks
- [ ] Implement `renderBridgeFloor(layout) -> [DrawCmd]`
- [ ] Implement `renderConsoles(consoles) -> [DrawCmd]`
- [ ] Implement `renderBridgeCrew(positions) -> [DrawCmd]`
- [ ] Implement `renderPlayer(pos, facing) -> DrawCmd`
- [ ] Implement `renderBridge(state) -> [DrawCmd]` combining all layers
- [ ] Test with `ailang run` or mock integration

### Files to Modify
```
sim/bridge.ail          # Add render functions
sim/step.ail            # Add bridge mode handling
```

### Acceptance Criteria
- [ ] `renderBridge` produces correct DrawCmd list
- [ ] Tiles rendered with proper IsoTile commands
- [ ] Entities rendered with IsoEntity commands

---

## Session 4: Go BridgeView Implementation

**Focus:** Create Go BridgeView type implementing View interface

### Tasks
- [ ] Create `engine/view/bridge_view.go`
- [ ] Implement `BridgeView` struct with isometric renderer reference
- [ ] Implement `View` interface (Init, Enter, Exit, Update, Draw)
- [ ] Wire bridge state from sim_gen to renderer
- [ ] Add BridgeView to ViewManager registration
- [ ] Create `cmd/demo-bridge/main.go` for testing

### Files to Create/Modify
```
engine/view/bridge_view.go    # NEW - BridgeView implementation
engine/view/manager.go        # MODIFY - Register BridgeView
cmd/demo-bridge/main.go       # NEW - Demo command
Makefile                      # MODIFY - Add demo-bridge target
```

### Acceptance Criteria
- [ ] `make demo-bridge` builds and runs
- [ ] Bridge floor tiles render as isometric grid
- [ ] Placeholder diamonds visible for tiles without sprites

---

## Session 5: Observation Dome

**Focus:** Render space view inside the bridge dome area

### Tasks
- [ ] Create `engine/view/dome_renderer.go`
- [ ] Define dome bounds (top 4 rows of bridge, elliptical mask)
- [ ] Render SpaceView to offscreen buffer
- [ ] Apply circular/elliptical mask shader
- [ ] Composite masked dome onto bridge view
- [ ] Pass DomeViewState from AILANG (target planet, velocity)

### Files to Create/Modify
```
engine/view/dome_renderer.go  # NEW - Dome rendering with mask
engine/view/bridge_view.go    # MODIFY - Integrate dome
engine/shader/dome_mask.go    # NEW (optional) - Mask shader
```

### Acceptance Criteria
- [ ] Dome shows planets from SpaceView
- [ ] Circular/elliptical mask clips space view
- [ ] Dome integrates with bridge isometric layout
- [ ] SR effects visible when velocity > 0

---

## Session 6: Player Movement

**Focus:** WASD movement with collision detection

### Tasks
- [ ] Add `processBridgeInput(state, input) -> BridgeState` to AILANG
- [ ] Implement direction detection from WASD keys
- [ ] Implement collision check against walkable tiles
- [ ] Update player position on valid moves
- [ ] Update player facing direction
- [ ] Wire input through FrameInput in Go

### Files to Modify
```
sim/bridge.ail              # Add input processing
engine/input/input.go       # Ensure WASD captured
engine/view/bridge_view.go  # Wire input to AILANG
```

### Acceptance Criteria
- [ ] Player moves with WASD
- [ ] Cannot walk through consoles/walls
- [ ] Player sprite faces movement direction
- [ ] Movement feels smooth (sub-tile interpolation)

---

## Session 7: Console Interactions

**Focus:** Hover highlights and click handling

### Tasks
- [ ] Implement `findInteractableAt(state, tileX, tileY) -> Option[InteractableID]`
- [ ] Add hover state to BridgeState
- [ ] Render highlight effect for hovered console
- [ ] Implement click handling - returns interaction result
- [ ] Define `BridgeInputResult` ADT (Stay, TransitionToGalaxyMap, TransitionToDialogue)
- [ ] Wire click results to ViewTransition in Go

### Files to Modify
```
sim/bridge.ail              # Add interaction logic
engine/view/bridge_view.go  # Handle interaction results
```

### Acceptance Criteria
- [ ] Consoles highlight on hover
- [ ] Click on Nav console triggers galaxy map transition (placeholder)
- [ ] Click on crew triggers dialogue transition (placeholder)
- [ ] Interaction cursor changes near interactables

---

## Session 8: Asset Generation & Polish

**Focus:** Generate AI art assets and final polish

### Tasks

**Asset Generation (use asset-manager skill):**
- [ ] Generate bridge floor tiles (1000-1007)
- [ ] Generate console sprites (1100-1105)
- [ ] Generate crew sprites (1200-1205) with animations
- [ ] Generate player sprite with 4-direction walk
- [ ] Update `assets/sprites/manifest.json` with new IDs

**Polish:**
- [ ] Add room entry label ("BRIDGE" appears on entry)
- [ ] Add transition animation from space view
- [ ] Add ambient console glow effects
- [ ] Performance test - verify 60 FPS
- [ ] Update design doc with implementation notes

### Files to Create/Modify
```
assets/sprites/bridge/          # NEW directory
assets/sprites/manifest.json    # MODIFY - Add bridge sprites
design_docs/planned/next/02-bridge-interior.md  # MODIFY - Mark complete
```

### Acceptance Criteria
- [ ] All placeholder sprites replaced with AI-generated art
- [ ] Visual style consistent with game aesthetic
- [ ] 60 FPS with full bridge rendering
- [ ] Smooth transition from exterior view

---

## AILANG Feedback Checkpoints

### Session 1 Check
After defining types, report any issues with:
- ADT syntax for BridgeStation
- Record field access patterns
- Import/export of types between modules

### Session 3 Check
After implementing render functions, report any issues with:
- List concatenation performance
- Recursive rendering of tiles
- DrawCmd construction patterns

### Session 6 Check
After implementing input handling, report any issues with:
- FrameInput field access
- Match expression on key events
- State update patterns

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| AILANG codegen issues | Medium | High | Fall back to mock sim_gen |
| Dome masking complex | Medium | Medium | Start with rectangular clip |
| Asset generation quality | Low | Medium | Iterate with refined prompts |
| Performance with full scene | Low | Low | Viewport culling already works |

---

## Dependencies

**Blocked By:**
- Nothing - all prerequisites complete

**Blocks:**
- Ship Exploration (uses bridge as template)
- Crew Dialogue (triggered from bridge)
- Galaxy Map (accessed from Nav console)
- Arrival Sequence completion (ends at bridge)

---

## Success Metrics

- [ ] Bridge renders with all visual layers
- [ ] Player can walk around bridge with WASD
- [ ] Observation dome shows space/planets
- [ ] Crew visible at stations
- [ ] Consoles are interactable
- [ ] 60 FPS maintained
- [ ] All assets AI-generated (no placeholders)
- [ ] Demo command works: `make demo-bridge`

---

## Post-Sprint

After completion:
1. Move design doc to `design_docs/implemented/v0_3_0/`
2. Update CLAUDE.md with bridge view notes
3. Create sprint for Ship Exploration (extending bridge patterns)
4. Report AILANG feedback summary via `ailang-feedback` skill
