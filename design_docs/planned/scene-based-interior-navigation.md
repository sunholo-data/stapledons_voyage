# Scene-Based Interior Navigation System

**Status**: Planned
**Target**: v0.6.0
**Priority**: P0 - Critical (core gameplay experience)
**Estimated**: 5-7 days
**Dependencies**: None (clean break from 3D approach)

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Choices Are Final | N/A | 0 | Navigation infrastructure |
| The Game Doesn't Judge | N/A | 0 | Presentation layer |
| Time Has Emotional Weight | ++ | +2 | Visual contrast cosmic/human scale reinforces isolation and intimacy |
| The Ship Is Home | ++ | +2 | Crew in environments, intimate conversations, cosmic vs human contrast |
| Grounded Strangeness | N/A | 0 | Infrastructure (but enables grounded crew interactions) |
| We Are Not Built For This | + | +1 | Inward decks show claustrophobia, psychological strain |
| **Net Score** | | **+5** | **Decision: Move forward** |

**Feature type:** Core Gameplay
- Primary interface for crew interactions and conversations
- Visual grounding for human-scale experience vs cosmic scale
- Enables all interior ship gameplay

**Reference:** See [design decision 2025-12-20](../../docs/vision/design-decisions.md#2025-12-20-interior-ship-experience-scene-based-navigation-with-outwardinward-deck-types)

## Problem Statement

**Previous Attempts (REJECTED):**
- **Isometric tiles:** AI asset generation couldn't match precise tile dimensions consistently
- **3D first-person:** Too complex for asset creation, pulls focus from conversations to spatial puzzles
- Both approaches misaligned with game focus: space visualization and SR/GR psychology through conversations

**Current State:**
- Exterior space visualization (starmaps, star fields, SR/GR effects) works beautifully
- No interior ship experience implemented
- Game's core strength is conversations about relativistic physics, not spatial navigation

**Impact:**
- Cannot access crew for conversations (Pillar 4: The Ship Is Home)
- Missing visual contrast between cosmic void and human intimacy (Pillar 3: Time Has Emotional Weight)
- Cannot show psychological strain in enclosed spaces (Pillar 6: We Are Not Built For This)

## Goals

**Primary Goal:** Implement scene-based interior navigation using AI-generated 2D/2.5D images with parallax layering, focusing on conversations and visual contrast.

**Success Metrics:**
- Player can navigate to different ship decks via UI selection
- Each deck has distinct visual identity and atmosphere
- Outward-facing decks show live starmap through windows (SR/GR effects visible)
- Inward-facing decks create claustrophobic, introspective feeling
- Crew appear in contextually appropriate locations based on mood/needs
- Seamless transition between starmap (exterior) and deck scenes (interior)
- AI-generated deck images can be created/iterated quickly (hours not days)

**Non-Goals:**
- Free spatial movement (WASD navigation) - NOT NEEDED
- 3D asset creation - REJECTED as too complex
- Spatial puzzles or exploration gameplay - Game focuses on conversations, not navigation

## Solution Design

### Overview

**Core Concept:** Player selects deck locations from ship UI. Each deck loads a scene composed of AI-generated 2D images with parallax layering. Outward-facing decks composite live starmap rendering in window regions. Crew placement is contextual based on mood, time, and events.

**Visual System:**

```
Deck Scene Composition:
  - Background layer: Deep background (walls, structural elements)
  - Midground layer: Functional elements (equipment, furniture, windows)
  - Foreground layer: Interactive elements (crew sprites, UI hotspots)
  - [For outward decks] Window compositing: Live starmap rendering in window regions
  - Parallax: Subtle depth effect when camera/view shifts
```

**Deck Types:**

**Type A: Outward-Facing Decks** (connected to cosmos)
- Bridge, Observation Deck, Navigation
- **Visual:** Windows showing space, live starmap composited
- **SR/GR effects:** Visible through windows during cruise
- **Atmosphere:** Awe, scale, connection to universe
- **Purpose:** Command decisions, stargazing, processing cosmic events

**Type B: Inward-Facing Decks** (introspective human scale)
- Engineering, Medical Bay, Crew Quarters, Archive, Hydroponics
- **Visual:** No windows, fully enclosed
- **Atmosphere:** Claustrophobic, intimate, human-scale
- **Purpose:** Personal conversations, system management, rest, psychological strain

### Architecture

**Layer Breakdown:**

```
Game State (AILANG):
  currentDeck: DeckID
  crewLocations: Map<CrewID, DeckID>
  deckStates: Map<DeckID, DeckState>
  shipVelocity: float (for SR/GR effects in windows)
  shipPosition: Vec3 (galactic coords)

Rendering Pipeline (Go Engine):
  1. Load deck scene layers (AI-generated PNGs)
  2. Apply parallax offset based on view
  3. [Outward decks only] Composite live starmap in window regions
  4. Render crew sprites at locations
  5. Render UI overlays (deck status, navigation)

Navigation Flow:
  User clicks deck in ship UI
    → AILANG updates currentDeck
    → Engine unloads old scene
    → Engine loads new deck scene
    → AILANG determines which crew are present
    → Engine renders crew in scene
```

**Data Flow:**

```
AILANG (sim/*.ail):
  - Tracks current deck
  - Determines crew placement (mood/need-based)
  - Generates DrawCmds for deck scene + crew
  - Outputs ship velocity for SR/GR effects

Go Engine (engine/):
  - Loads deck scene assets (layered PNGs)
  - Renders DrawCmds (sprites, UI, text)
  - [Outward decks] Composites starmap in windows
  - Handles deck selection input
```

### Initial Deck Locations (Expandable)

**v0.6.0 Implementation: Bridge + Observation Deck ONLY** (outward-facing decks demonstrating window compositing)

The bubble ship is enormous - starting with 2 decks for first implementation, more can be added:

| Deck | Type | Priority | Function | Atmosphere | Crew Activities |
|------|------|----------|----------|------------|-----------------|
| **Bridge** | Outward | **P0 (v0.6.0)** | Command, navigation decisions | Professional, wide panoramic windows | Captain makes journey decisions, Pilot at controls |
| **Observation Deck** | Outward | **P0 (v0.6.0)** | Stargazing, reflection | Contemplative, massive curved windows | Crew process cosmic events, philosophical conversations |
| **Engineering** | Inward | P1 (future) | System management, crisis response | Industrial, claustrophobic | Engineer maintains systems, crisis management |
| **Medical Bay** | Inward | P1 (future) | Health, psychological care | Clinical, sterile | Medic treats injuries, psychological counseling |
| **Crew Quarters** | Inward | P1 (future) | Personal space, rest | Intimate, personalized | Private conversations, rest, crew homes |
| **Archive/AI Core** | TBD | P2 (future) | Data, civilization records | Depends on window decision | Access AI, review civ data, research |
| **Hydroponics** | Inward | P2 (future) | Life support, sustainability | Organic, alive | Food production, trace synthesis, growth |

**Future additions:** Cargo bay, workshop, recreation, chapel/meditation, docking port

### Crew Placement Logic

**Crew have favorite locations based on:**
- **Time of day/shift** - Engineer in Engineering during work, Quarters during rest
- **Mood/state** - Stressed crew seek private spaces, social crew seek common areas
- **Needs** - Hungry → Hydroponics, Injured → Medical, Contemplative → Observation
- **Events** - Crisis → relevant deck (Engineering for system failure)
- **Relationships** - Crew with bonds may appear together, conflicts avoid each other

**AI-Assisted Placement:**
- AILANG can use AI effect to suggest contextually appropriate crew locations
- Example: "Where would the Medic be after witnessing a civilization collapse?"
  → AI suggests: Medical Bay (treating stressed crew) or Observation Deck (processing trauma)

**Implementation:**
```ailang
-- In AILANG sim/*.ail
func getCrewLocation(crew: Crew, time: float, recentEvents: [Event]) -> DeckID {
  -- Base on archetype, OCEAN traits, mood, needs
  -- Can use AI.call() for contextual suggestions
  match crew.mood {
    Stressed => Quarters | MedicalBay
    Contemplative => ObservationDeck
    Focused => crew.workDeck
    Social => commonAreas
    ...
  }
}
```

### Crew Visual Representation

**Decision (2025-12-20):** Sprite overlay system with emotion-based portraits

**In-Deck Rendering:**
- Crew appear as **sprite overlays** positioned on deck scenes
- Sprites can move and animate (walk cycles, idle animations)
- Separate from deck background → can be easily positioned/animated

**Conversation UI:**
- **Portrait pop-ups** when talking to crew
- Multiple portrait variations per crew member based on emotion/mood
- Example: Captain has portraits for: neutral, stressed, contemplative, angry, hopeful

**Technical:**
```
Deck rendering layers (bottom to top):
  1. Deck background scene (AI-generated, static)
  2. Crew sprites (positioned, animated)
  3. UI overlays (deck status, navigation)
  4. Conversation portraits (when talking to crew)
```

### Asset Pipeline

**Art Style:** Painterly/stylized like 70s sci-fi comics (NOT photorealistic)

**Target Dimensions:**
- **Deck scenes:** 1280x720 (16:9, matches game resolution)
- **Crew sprites:** ~128x256 (character height ~1/3 of screen)
- **Portraits:** 256x256 (square, for UI panels)

**Phase 1: Initial Deck Scene Generation (AI)**

For each deck:
1. **Generate base image** - Use `bin/voyage ai -generate-image -aspect 16:9 -size 2K -prompt "..."`

   **CRITICAL: For outward decks, windows MUST be pure black or very dark (RGB < 32)**

   **Working prompts (tested 2025-12-20):**
   - Bridge: "Retro 1970s sci-fi comic book illustration style spaceship bridge interior, wide panoramic curved window showing **empty black space**, command consoles with glowing cyan displays, captain's chair in center, moody atmospheric lighting with orange accent colors, painterly brush strokes, dramatic perspective, **no people, no stars in windows yet**, cinematic composition"

   - Observation Deck: "Retro 1970s sci-fi comic book art style spaceship observation deck, massive curved panoramic floor-to-ceiling windows showing **empty black space**, comfortable retro seating areas with green and orange chairs for contemplation, warm atmospheric lighting with orange and yellow tones, wooden or metallic accents, intimate contemplative lounge space, painterly illustration style, dramatic moody lighting, **no people, no stars in windows yet**"

   - Engineering (future): "Retro sci-fi comic book industrial engineering bay, no windows, pipes and machinery, dim moody lighting, claustrophobic, painterly illustration, no people"

2. **Window mask creation** (for outward decks):
   - Identify black window regions (RGB < 32) in deck scene
   - Create PNG mask: white pixels = windows, transparent = solid walls
   - Wide panoramic windows (not narrow splits)
   - Engine composites live starmap into white mask regions

   **Why black windows:** AI-generated windows must be pure black so masking logic can identify them. The prompt phrases "empty black space" and "no stars in windows yet" prevent AI from filling windows with nebulas/stars.

3. **Iteration:**
   - Quick regeneration if style/composition doesn't match
   - Swap images easily (same dimensions, update manifest)
   - Consistent 70s comic aesthetic across all decks
   - Each deck has distinct color palette/atmosphere (later iteration)

**Phase 2: Iteration & Refinement**
- Quick AI regeneration if scene doesn't match vision
- Add variations (damaged, upgraded, different lighting)
- Create transition frames for animations (optional polish)

**Asset Organization:**
```
assets/decks/
  bridge/
    background.png
    midground.png
    foreground.png
    window_mask.png (alpha mask for starmap compositing)
    manifest.json (layer metadata)
  observation/
    background.png
    ...
  engineering/
    background.png
    ...
```

**Manifest Format:**
```json
{
  "deckID": "bridge",
  "deckType": "outward",
  "layers": [
    {"id": "background", "file": "background.png", "parallax": 0.0},
    {"id": "midground", "file": "midground.png", "parallax": 0.3},
    {"id": "foreground", "file": "foreground.png", "parallax": 0.6}
  ],
  "windows": {
    "enabled": true,
    "maskFile": "window_mask.png",
    "starfieldDepth": 500.0
  },
  "crewSpawnPoints": [
    {"x": 512, "y": 768, "role": "captain"},
    {"x": 768, "y": 800, "role": "pilot"}
  ]
}
```

### Window Compositing System (Outward Decks Only)

**Goal:** Blend AI-generated interior with live starmap rendering in window regions.

**Implementation:**
```go
// In engine/render/deck_renderer.go

type DeckRenderer struct {
    layers        []*ebiten.Image
    windowMask    *ebiten.Image
    starmapBuffer *ebiten.Image
}

func (dr *DeckRenderer) Render(screen *ebiten.Image, shipVelocity float64, shipPos Vec3) {
    // 1. Render background layer
    screen.DrawImage(dr.layers[0], nil)

    // 2. [Outward decks only] Composite starmap in window regions
    if dr.windowMask != nil {
        // Render starmap to buffer with SR/GR effects
        renderStarmap(dr.starmapBuffer, shipVelocity, shipPos)

        // Composite using window mask
        op := &ebiten.DrawImageOptions{}
        op.CompositeMode = ebiten.CompositeModeSourceIn
        dr.starmapBuffer.DrawImage(dr.windowMask, op)
        screen.DrawImage(dr.starmapBuffer, nil)
    }

    // 3. Render midground and foreground layers
    screen.DrawImage(dr.layers[1], nil)
    screen.DrawImage(dr.layers[2], nil)

    // 4. Render crew sprites (from AILANG DrawCmds)
    // 5. Render UI overlays
}
```

**SR/GR Effects in Windows:**
- During cruise (v > 0.2c): Full SR aberration, Doppler shift, beaming
- Docked/orbital (v ≈ 0): Normal starfield rendering
- Transition: Smooth fade between modes

### Navigation UI

**Deck Selection Interface:**

**Option A: Ship Schematic View**
- Top-down or side cutaway view of bubble ship
- Click deck to navigate
- Visual: Shows ship structure, deck locations
- Pro: Spatial context, ship layout understanding
- Con: Needs ship schematic artwork

**Option B: Deck List/Menu**
- Simple list of available decks with icons
- Click to navigate
- Visual: Clean UI, deck icons + status
- Pro: Fast to implement, no artwork needed
- Con: Less immersive, no spatial sense

**Recommendation:** Start with Option B (fast), upgrade to Option A later (polish).

**Deck Selection UI Mockup (Option B):**
```
┌─────────────────────────────────┐
│ SHIP DECKS                      │
├─────────────────────────────────┤
│ ⊙ Bridge            [2 crew]    │
│ ◉ Observation Deck  [4 crew]    │  ← Current location
│ ⚙ Engineering       [1 crew]    │
│ + Medical Bay       [1 crew]    │
│ ⌂ Crew Quarters     [3 crew]    │
│ ◘ Archive           [0 crew]    │
│ ❀ Hydroponics       [1 crew]    │
└─────────────────────────────────┘
```

**Transition Animation (Optional Polish):**
- Fade out current deck → Fade in new deck
- Or: Slide/wipe transition
- Or: No animation (instant switch) - simplest

## Implementation Plan

### Phase 1: Core Architecture (2 days)

**AILANG (`sim/interior.ail`):**
- [ ] Define DeckID enum (Bridge, Observation, Engineering, Medical, Quarters, Archive, Hydroponics)
- [ ] Define DeckType (Outward, Inward)
- [ ] Define DeckState type (crew present, status, atmosphere)
- [ ] Define InteriorState type (currentDeck, deckStates map)
- [ ] Implement `init_interior() -> InteriorState`
- [ ] Implement `step_interior(state, input) -> InteriorState`
- [ ] Implement `render_interior(state) -> [DrawCmd]`
  - Generate deck scene DrawCmds
  - Place crew sprites based on location logic
  - Generate navigation UI

**Go Engine (`engine/deck/`):**
- [ ] Create `deck_renderer.go` - DeckRenderer type, layer loading, parallax
- [ ] Create `deck_loader.go` - Load deck manifests, parse JSON
- [ ] Create `window_compositor.go` - Starmap compositing for outward decks
- [ ] Update `engine/render/draw.go` - Handle new DrawCmds if needed

**Integration (`cmd/game/` or new demo):**
- [ ] Wire AILANG interior state into game loop
- [ ] Handle deck selection input → pass to AILANG
- [ ] Render AILANG DrawCmds via DeckRenderer

### Phase 2: Asset Generation (1-2 days)

**For each initial deck (7 decks):**
- [ ] Generate base AI image (2048x1536 recommended)
- [ ] Separate into layers (background, midground, foreground)
- [ ] [Outward decks] Create window mask
- [ ] Create manifest.json with metadata
- [ ] Add to `assets/decks/` folder
- [ ] Test loading in engine

**Tooling (optional):**
- [ ] Script to batch-generate AI prompts for decks
- [ ] Tool to preview deck scenes with layers
- [ ] Manifest validator

### Phase 3: Crew Placement Logic (1-2 days)

**AILANG (`sim/crew_placement.ail`):**
- [ ] Implement `getCrewLocation(crew, time, events) -> DeckID`
  - Base logic on archetype, OCEAN, mood
  - Handle time-of-day (work shift vs rest)
  - Handle events (crisis → relevant deck)
- [ ] Implement `updateCrewLocations(state) -> state`
  - Called each step to update crew positions
  - Can use AI effect for contextual suggestions
- [ ] Add crew movement transitions (optional: crew can move between decks)

**Testing:**
- [ ] Verify crew appear in expected locations
- [ ] Test mood-based placement
- [ ] Test event-driven placement (crisis scenarios)

### Phase 4: Window Compositing (1 day)

**Go Engine (`engine/deck/window_compositor.go`):**
- [ ] Implement starmap rendering to buffer
- [ ] Implement mask-based compositing
- [ ] Integrate SR/GR effects (use existing space_view.go)
- [ ] Handle velocity transitions (docked ↔ cruise)

**Testing:**
- [ ] Verify starmap visible through windows on outward decks
- [ ] Test SR effects at high velocity (aberration, Doppler)
- [ ] Verify inward decks have no windows

### Phase 5: Navigation UI (1 day)

**AILANG (`sim/interior_ui.ail`):**
- [ ] Generate deck list UI DrawCmds
- [ ] Handle deck selection input
- [ ] Show crew count per deck
- [ ] Highlight current deck

**Go Engine:**
- [ ] Render UI DrawCmds
- [ ] Handle click detection on deck list
- [ ] Pass selection to AILANG

**Polish (optional):**
- [ ] Transition animation between decks
- [ ] Deck icons/status indicators
- [ ] Sound effects for deck transitions

### Phase 6: Testing & Polish (1 day)

**Manual Testing:**
- [ ] Navigate to all decks, verify scenes load
- [ ] Check crew appear in contextual locations
- [ ] Verify outward decks show starmap in windows
- [ ] Test SR/GR effects during cruise
- [ ] Verify inward decks feel claustrophobic (no windows)
- [ ] Test deck selection UI responsiveness

**Polish:**
- [ ] Add ambient sounds per deck type
- [ ] Lighting adjustments for atmosphere
- [ ] Crew idle animations (if time permits)
- [ ] Deck status overlays (system health, etc.)

## Files to Create/Modify

**New Files (AILANG):**
- `sim/interior.ail` - Core interior navigation state and logic
- `sim/crew_placement.ail` - Crew location logic
- `sim/interior_ui.ail` - Navigation UI generation

**New Files (Go Engine):**
- `engine/deck/deck_renderer.go` - Deck scene rendering
- `engine/deck/deck_loader.go` - Manifest parsing, asset loading
- `engine/deck/window_compositor.go` - Starmap compositing for windows
- `engine/deck/types.go` - Go types for deck data

**Modified Files:**
- `engine/render/draw.go` - Handle any new DrawCmd types if needed
- `cmd/game/main.go` - Integrate interior navigation into game loop
- `engine/assets/manifest.go` - Support deck manifests (if not already generic)

**New Assets:**
- `assets/decks/{bridge,observation,engineering,medical,quarters,archive,hydroponics}/`
  - Each folder: background.png, midground.png, foreground.png, manifest.json
  - Outward decks: +window_mask.png

## Success Criteria

**Must Have:**
- [ ] Player can navigate to all 7 initial decks via UI
- [ ] Each deck has distinct visual identity matching its function
- [ ] Outward decks (Bridge, Observation) show live starmap through windows
- [ ] Inward decks (Engineering, Medical, Quarters, Archive, Hydroponics) have no windows
- [ ] SR/GR effects visible in windows during cruise (v > 0.2c)
- [ ] Crew appear in contextually appropriate locations (work decks during shift, quarters during rest)
- [ ] Smooth transition between starmap (exterior) and deck scenes (interior)
- [ ] No 3D navigation or spatial puzzles (scene-based only)

**Should Have:**
- [ ] Parallax effect gives subtle depth to deck scenes
- [ ] Crew count visible in deck selection UI
- [ ] Deck status indicators (system health, crisis alerts)
- [ ] Transition animation between decks (fade or wipe)

**Could Have:**
- [ ] Multiple variations per deck (damaged, upgraded, different times of day)
- [ ] Crew idle animations in scenes
- [ ] Ambient sounds per deck type
- [ ] Ship schematic view for deck selection (upgrade from list UI)

## Testing Strategy

**Unit Tests (AILANG):**
- Test `getCrewLocation()` with various crew states
- Test `updateCrewLocations()` with events
- Test deck state transitions

**Integration Tests:**
- Not applicable (primarily visual/rendering system)

**Manual Testing:**
- Navigate to each deck, verify visuals
- Check crew placement logic with different scenarios
- Test window compositing with velocity changes
- Verify UI responsiveness

**Visual Regression:**
- Capture screenshots of each deck
- Compare after asset updates

## Non-Goals

**Not in this feature:**
- 3D navigation or spatial movement (REJECTED)
- Isometric tile-based rendering (REJECTED)
- Complex interior puzzles or exploration gameplay
- Multiplayer or networked crew
- VR support
- Procedurally generated deck layouts (future consideration)

## Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|-----------|
| AI-generated deck images inconsistent style | Medium | Medium | Establish style guide, use consistent prompts, iterate quickly |
| Window compositing performance issues | Low | Low | Pre-render starmap to buffer, use GPU compositing |
| Crew placement feels random/wrong | Medium | Medium | Extensive testing, AI-assisted suggestions, user feedback |
| Deck scenes feel static/lifeless | Medium | Low | Add subtle animations (crew idle, equipment blinks), ambient sounds |
| Players expect spatial exploration | Low | Medium | Clear framing in game intro, focus on conversation UI |

## Future Enhancements

**Post-MVP:**
- **Dynamic deck states:** Damaged, upgraded, different lighting/time of day
- **Crew animations:** Idle, working, talking to each other
- **Interactive objects:** Click equipment for info, inspect crew workstations
- **Ship schematic UI:** Visual ship layout for deck selection
- **Deck unlocks:** Discover/unlock new areas through gameplay
- **Multiple rooms per deck:** Expand large decks into sub-areas
- **Custom deck backgrounds:** Per-playthrough variations, player ship customization
- **Procedural details:** AI-assisted population of deck details based on game state

**Long-term:**
- **Modding support:** Community-created decks with custom imagery
- **Dynamic window views:** Show nearby planets, stations, not just starmaps
- **Crew movement animations:** See crew walk between decks
- **Camera rotation/zoom:** Subtle parallax camera control within scenes

## References

- [Design Decision: Interior Ship Experience (2025-12-20)](../../docs/vision/design-decisions.md#2025-12-20-interior-ship-experience-scene-based-navigation-with-outwardinward-deck-types)
- [Core Pillars](../../docs/vision/core-pillars.md) - Pillar 3 (Time Has Emotional Weight), Pillar 4 (The Ship Is Home), Pillar 6 (We Are Not Built For This)
- [Rejected: Bubble Ship Dome System](../rejected/bubble-ship-dome-system.md) - 3D approach, see why rejected
- [engine-capabilities.md](../reference/engine-capabilities.md) - Existing engine DrawCmds and effects
- [game-capabilities.md](../reference/game-capabilities.md) - AILANG features available

## Related Features

**Depends on this:**
- Crew conversation system (needs deck scenes as backdrop)
- Crisis management UI (needs deck-specific visuals)
- Ship upgrade visualization (decks change appearance)

**This depends on:**
- None (clean break from previous approaches)

---

**Document created**: 2025-12-20
**Last updated**: 2025-12-20
