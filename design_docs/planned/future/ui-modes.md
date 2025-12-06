# UI Modes Architecture

**Version:** 0.5.0
**Status:** Planned
**Priority:** P0 (Core Game Architecture)
**Complexity:** High
**AILANG Workarounds:** WorldMode enum for state machine, record-based UI state
**Depends On:** v0.3.0 Camera/Viewport, v0.4.0 NPC Movement

## Related Documents

- [Architecture Overview](../v0_1_0/architecture.md) - Data flow context
- [Camera & Viewport](../v0_3_0/camera-viewport.md) - Rendering system
- [NPC Movement](../v0_4_0/npc-movement.md) - Entity rendering
- [Game Vision](../../../docs/game-vision.md) - Core gameplay loop

## Problem Statement

The game needs multiple distinct UI "surfaces" or modes to support its gameplay:

- Exploring the ship between journeys
- Navigating the galaxy map to plan routes
- Interacting with civilizations
- Managing the crew and trade
- Viewing the endgame legacy report

**Current State:**
- Single rendering mode (tilemap + NPCs)
- No mode switching logic
- No layered UI system
- No dialogue/panel rendering

**What's Needed:**
- WorldMode enum to track current UI mode
- Mode-specific rendering and input handling
- Transitions between modes
- Layered drawing (world → UI panels → dialogs)

---

## Design Overview

### UI Mode Architecture

The game operates in one of several **UI Modes** at any time. Each mode:
1. Has its own input handling logic
2. Renders its own set of DrawCmds
3. Can transition to other modes via player actions or events

```
┌─────────────────────────────────────────────────────────────────┐
│                         WorldMode                                │
├─────────────────────────────────────────────────────────────────┤
│  ShipExploration ←→ Dialogue ←→ CrewManagement                  │
│         ↓                              ↓                         │
│    GalaxyMap ←→ JourneyPlanning → JourneyEvents                 │
│         ↓                              ↓                         │
│  CivilizationDetail ←→ Trade    PlanetSurface ← Ruins           │
│                                        ↓                         │
│                              EndgameLegacy                       │
└─────────────────────────────────────────────────────────────────┘
```

### Engine Rendering Layers

| Z-Layer | Content |
|---------|---------|
| 0-9 | World tiles (terrain, structures) |
| 10-19 | World entities (NPCs, player, ships) |
| 20-29 | Selection highlights, world UI |
| 30-39 | Panels (sidebars, status bars) |
| 40-49 | Overlays (full-screen UIs like galaxy map) |
| 50-59 | Modals (dialogs, trade UI) |
| 60+ | System UI (pause menu, notifications) |

---

## UI Modes Specification

### 1. Ship Exploration Mode

**Purpose:** The "home space" where players explore their ship, interact with crew, manage systems.

**What's On Screen:**
- Isometric or top-down ship interior tileset
- Animated crew NPC sprites
- Room labels and system indicators
- Hover UI showing crew stats/moods

**Interactions:**
- Click-to-move or WASD movement
- Click crew member → transition to Dialogue mode
- Click ship system → system management panel
- Press M → transition to Galaxy Map

**Key Gameplay:**
- Crew relationships develop here
- Ship upgrades require visiting specific rooms
- Events (breakdowns, births, debates) trigger here

```ailang
type ShipExplorationState = {
    playerPos: Coord,
    hoveredEntity: Maybe(EntityID),
    currentDeck: int,
    roomHighlights: [RoomID]
}
```

### 2. Full-Screen Dialogue Mode

**Purpose:** Rich character interactions for crew conversations, first contacts, events.

**Layout:**
- Large NPC portrait (left 40%)
- Dialogue text area (right top)
- Player choice buttons (right bottom)
- Relationship/morale indicators (corner)

**Interactions:**
- Click choice → advance dialogue
- Press Escape → return to previous mode
- Hover choice → show predicted outcome hints

**AI Integration:**
- NPC portraits can be AI-generated
- Dialogue text can be LLM-powered
- Relationship/outcome predictions calculated by simulation

```ailang
type DialogueState = {
    speakerID: EntityID,
    portrait: SpriteID,
    currentText: string,
    choices: [DialogueChoice],
    emotionState: Emotion,
    relationshipDelta: int
}

type DialogueChoice = {
    text: string,
    outcome: DialogueOutcome,
    available: bool,
    tooltip: Maybe(string)
}
```

### 3. Galaxy Map Mode

**Purpose:** Strategic navigation layer for planning journeys across the galaxy.

**What's On Screen:**
- Starfield background (parallax layers)
- Star nodes (color-coded by civilization status)
- Network edges (contact graph)
- Current position marker
- Time dilation calculator sidebar

**Interactions:**
- Pan/zoom with mouse or keyboard
- Click star → open Civilization Detail panel
- Right-click star → open Journey Planning UI
- Shift-click → compare two destinations
- Scroll → zoom in/out

**Key Displays:**
- Civilization states: Thriving (green), Declining (yellow), Extinct (gray), Unknown (blue)
- Contact network connections
- Distance and estimated journey times
- Last known state/last visit year

```ailang
type GalaxyMapState = {
    cameraPos: Vec2,
    zoomLevel: float,
    selectedStar: Maybe(StarID),
    hoveredStar: Maybe(StarID),
    filterMode: MapFilter,
    showNetwork: bool,
    timeDilationPreview: Maybe(JourneyPreview)
}

type MapFilter = AllStars | Visited | Unvisited | HasCivilization | Extinct
```

### 4. Planet Surface Mode

**Purpose:** Ground-level exploration for landings, first contacts, archaeology.

**What's On Screen:**
- Isometric or orthographic planetary tileset
- Landing zone environment
- Local NPCs or ruins
- Resource/artifact indicators

**Interactions:**
- Move around landing area
- Interact with locals or objects
- Gather resources/artifacts
- Return to ship (transition out)

**Visual Variation:**
- Thriving civilization → populated, active
- Declining civilization → sparse, worn
- Ruins → broken structures, artifacts

```ailang
type PlanetSurfaceState = {
    playerPos: Coord,
    landingZone: ZoneID,
    visibleEntities: [Entity],
    discoveredArtifacts: [ArtifactID],
    civilizationState: CivState
}
```

### 5. Civilization Detail Screen

**Purpose:** Full information about a civilization when arriving or reviewing known civs.

**Sections:**
- **Overview:** Population, energy, stability, openness
- **Philosophy:** Core question, traits, modifiers
- **Timeline:** Major events since last visit
- **Relationships:** Allies, conflicts, contact history
- **Trade:** Available exchanges

**Interactions:**
- Tab between sections
- Click Trade → open Trade UI
- Click Contact → initiate dialogue
- Press Escape → return to Galaxy Map

```ailang
type CivDetailState = {
    civID: CivilizationID,
    activeTab: CivTab,
    timelineScroll: int,
    tradePreview: Maybe(TradeOffer)
}

type CivTab = Overview | Philosophy | Timeline | Relationships | Trade
```

### 6. Trade & Exchange UI

**Purpose:** Structured trading interface for technology/knowledge exchange.

**Layout:**
- Two columns: Offer (left) and Request (right)
- Draggable items between sides
- Impact preview panel
- Accept/Cancel buttons

**Interactions:**
- Drag items to offer/request
- See predicted impact (stability, philosophy shift)
- See acceptance probability
- Confirm or cancel trade

**Risk Warnings:**
- Giving advanced tech to unstable civ → warning
- Philosophy-altering trade → confirmation required

```ailang
type TradeState = {
    civID: CivilizationID,
    offering: [TradeItem],
    requesting: [TradeItem],
    acceptProbability: float,
    impactPreview: TradeImpact,
    warnings: [string]
}

type TradeItem = Technology(TechID) | Knowledge(KnowledgeID) | Artifact(ArtifactID) | Philosophy(PhilosophyID)
```

### 7. Journey Planning UI

**Purpose:** The core identity of the game - planning irreversible relativistic journeys.

**Display:**
- Current year (objective and subjective)
- Selected destination
- Velocity slider (0.9c to 0.999999c)
- Travel time (subjective years)
- External time elapsed (objective years)
- Crew age projections (births, deaths during transit)
- Fuel/tech constraints
- Crew vote display

**Key Element: The Commit Button**
- Irreversible once pressed
- Major emotional beat
- Shows "Are you sure?" confirmation

```ailang
type JourneyPlanState = {
    destination: StarID,
    velocity: float,           -- 0.9 to 0.999999
    subjectiveTime: float,     -- years you experience
    objectiveTime: float,      -- years that pass outside
    crewProjections: CrewProjection,
    fuelCost: float,
    crewVote: VoteResult,
    committed: bool
}

type CrewProjection = {
    expectedBirths: int,
    expectedDeaths: int,
    crewAges: [(CrewID, int, int)]  -- id, current age, arrival age
}
```

### 8. Journey Events Mode

**Purpose:** Narrative layer during transit, where crew life unfolds.

**Hybrid Mode:**
- Ship interior visible (reduced)
- Event popups overlay
- System status sidebar
- Voyage log entries

**What Happens:**
- Crew ages
- Relationships develop
- Random events fire (breakdowns, debates, births, deaths)
- Voyage logs generated

**Interactions:**
- Navigate ship during transit
- Respond to events (dialogue mode)
- Check system status
- View voyage log

```ailang
type JourneyEventState = {
    currentEvent: Maybe(Event),
    yearsElapsed: float,
    yearsRemaining: float,
    voyageLog: [LogEntry],
    systemStatus: [SystemState],
    activeCrewEvents: [CrewEvent]
}

type Event =
    | SystemFailure(SystemID, Severity)
    | CrewBirth(CrewID, CrewID)
    | CrewDeath(CrewID, DeathCause)
    | PhilosophicalDebate(Topic, [CrewID])
    | RelationshipChange(CrewID, CrewID, RelationType)
    | Discovery(DiscoveryType)
```

### 9. Ancient Ruins / Archaeology Mode

**Purpose:** Exploration of extinct civilizations, environmental storytelling.

**What's On Screen:**
- Isometric ruins tileset
- Broken structures, artifacts
- Data terminals, murals, inscriptions
- Recovery progress indicators

**Interactions:**
- Explore ruins area
- Interact with artifacts
- Access data recovery interface
- Read translated logs/messages

**Thematic Purpose:**
- Shows consequences of civilizations you didn't save
- Delivers backstory from millennia ago
- High emotional weight, minimal art needed if stylized

```ailang
type RuinsState = {
    siteID: RuinsSiteID,
    playerPos: Coord,
    discoveredArtifacts: [Artifact],
    decipheredLogs: [AncientLog],
    explorationProgress: float
}

type AncientLog = {
    originalDate: int,      -- year when written
    civilization: string,
    content: string,
    emotionalTone: Emotion
}
```

### 10. Endgame Legacy Visualization

**Purpose:** The final "score" screen - narrative-first, showing your impact on the galaxy.

**Display Elements:**
- Galaxy network: before vs. after your journey
- Civilization timelines (life → death → descendants)
- Philosophy diversity graph
- Ship generational lineage
- Counterfactual simulations ("What if...?")
- AI-generated epilogue text

**Sections:**
- **Network Impact:** Graph visualization of changes
- **Civilization Fates:** Who lived, who died, why
- **Philosophy Evolution:** What ideas spread or died
- **Your Legacy:** Lineage, founded institutions, lasting influence
- **Counterfactuals:** What would have happened differently

```ailang
type LegacyState = {
    activeSection: LegacySection,
    networkBefore: NetworkSnapshot,
    networkAfter: NetworkSnapshot,
    civilizationFates: [CivFate],
    philosophyTree: PhilosophyTree,
    playerLineage: [CrewID],
    counterfactuals: [Counterfactual],
    epilogueText: string
}

type LegacySection = Network | Fates | Philosophy | Lineage | Counterfactuals | Epilogue

type CivFate = {
    civID: CivilizationID,
    finalState: FinalState,
    causeOfFate: string,
    yourInfluence: InfluenceLevel
}

type FinalState = Thriving | Transcended | Extinct | Transformed(CivilizationID)
```

---

## Supporting UI Components

### A. Logbook / Chronicle UI

**Purpose:** Timeline of everything that happened during your voyage.

**Content:**
- Contact events
- Births/deaths
- Trade outcomes
- Civilization changes
- Personal milestones

```ailang
type LogbookState = {
    entries: [LogEntry],
    filter: LogFilter,
    scrollPosition: int
}

type LogEntry = {
    year: int,
    yearSubjective: float,
    category: LogCategory,
    title: string,
    description: string
}

type LogCategory = Contact | Crew | Trade | Civilization | Personal
```

### B. Crew Relationship Matrix (Sociogram)

**Purpose:** Visual map of crew relationships.

**Shows:**
- Bonds (friendship, mentorship)
- Conflicts
- Romance
- Parent/child links
- Generational connections

```ailang
type SociogramState = {
    selectedCrew: Maybe(CrewID),
    relationshipFilter: RelationFilter,
    layoutMode: SociogramLayout
}

type RelationFilter = AllRelations | Positive | Negative | Family | Romantic
```

### C. Technology Inventory

**Purpose:** Horizontal slots showing installed and available technologies.

**Not a tree** - just categories and slots:
- Drive systems
- Energy systems
- Life support
- Communications
- Archives

### D. Philosophy Browser

**Purpose:** Gallery of encountered philosophies.

**Shows:**
- Philosophy name and symbol
- Core question it addresses
- Civilization modifiers
- Civilizations that follow it
- Compatibility with other philosophies

### E. Time Comparison UI

**Purpose:** Slider to visualize subjective vs. objective time.

**Shows:**
- Your subjective timeline (years lived)
- Galaxy objective timeline (years passed)
- Events synchronized to both timelines
- "Time debt" accumulated

---

## AILANG Implementation

### Core WorldMode Type

```ailang
module sim/ui

-- Master UI mode enum
type WorldMode =
    | ModeShipExploration(ShipExplorationState)
    | ModeDialogue(DialogueState)
    | ModeGalaxyMap(GalaxyMapState)
    | ModePlanetSurface(PlanetSurfaceState)
    | ModeCivDetail(CivDetailState)
    | ModeTrade(TradeState)
    | ModeJourneyPlan(JourneyPlanState)
    | ModeJourneyEvents(JourneyEventState)
    | ModeRuins(RuinsState)
    | ModeLegacy(LegacyState)
    | ModeLogbook(LogbookState)
    | ModeSociogram(SociogramState)

-- Updated World type to include mode
type World = {
    tick: int,
    mode: WorldMode,
    ship: Ship,
    galaxy: Galaxy,
    crew: [Crew],
    civilizations: [Civilization],
    logbook: [LogEntry],
    gameYear: int,           -- objective year
    subjectiveYear: float    -- player's experienced time
}
```

### Mode Transition Function

```ailang
-- Process mode-specific input and potentially transition
pure func processMode(world: World, input: FrameInput) -> World {
    match world.mode {
        ModeShipExploration(state) => processShipExploration(world, state, input),
        ModeDialogue(state) => processDialogue(world, state, input),
        ModeGalaxyMap(state) => processGalaxyMap(world, state, input),
        ModePlanetSurface(state) => processPlanetSurface(world, state, input),
        ModeCivDetail(state) => processCivDetail(world, state, input),
        ModeTrade(state) => processTrade(world, state, input),
        ModeJourneyPlan(state) => processJourneyPlan(world, state, input),
        ModeJourneyEvents(state) => processJourneyEvents(world, state, input),
        ModeRuins(state) => processRuins(world, state, input),
        ModeLegacy(state) => processLegacy(world, state, input),
        ModeLogbook(state) => processLogbook(world, state, input),
        ModeSociogram(state) => processSociogram(world, state, input)
    }
}

-- Transition to new mode
pure func transitionTo(world: World, newMode: WorldMode) -> World {
    { tick: world.tick,
      mode: newMode,
      ship: world.ship,
      galaxy: world.galaxy,
      crew: world.crew,
      civilizations: world.civilizations,
      logbook: world.logbook,
      gameYear: world.gameYear,
      subjectiveYear: world.subjectiveYear }
}
```

### Mode-Specific Rendering

```ailang
-- Generate draw commands based on current mode
pure func modeToDrawCmds(world: World) -> [DrawCmd] {
    match world.mode {
        ModeShipExploration(state) => renderShipExploration(world, state),
        ModeDialogue(state) => renderDialogue(world, state),
        ModeGalaxyMap(state) => renderGalaxyMap(world, state),
        ModePlanetSurface(state) => renderPlanetSurface(world, state),
        ModeCivDetail(state) => renderCivDetail(world, state),
        ModeTrade(state) => renderTrade(world, state),
        ModeJourneyPlan(state) => renderJourneyPlan(world, state),
        ModeJourneyEvents(state) => renderJourneyEvents(world, state),
        ModeRuins(state) => renderRuins(world, state),
        ModeLegacy(state) => renderLegacy(world, state),
        ModeLogbook(state) => renderLogbook(world, state),
        ModeSociogram(state) => renderSociogram(world, state)
    }
}
```

---

## Go/Engine Integration

### DrawCmd Extensions

```go
// sim_gen/types.go - Extended DrawCmd for UI rendering

type DrawCmd interface {
    isDrawCmd()
    ZIndex() int  // For proper layering
}

// Existing commands
type Sprite struct { ... }
type Rect struct { ... }
type Text struct { ... }

// New UI-specific commands
type Panel struct {
    X, Y, Width, Height float64
    BackgroundColor     int
    BorderColor         int
    ZIndex_             int
}

type Button struct {
    X, Y, Width, Height float64
    Label               string
    State               ButtonState  // Normal, Hover, Pressed, Disabled
    ZIndex_             int
}

type Portrait struct {
    X, Y, Width, Height float64
    SpriteID            int
    Emotion             Emotion
    ZIndex_             int
}

type Graph struct {
    X, Y, Width, Height float64
    Nodes               []GraphNode
    Edges               []GraphEdge
    ZIndex_             int
}
```

### Mode-Specific Renderers

```go
// engine/render/modes.go

func RenderMode(screen *ebiten.Image, output FrameOutput, mode WorldMode) {
    // Sort draw commands by Z-index
    cmds := sortByZIndex(output.Draw)

    // Render world layer (Z 0-29)
    for _, cmd := range cmds {
        if cmd.ZIndex() < 30 {
            renderCmd(screen, cmd)
        }
    }

    // Render UI layer (Z 30+)
    for _, cmd := range cmds {
        if cmd.ZIndex() >= 30 {
            renderUICmd(screen, cmd)
        }
    }
}

func renderUICmd(screen *ebiten.Image, cmd DrawCmd) {
    switch c := cmd.(type) {
    case Panel:
        renderPanel(screen, c)
    case Button:
        renderButton(screen, c)
    case Portrait:
        renderPortrait(screen, c)
    case Graph:
        renderGraph(screen, c)
    // ... other UI elements
    }
}
```

### Input Routing

```go
// engine/input/modes.go

func CaptureInput(mode WorldMode) FrameInput {
    base := captureBaseInput()  // Keys, mouse position

    switch mode.(type) {
    case ModeGalaxyMap:
        return captureMapInput(base)  // Pan, zoom
    case ModeDialogue:
        return captureDialogueInput(base)  // Choice selection
    case ModeTrade:
        return captureTradeInput(base)  // Drag and drop
    default:
        return base
    }
}
```

---

## Implementation Phases

### Phase 1: Mode Framework (v0.5.0)

| Task | Description |
|------|-------------|
| 1.1 | Define WorldMode enum and state types |
| 1.2 | Implement mode switching in Step |
| 1.3 | Add mode-specific input routing |
| 1.4 | Add Z-index sorting to renderer |
| 1.5 | Test mode transitions |

### Phase 2: Ship Exploration (v0.5.1)

| Task | Description |
|------|-------------|
| 2.1 | Ship interior tileset |
| 2.2 | Player movement on ship |
| 2.3 | Crew NPC display |
| 2.4 | Room interaction hotspots |
| 2.5 | Transition to Dialogue from crew click |

### Phase 3: Galaxy Map (v0.5.2)

| Task | Description |
|------|-------------|
| 3.1 | Starfield rendering |
| 3.2 | Star node display |
| 3.3 | Pan and zoom |
| 3.4 | Network edge rendering |
| 3.5 | Civilization tooltips |
| 3.6 | Journey planning trigger |

### Phase 4: Dialogue System (v0.5.3)

| Task | Description |
|------|-------------|
| 4.1 | Dialogue panel layout |
| 4.2 | Portrait rendering |
| 4.3 | Choice button display |
| 4.4 | Dialogue tree state machine |
| 4.5 | Outcome integration |

### Phase 5: Journey System (v0.6.0)

| Task | Description |
|------|-------------|
| 5.1 | Journey planning UI |
| 5.2 | Time dilation calculator |
| 5.3 | Crew projection display |
| 5.4 | Commit confirmation |
| 5.5 | Journey events mode |
| 5.6 | Transit simulation |

### Phase 6: Trade & Civilization (v0.6.1)

| Task | Description |
|------|-------------|
| 6.1 | Civilization detail screen |
| 6.2 | Trade UI layout |
| 6.3 | Drag-and-drop items |
| 6.4 | Impact preview |
| 6.5 | Trade execution |

### Phase 7: Exploration Modes (v0.7.0)

| Task | Description |
|------|-------------|
| 7.1 | Planet surface mode |
| 7.2 | Ruins mode |
| 7.3 | Artifact discovery |
| 7.4 | Ancient log reading |

### Phase 8: Endgame (v0.8.0)

| Task | Description |
|------|-------------|
| 8.1 | Legacy visualization layout |
| 8.2 | Network comparison view |
| 8.3 | Timeline rendering |
| 8.4 | Counterfactual display |
| 8.5 | Epilogue generation |

### Phase 9: Supporting UIs (v0.9.0)

| Task | Description |
|------|-------------|
| 9.1 | Logbook UI |
| 9.2 | Sociogram |
| 9.3 | Technology inventory |
| 9.4 | Philosophy browser |
| 9.5 | Time comparison slider |

---

## AILANG Constraints

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No mutable state | Can't incrementally update mode state | Functional update via `transitionTo` |
| Large enum pattern matching | 12 modes = deep match | Factor into mode-specific modules |
| Recursion depth | Large UI element lists | Keep UI element count bounded |
| No RNG (until v0.5.1) | Dialogue variety limited | Seed-based deterministic selection |
| List O(n) | Finding UI elements by ID | Use sorted lists, consider indices |

### Module Organization

To avoid massive files:

```
sim/
├── ui/
│   ├── mode.ail         -- WorldMode type, transitions
│   ├── ship.ail         -- ShipExploration logic
│   ├── dialogue.ail     -- Dialogue logic
│   ├── galaxy.ail       -- Galaxy map logic
│   ├── journey.ail      -- Journey planning/events
│   ├── civilization.ail -- Civ detail, trade
│   ├── exploration.ail  -- Planet surface, ruins
│   └── legacy.ail       -- Endgame
```

---

## Visual Style Considerations

### Consistent Design Language

All UI modes should share:
- Color palette (dark backgrounds, bright accents)
- Font hierarchy (titles, body, labels)
- Panel styling (borders, shadows)
- Button appearance (states, hover effects)
- Icon set (navigation, status indicators)

### Theme: Deep Time + Philosophical Weight

- Muted, contemplative colors
- Clean, readable typography
- Visual weight on choices (especially Journey Commit)
- Timelines prominent throughout
- Network graphs show interconnection

---

## Engine Feasibility Assessment

### Fully Supported

- Tile-based isometric/orthographic scenes
- Sprite animation
- Text rendering
- Panel overlays
- Button interactions
- 2D graph/network visualization
- Parallax backgrounds (layered 2D)
- Zoom via Ebiten transforms

### Achievable with Effort

- Smooth scrolling/panning
- Drag-and-drop (trade UI)
- Complex layouts (responsive panels)

### Not Feasible (Not Needed)

- True 3D planetary globes
- Free 3D camera
- Real-time lighting
- Full 3D characters

**Conclusion:** All proposed UI modes are implementable with current engine.

---

## Success Criteria

### Framework
- [ ] WorldMode enum defined in AILANG
- [ ] Mode switching works
- [ ] Z-index layering correct
- [ ] Input routing per mode

### Core Modes
- [ ] Ship exploration functional
- [ ] Galaxy map navigable
- [ ] Dialogue system working
- [ ] Journey planning complete with commit

### Supporting UIs
- [ ] Trade UI functional
- [ ] Civilization detail readable
- [ ] Logbook populated

### Endgame
- [ ] Legacy visualization renders
- [ ] Timeline displays correct
- [ ] Epilogue text generated

### Polish
- [ ] Consistent visual style
- [ ] Smooth transitions
- [ ] No rendering glitches
- [ ] Responsive input

---

## Future Considerations

### AI Integration Points

| Mode | AI Feature |
|------|------------|
| Dialogue | LLM-generated conversation |
| Portraits | AI-generated crew/alien faces |
| Legacy | AI-written epilogue |
| Ruins | AI-generated ancient logs |
| Civilization | AI-generated descriptions |

### Accessibility

- Keyboard navigation for all modes
- Text scaling options
- High-contrast mode
- Screen reader hints (future)

### Modding

- UI layout defined in data files
- Custom tilesets for modes
- Dialogue trees in external format
