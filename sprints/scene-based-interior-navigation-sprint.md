# Sprint: Scene-Based Interior Navigation

**Sprint ID:** `scene-based-interior-navigation`
**Design Doc:** [design_docs/planned/scene-based-interior-navigation.md](../design_docs/planned/scene-based-interior-navigation.md)
**Target:** v0.6.0
**Estimated Duration:** 5-7 days
**Priority:** P0 - Critical (core gameplay experience)
**Created:** 2025-12-20

## Sprint Goal

Replace 3D interior navigation with scene-based 2D/2.5D deck system. Player navigates by selecting decks from ship UI. Outward decks (Bridge, Observation) composite live starmap in windows. Crew placement is contextual based on mood/needs.

## Vision Alignment

- **Pillar 3 (Time Has Emotional Weight):** +2 - Visual contrast cosmic/human scale
- **Pillar 4 (The Ship Is Home):** +2 - Crew in environments, intimate conversations
- **Pillar 6 (We Are Not Built For This):** +1 - Claustrophobia, psychological strain
- **Net Score:** +5 ✅

## Success Criteria

- [x] **Prototype validated** - demo-engine-scene-bridge proves window compositing works
- [ ] Player can switch between Bridge and Observation Deck via UI
- [ ] Deck scenes render with AI-generated backgrounds
- [ ] Windows composite live starmap (planets, stars, SR/GR effects)
- [ ] Crew sprites appear on decks based on mood/time
- [ ] Smooth transition between exterior (starmap) and interior (deck) modes
- [ ] Can easily add new decks by dropping in assets + manifest

## Pre-Sprint Checklist

- [x] **Vision interview complete** - decisions on MVP scope, crew rendering, art style
- [x] **Assets created** - Bridge & Observation backgrounds + window masks (70s comic style)
- [x] **Prototype working** - demo-engine-scene-bridge validates window compositing
- [ ] **Check AILANG messages** - Review any pending issues from AILANG team
- [ ] **Acknowledge codegen message** - Non-exported function namespacing issue

---

## Day 1: AILANG Types & Core Data (REPLACE sim/interior.ail)

**Goal:** Define new scene-based types, replacing 3D interior module

### Tasks

- [ ] **Backup old 3D interior code**
  ```bash
  mv sim/interior.ail sim/interior_3d_DEPRECATED.ail
  mv sim/interior_themes.ail sim/interior_themes_DEPRECATED.ail
  mv sim/window.ail sim/window_DEPRECATED.ail
  ```

- [ ] **Create new sim/interior.ail** with scene-based types
  - `DeckID` - enum: Bridge, ObservationDeck, Engineering, Medical, Quarters, Archive, Hydroponics
  - `DeckType` - enum: Outward (has windows), Inward (enclosed)
  - `DeckInfo` - metadata: deckID, deckType, name, description, capacity
  - `WindowRegion` - for outward decks: x, y, width, height (normalized 0-1)
  - `CrewSpawnPoint` - x, y, role (captain, pilot, engineer, medic, etc.)

- [ ] **Define DeckState** - runtime state per deck
  - `deckID: DeckID`
  - `crewPresent: [CrewID]` - which crew are on this deck right now
  - `lastVisited: float` - game time

- [ ] **Define InteriorState** - global interior navigation state
  - `currentDeck: DeckID`
  - `deckStates: Map<DeckID, DeckState>` (or list with helpers)
  - `transitionProgress: Option<float>` - for fade transitions

- [ ] **Add to World type** in sim/world.ail
  ```ailang
  export type World = {
    ...existing fields...
    interior: InteriorState
  }
  ```

- [ ] **Type-check:** `ailang check sim/interior.ail`

**Estimated:** 4-6 hours

**AILANG Complexity:** Low - basic ADT types, no complex recursion

---

## Day 2: Deck Metadata & Crew Placement Logic

**Goal:** Hardcode deck metadata, implement crew location logic

### Tasks

- [ ] **Create getDeckInfo() function** - return DeckInfo for each deck
  ```ailang
  pure func getDeckInfo(deckID: DeckID) -> DeckInfo {
    match deckID {
      Bridge => {
        deckID: Bridge,
        deckType: Outward,
        name: "Bridge",
        description: "Command center with panoramic windows",
        capacity: 10
      }
      ObservationDeck => { ... }
      ...
    }
  }
  ```

- [ ] **Create getWindowRegions() function** - for outward decks
  ```ailang
  pure func getWindowRegions(deckID: DeckID) -> [WindowRegion] {
    match deckID {
      Bridge => [{ x: 0.1, y: 0.0, width: 0.8, height: 0.5 }]
      ObservationDeck => [{ x: 0.05, y: 0.0, width: 0.9, height: 0.6 }]
      _ => []
    }
  }
  ```

- [ ] **Create getCrewSpawnPoints() function** - position hints per deck
  ```ailang
  pure func getCrewSpawnPoints(deckID: DeckID) -> [CrewSpawnPoint] {
    match deckID {
      Bridge => [
        { x: 512, y: 650, role: "captain" },
        { x: 300, y: 650, role: "pilot" },
        ...
      ]
      ObservationDeck => [
        { x: 400, y: 650, role: "any" },
        { x: 900, y: 650, role: "any" },
        ...
      ]
      ...
    }
  }
  ```

- [ ] **Implement getCrewDeck() function** - determine where crew should be
  ```ailang
  -- Based on crew mood, time, needs, events
  func getCrewDeck(crew: Crew, time: float, recentEvents: [Event]) -> DeckID {
    match crew.mood {
      Stressed => Quarters  -- or MedicalBay
      Contemplative => ObservationDeck
      Focused => getWorkDeck(crew.role)
      Social => Bridge  -- or ObservationDeck
      ...
    }
  }
  ```

- [ ] **Helper: getWorkDeck()** - crew's primary work location
  ```ailang
  pure func getWorkDeck(role: CrewRole) -> DeckID {
    match role {
      Captain => Bridge
      Pilot => Bridge
      Engineer => Engineering
      Medic => Medical
      ...
    }
  }
  ```

- [ ] **Type-check:** `ailang check sim/interior.ail`

**Estimated:** 4-6 hours

**AILANG Complexity:** Medium - nested pattern matching, list building

---

## Day 3: Deck Navigation & State Management

**Goal:** Implement deck switching, update world state

### Tasks

- [ ] **Create initInterior() function**
  ```ailang
  export func initInterior() -> InteriorState {
    {
      currentDeck: Bridge,  -- Start on bridge
      deckStates: initDeckStates(),
      transitionProgress: None
    }
  }
  ```

- [ ] **Helper: initDeckStates()** - initialize all deck states
  ```ailang
  func initDeckStates() -> [DeckState] {
    [
      { deckID: Bridge, crewPresent: [], lastVisited: 0.0 },
      { deckID: ObservationDeck, crewPresent: [], lastVisited: 0.0 },
      ...
    ]
  }
  ```

- [ ] **Create changeDeck() function** - handle deck navigation
  ```ailang
  func changeDeck(interior: InteriorState, newDeck: DeckID, time: float) -> InteriorState {
    let updatedStates = updateLastVisited(interior.deckStates, newDeck, time);
    {
      currentDeck: newDeck,
      deckStates: updatedStates,
      transitionProgress: Some(0.0)  -- Start fade transition
    }
  }
  ```

- [ ] **Create updateCrewLocations() function** - assign crew to decks each tick
  ```ailang
  func updateCrewLocations(
    interior: InteriorState,
    crew: [Crew],
    time: float,
    events: [Event]
  ) -> InteriorState {
    -- For each crew member, determine their current deck
    -- Update deckStates.crewPresent lists
    ...
  }
  ```

- [ ] **Update step() in sim/step.ail** to call updateCrewLocations()

- [ ] **Add input handling** for deck navigation in sim/step.ail
  ```ailang
  -- If player presses '1', go to Bridge
  -- If player presses '2', go to Observation Deck
  -- (Or handle via UI button clicks)
  ```

- [ ] **Type-check:** `ailang check sim/*.ail`
- [ ] **Test:** `ailang run --entry step sim/step.ail` (if possible)

**Estimated:** 5-7 hours

**AILANG Complexity:** High - list updates, nested record updates

**Potential Blocker:** List/Map manipulation might be verbose without helper functions

---

## Day 4: DrawCmd Generation for Deck Rendering

**Goal:** Generate DrawCmds for deck scenes, crew sprites

### Tasks

- [ ] **Create renderDeck() function** - main rendering entry point
  ```ailang
  export func renderDeck(
    interior: InteriorState,
    crew: [Crew],
    shipVelocity: Vec3,
    shipPosition: Vec3
  ) -> [DrawCmd] {
    let deckID = interior.currentDeck;
    let deckInfo = getDeckInfo(deckID);

    concat([
      renderDeckBackground(deckID),
      renderWindows(deckID, shipVelocity, shipPosition),
      renderCrewOnDeck(deckID, interior, crew),
      renderDeckUI(deckInfo)
    ])
  }
  ```

- [ ] **Implement renderDeckBackground()** - load deck scene asset
  ```ailang
  func renderDeckBackground(deckID: DeckID) -> [DrawCmd] {
    -- DrawCmd to render deck background PNG
    -- Layer 0 (background)
    [Sprite(getDeckSpriteID(deckID), 0.0, 0.0, 0)]
  }
  ```

- [ ] **Implement renderWindows()** - window compositing metadata
  ```ailang
  func renderWindows(
    deckID: DeckID,
    shipVelocity: Vec3,
    shipPosition: Vec3
  ) -> [DrawCmd] {
    let deckInfo = getDeckInfo(deckID);
    match deckInfo.deckType {
      Outward => {
        let regions = getWindowRegions(deckID);
        -- Generate Window3D DrawCmds with velocity/position for starmap
        map(regions, \r. Window3D(r.x, r.y, r.width, r.height, shipVelocity, shipPosition))
      }
      Inward => []
    }
  }
  ```

- [ ] **Implement renderCrewOnDeck()** - crew sprites
  ```ailang
  func renderCrewOnDeck(
    deckID: DeckID,
    interior: InteriorState,
    crew: [Crew]
  ) -> [DrawCmd] {
    let deckState = getDeckState(interior.deckStates, deckID);
    let crewHere = filter(crew, \c. contains(deckState.crewPresent, c.id));

    -- For each crew member, get spawn point and render sprite
    -- Layer 50 (foreground)
    map(crewHere, \c. Sprite(
      getCrewSpriteID(c),
      getCrewX(c, deckID),
      getCrewY(c, deckID),
      50
    ))
  }
  ```

- [ ] **Implement renderDeckUI()** - deck name, navigation buttons
  ```ailang
  func renderDeckUI(deckInfo: DeckInfo) -> [DrawCmd] {
    [
      Text(deckInfo.name, 20.0, 20.0, 100),
      UiButton("Bridge", 20.0, 60.0, 100.0, 40.0, 100, 1),
      UiButton("Obs Deck", 20.0, 110.0, 100.0, 40.0, 100, 2)
    ]
  }
  ```

- [ ] **Update sim/step.ail renderFrame()** to include interior rendering
  ```ailang
  export func renderFrame(world: World) -> FrameOutput {
    ...
    let interiorCmds = renderDeck(
      world.interior,
      world.crew,
      world.ship.velocity,
      world.ship.position
    );

    { drawCmds: concat([exteriorCmds, interiorCmds, uiCmds]) }
  }
  ```

- [ ] **Type-check:** `ailang check sim/*.ail`
- [ ] **Compile:** `make sim`

**Estimated:** 6-8 hours

**AILANG Complexity:** High - list manipulation, DrawCmd generation

**Potential Blocker:** Window3D DrawCmd might need new type in protocol.ail

---

## Day 5: Engine Integration & Window Compositing

**Goal:** Wire AILANG DrawCmds to Go engine, implement window compositing

### Tasks

- [ ] **Add Window3D to protocol.ail** (if needed)
  ```ailang
  export type DrawCmd =
    | ...existing variants...
    | Window3D(float, float, float, float, Vec3, Vec3)  -- x, y, w, h, velocity, position
  ```

- [ ] **Regenerate sim_gen:** `make sim`

- [ ] **Add Window3D case to engine/render/draw.go**
  ```go
  case *sim_gen.DrawCmdWindow3D:
    // Extract window region from DrawCmd
    // Composite starmap rendering (from tetra.Scene) into this region
    // Use window mask + DestinationIn blending (as in prototype)
  ```

- [ ] **Load deck assets** in engine/assets/
  - Add deck background sprites to sprite atlas or load separately
  - Window masks loaded as separate images
  - Crew sprites in atlas

- [ ] **Update engine/render/draw.go** to render deck backgrounds
  - Handle Sprite DrawCmds with deck sprite IDs
  - Scale/position to fit screen (1344x768 → 1280x720)

- [ ] **Test with Go engine:** `make run`
  - Should see Bridge deck render
  - Windows should show starmap
  - Should be able to switch to Observation Deck (if UI working)

**Estimated:** 6-8 hours

**Engine Complexity:** Medium - window compositing already prototyped

**Reference:** demo-engine-scene-bridge/main.go compositeWindows() function

---

## Day 6: Asset Iteration & Polish

**Goal:** Refine deck images, test with live starmap, crew sprites

### Tasks

- [ ] **Test window compositing with real starmap**
  - Verify Saturn/planets visible through windows
  - Check SR/GR effects visible when ship accelerates
  - Adjust window regions if mask coverage wrong

- [ ] **Iterate deck images if needed**
  - If style doesn't match → regenerate with updated prompts
  - Use `bin/voyage ai -generate-image -aspect 16:9 -size 2K -prompt "..."`
  - Regenerate window masks: `go run ./cmd/generate-window-mask`

- [ ] **Add placeholder crew sprites**
  - Simple colored rectangles or basic character sprites
  - Test crew appearing on correct decks based on mood

- [ ] **Test deck navigation**
  - Switch between Bridge and Observation Deck
  - Verify crew locations update
  - Check transitions smooth

- [ ] **Performance check**
  - Rendering at 60 FPS?
  - Window compositing performant?
  - Any visual glitches?

**Estimated:** 4-6 hours

**Asset Work:** Mostly validation, minor tweaks

---

## Day 7: Testing, Documentation & Sprint Wrap-up

**Goal:** Comprehensive testing, document what was learned

### Tasks

- [ ] **Integration testing**
  - Start game → navigate to interior → switch decks
  - Crew appear in right locations
  - Windows show live starmap with SR/GR effects
  - Can return to exterior starmap mode

- [ ] **Screenshot testing**
  - Capture Bridge scene with Saturn visible
  - Capture Observation Deck with nebula visible
  - Verify visual quality matches design intent

- [ ] **Update design doc status**
  - Move `design_docs/planned/scene-based-interior-navigation.md` →
  - `design_docs/implemented/v0_6_0/scene-based-interior-navigation.md`
  - Add "Implemented" date
  - Add "Files Created" section

- [ ] **Document AILANG issues encountered**
  - Any blockers hit?
  - Features that would have helped?
  - Submit feedback via `ailang messages send`

- [ ] **Check AILANG inbox**
  - `ailang messages list --unread`
  - Respond to any questions from AILANG team

- [ ] **Update demos.md**
  - Document how to test interior navigation
  - Add any new demo commands

- [ ] **Git commit**
  - Clear commit message describing what was implemented
  - Reference design doc

**Estimated:** 3-4 hours

---

## Total Estimated Effort

| Day | Tasks | Hours |
|-----|-------|-------|
| 1 | AILANG types | 4-6 |
| 2 | Crew placement logic | 4-6 |
| 3 | Navigation & state | 5-7 |
| 4 | DrawCmd generation | 6-8 |
| 5 | Engine integration | 6-8 |
| 6 | Asset iteration | 4-6 |
| 7 | Testing & docs | 3-4 |
| **Total** | | **32-45 hours** → **5-7 days** |

---

## AILANG Complexity Assessment

| Factor | Impact | Notes |
|--------|--------|-------|
| Module imports | Low | Existing modules work well |
| List manipulation | Medium | Crew assignment, DrawCmd building |
| Record updates | Medium | Updating interior state |
| Pattern matching | Low | DeckID enum is simple |
| Recursion | Low | No deep recursion needed |
| **Overall** | **Medium** | No major blockers expected |

---

## Risks & Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Window3D DrawCmd not in protocol | Medium | High | Add it (Day 5) or use existing Window type |
| List updates verbose | Medium | Low | Create helper functions |
| Window compositing performance | Low | Medium | Already validated in prototype |
| Asset iteration slow | Low | Low | Prompts documented, can regenerate quickly |
| Codegen bug (unread message) | Low | Medium | Check message, workaround if needed |

---

## Files to Create/Modify

### New Files
- `sim/interior.ail` (REWRITE - scene-based, not 3D)
- `sprints/scene-based-interior-navigation-sprint.md` (this file)
- `design_docs/implemented/v0_6_0/scene-based-interior-navigation.md` (moved from planned/)

### Modified Files
- `sim/world.ail` - add `interior: InteriorState` field
- `sim/step.ail` - update step(), integrate deck navigation
- `sim/protocol.ail` - potentially add Window3D DrawCmd
- `engine/render/draw.go` - add Window3D rendering case
- `engine/assets/` - load deck backgrounds, window masks
- `design_docs/reference/demos.md` - document new features

### Deprecated Files (backed up)
- `sim/interior.ail` → `sim/interior_3d_DEPRECATED.ail`
- `sim/interior_themes.ail` → `sim/interior_themes_DEPRECATED.ail`
- `sim/window.ail` → `sim/window_DEPRECATED.ail`

---

## AILANG Feedback Checkpoints

### Pre-Sprint
- [x] Check inbox: 1 unread message (codegen non-exported functions)
- [ ] Acknowledge message: `ailang messages ack <msg-id>`

### Mid-Sprint (Day 3-4)
- [ ] Report any blockers immediately via `ailang messages send`
- [ ] Check inbox for responses

### Post-Sprint (Day 7)
- [ ] Send summary of experience
- [ ] Report: bugs, feature requests, documentation gaps
- [ ] Check inbox for any follow-ups

---

## Success Validation

Sprint is **COMPLETE** when:
- ✅ Player can navigate between Bridge and Observation Deck
- ✅ Deck scenes render with AI-generated backgrounds
- ✅ Windows composite live starmap (planets, SR/GR effects visible)
- ✅ Crew sprites appear on decks (even if placeholder graphics)
- ✅ No visual glitches, rendering at 60 FPS
- ✅ Design doc moved to implemented/
- ✅ AILANG feedback sent

---

## Next Steps After Sprint

1. **Add inward-facing decks** (Engineering, Medical, Quarters)
2. **Real crew sprites** with animations
3. **Conversation system** when clicking crew
4. **Portrait pop-ups** for crew emotions
5. **Deck discovery/unlocking** as ship expands

---

**Sprint Created:** 2025-12-20
**Sprint Planner:** game-vision-designer + sprint-planner
**Ready for Execution:** ✅ YES
