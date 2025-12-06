# Civilization Detail & Trade System

**Version:** 0.6.1
**Status:** Planned
**Priority:** P1 (Core Interaction)
**Complexity:** High
**AILANG Workarounds:** Record nesting, list filtering
**Depends On:** v0.5.2 Galaxy Map, v0.5.3 Dialogue System

## Related Documents

- [UI Modes Architecture](../v0_5_0/ui-modes.md) - Mode framework
- [Galaxy Map](../v0_5_2/galaxy-map.md) - Opens civ detail
- [Journey System](../v0_6_0/journey-system.md) - Changes civ states
- [Game Vision](../../../docs/game-vision.md) - Civilization simulation

## Problem Statement

Civilizations are the core entities players interact with. Players need to:
- Understand a civilization's current state, history, and philosophy
- Engage in meaningful trade of technology, knowledge, and ideas
- See the consequences of their trades on civilization trajectories
- Build relationships that span millennia

**Current State:**
- Civilizations exist in abstract
- No detail screen
- No trade mechanics
- No relationship tracking

**What's Needed:**
- Comprehensive civilization detail screen
- Trade UI with drag-and-drop
- Impact preview system
- Relationship and history tracking

---

## Part 1: Civilization Detail Screen

### Civilization Data

```ailang
module sim/civilization

type Civilization = {
    id: CivilizationID,
    name: string,
    species: Species,
    homeworld: StarID,
    colonies: [StarID],

    -- Core stats (0-100)
    population: int,              -- Population tier (log scale)
    energy: int,                  -- Energy access level
    technology: int,              -- Technological advancement
    stability: int,               -- Internal cohesion
    expansionDrive: int,          -- Desire to expand
    sustainability: int,          -- Long-term thinking
    contactOpenness: int,         -- Willingness to engage

    -- Philosophy
    philosophy: Philosophy,
    philosophyStrength: int,      -- How strongly held

    -- Risks
    existentialRisks: [ExistentialRisk],
    riskLevel: int,               -- Overall risk 0-100

    -- History
    foundingYear: int,
    currentState: CivState,
    history: [HistoricalEvent],
    contactHistory: [ContactEvent],

    -- Relationship with player
    trustLevel: int,              -- -100 to 100
    knowledgeOf: [KnowledgeID],   -- What they know about player
    tradesCompleted: int,
    lastInteractionYear: int
}

type Species = {
    name: string,
    physiology: Physiology,
    lifespan: int,                -- Average years
    communicationMode: CommMode,
    distinctiveTraits: [string]
}

type Physiology =
    | Humanoid
    | Crystalline
    | Gaseous
    | Aquatic
    | Silicon
    | Collective
    | Energy
    | Other(string)

type CommMode =
    | Verbal
    | Telepathic
    | Chemical
    | Electromagnetic
    | Symbolic
    | Temporal                    -- Communication across time
```

### Philosophy System

```ailang
type Philosophy = {
    id: PhilosophyID,
    name: string,
    coreQuestion: string,         -- The question this philosophy answers
    tenets: [Tenet],
    modifiers: PhilosophyModifiers,
    compatibleWith: [PhilosophyID],
    incompatibleWith: [PhilosophyID]
}

type Tenet = {
    name: string,
    description: string,
    effect: TenetEffect
}

type TenetEffect =
    | StabilityModifier(int)
    | ExpansionModifier(int)
    | TechModifier(int)
    | ContactModifier(int)
    | RiskModifier(int)

type PhilosophyModifiers = {
    stabilityBonus: int,
    techBonus: int,
    expansionBonus: int,
    contactBonus: int,
    sustainabilityBonus: int
}

-- Example philosophies
pure func examplePhilosophies() -> [Philosophy] {
    [
        {
            id: 1,
            name: "The Long View",
            coreQuestion: "How do we persist across deep time?",
            tenets: [
                { name: "Patience", description: "Measure progress in millennia", effect: StabilityModifier(20) },
                { name: "Sustainability", description: "Every action must be repeatable forever", effect: RiskModifier(-15) }
            ],
            modifiers: { stabilityBonus: 20, techBonus: -5, expansionBonus: -10, contactBonus: 0, sustainabilityBonus: 30 },
            compatibleWith: [3, 5],
            incompatibleWith: [2]
        },
        {
            id: 2,
            name: "The Burning",
            coreQuestion: "How do we achieve greatness before entropy claims us?",
            tenets: [
                { name: "Urgency", description: "Act now, for tomorrow is uncertain", effect: ExpansionModifier(30) },
                { name: "Glory", description: "Better a bright flame than a dim ember", effect: RiskModifier(20) }
            ],
            modifiers: { stabilityBonus: -15, techBonus: 20, expansionBonus: 30, contactBonus: 10, sustainabilityBonus: -25 },
            compatibleWith: [4],
            incompatibleWith: [1, 3]
        },
        {
            id: 3,
            name: "Sacred Mortality",
            coreQuestion: "What gives meaning to finite existence?",
            tenets: [
                { name: "Acceptance", description: "Death gives life meaning", effect: StabilityModifier(15) },
                { name: "Legacy", description: "We live on in what we create", effect: TechModifier(10) }
            ],
            modifiers: { stabilityBonus: 15, techBonus: 10, expansionBonus: 0, contactBonus: 5, sustainabilityBonus: 10 },
            compatibleWith: [1, 5],
            incompatibleWith: [4]
        }
    ]
}
```

### Historical Events

```ailang
type HistoricalEvent = {
    year: int,
    eventType: HistoricalEventType,
    description: string,
    impact: EventImpact
}

type HistoricalEventType =
    | Founded
    | TechnologicalBreakthrough(string)
    | PhilosophicalShift(PhilosophyID, PhilosophyID)
    | Expansion(StarID)
    | War(CivilizationID, WarOutcome)
    | Plague(int)                       -- Mortality percentage
    | FirstContact(CivilizationID)
    | TradeAgreement(CivilizationID)
    | CivilWar(string)
    | Transcendence
    | NearExtinction(string)
    | Extinction(string)

type EventImpact = {
    populationDelta: int,
    stabilityDelta: int,
    techDelta: int,
    philosophyEffect: Maybe(PhilosophyID)
}

type ContactEvent = {
    year: int,
    withPlayer: bool,
    withCiv: Maybe(CivilizationID),
    outcome: ContactOutcome,
    tradedItems: [TradeItem],
    relationshipChange: int
}

type ContactOutcome =
    | Peaceful
    | Tense
    | Hostile
    | Transformative
    | Enlightening
```

### Civilization Detail State

```ailang
type CivDetailState = {
    civID: CivilizationID,
    civ: Civilization,
    activeTab: CivTab,
    historyScrollPos: int,
    selectedEvent: Maybe(int),
    tradePreview: Maybe(TradePreview),
    relationshipExpanded: bool
}

type CivTab =
    | TabOverview
    | TabPhilosophy
    | TabHistory
    | TabRelationships
    | TabTrade
```

### Detail Screen Rendering

```ailang
pure func renderCivDetail(state: CivDetailState) -> [DrawCmd] {
    let civ = state.civ;

    -- Background panel
    let bgCmds = [Panel(50.0, 50.0, 1180.0, 620.0, 1, 7, 40)];

    -- Header
    let headerCmds = renderCivHeader(civ);

    -- Tab bar
    let tabCmds = renderTabBar(state.activeTab);

    -- Tab content
    let contentCmds = match state.activeTab {
        TabOverview => renderOverviewTab(civ),
        TabPhilosophy => renderPhilosophyTab(civ),
        TabHistory => renderHistoryTab(civ, state),
        TabRelationships => renderRelationshipsTab(civ),
        TabTrade => renderTradeTab(civ, state)
    };

    concat(concat(bgCmds, headerCmds), concat(tabCmds, contentCmds))
}

pure func renderCivHeader(civ: Civilization) -> [DrawCmd] {
    [
        -- Name and species
        Text(civ.name, 80.0, 70.0, 10, 41),
        Text("(" ++ civ.species.name ++ ")", 80.0, 100.0, 6, 41),

        -- State indicator
        renderStateIndicator(civ.currentState, 1100.0, 70.0),

        -- Trust meter
        Panel(900.0, 90.0, 200.0, 20.0, 0, 7, 41),
        renderTrustBar(civ.trustLevel, 905.0, 95.0),
        Text("Trust", 900.0, 75.0, 6, 41)
    ]
}

pure func renderOverviewTab(civ: Civilization) -> [DrawCmd] {
    let y = 180.0;
    let colWidth = 350.0;

    [
        -- Column 1: Core Stats
        Text("Core Statistics", 80.0, y, 8, 41),
        renderStatBar("Population", civ.population, 80.0, y + 30.0),
        renderStatBar("Energy", civ.energy, 80.0, y + 60.0),
        renderStatBar("Technology", civ.technology, 80.0, y + 90.0),
        renderStatBar("Stability", civ.stability, 80.0, y + 120.0),

        -- Column 2: Traits
        Text("Characteristics", 80.0 + colWidth, y, 8, 41),
        renderStatBar("Expansion Drive", civ.expansionDrive, 80.0 + colWidth, y + 30.0),
        renderStatBar("Sustainability", civ.sustainability, 80.0 + colWidth, y + 60.0),
        renderStatBar("Contact Openness", civ.contactOpenness, 80.0 + colWidth, y + 90.0),
        renderStatBar("Risk Level", civ.riskLevel, 80.0 + colWidth, y + 120.0),

        -- Column 3: Summary
        Text("Summary", 80.0 + colWidth * 2.0, y, 8, 41),
        TextWrapped(generateCivSummary(civ), 80.0 + colWidth * 2.0, y + 30.0, 300.0, 6, 41)
    ]
}

pure func generateCivSummary(civ: Civilization) -> string {
    let stateDesc = match civ.currentState {
        Thriving(_) => "a thriving civilization",
        Declining(_) => "a declining civilization",
        Extinct(_) => "an extinct civilization",
        Transcended(_) => "a transcended civilization",
        _ => "a civilization"
    };

    "The " ++ civ.name ++ " are " ++ stateDesc ++ " following the philosophy of " ++
    civ.philosophy.name ++ ". They believe that the answer to '" ++
    civ.philosophy.coreQuestion ++ "' shapes all their decisions."
}
```

---

## Part 2: Trade System

### Trade Data Structures

```ailang
type TradeState = {
    partnerCiv: CivilizationID,
    offering: [TradeItem],
    requesting: [TradeItem],
    balance: TradeBalance,
    acceptProbability: float,
    impactPreview: TradeImpact,
    warnings: [string],
    phase: TradePhase
}

type TradePhase =
    | Composing              -- Player building offer
    | Reviewing              -- Showing full impact
    | Negotiating            -- Civ counter-offer
    | Accepting              -- Final confirmation
    | Completed              -- Trade done
    | Rejected               -- Trade refused

type TradeItem =
    | Technology(TechID)
    | Knowledge(KnowledgeID)
    | Artifact(ArtifactID)
    | Philosophy(PhilosophyID)
    | Resource(ResourceType, int)
    | Contact(CivilizationID)     -- Introduction to another civ
    | Promise(PromiseType)        -- Future commitment

type TradeBalance = {
    playerValue: int,
    civValue: int,
    fairness: float              -- 0-1, how balanced
}

type TradeImpact = {
    -- Immediate effects
    civStabilityDelta: int,
    civTechDelta: int,
    civPhilosophyRisk: float,
    trustChange: int,

    -- Long-term projections
    yearOneProjection: CivStateProjection,
    yearTenProjection: CivStateProjection,
    yearHundredProjection: CivStateProjection,

    -- Warnings
    destabilizationRisk: bool,
    extinctionRisk: bool,
    philosophyConflict: bool
}

type CivStateProjection = {
    survivalProbability: float,
    expectedState: CivState,
    keyFactors: [string]
}
```

### Trade Value Calculation

```ailang
-- Calculate value of trade item to a civilization
pure func itemValue(item: TradeItem, civ: Civilization) -> int {
    match item {
        Technology(techID) => techValueForCiv(techID, civ),
        Knowledge(knowID) => knowledgeValueForCiv(knowID, civ),
        Artifact(artID) => artifactValue(artID),
        Philosophy(philID) => philosophyValueForCiv(philID, civ),
        Resource(resType, amount) => resourceValue(resType, amount, civ),
        Contact(civID) => contactValue(civID, civ),
        Promise(promType) => promiseValue(promType)
    }
}

-- Tech value depends on civ's current level
pure func techValueForCiv(techID: TechID, civ: Civilization) -> int {
    let tech = getTech(techID);
    let techLevel = tech.level;
    let civLevel = civ.technology;

    -- Higher value if tech is significantly above their level
    if techLevel > civLevel + 30 then
        100  -- Revolutionary
    else if techLevel > civLevel + 15 then
        60   -- Advanced
    else if techLevel > civLevel then
        30   -- Useful
    else
        5    -- Already known or obsolete
}

-- Philosophy value depends on compatibility
pure func philosophyValueForCiv(philID: PhilosophyID, civ: Civilization) -> int {
    let phil = getPhilosophy(philID);
    let civPhil = civ.philosophy;

    if contains(civPhil.compatibleWith, philID) then
        50   -- Compatible, interesting
    else if contains(civPhil.incompatibleWith, philID) then
        -30  -- Destabilizing
    else
        20   -- Neutral curiosity
}
```

### Trade Impact Calculation

```ailang
-- Calculate full impact of proposed trade
pure func calculateTradeImpact(trade: TradeState, civ: Civilization) -> TradeImpact {
    -- Calculate immediate effects
    let stabilityDelta = calculateStabilityImpact(trade.offering, civ);
    let techDelta = calculateTechImpact(trade.offering, civ);
    let philRisk = calculatePhilosophyRisk(trade.offering, civ);
    let trustDelta = calculateTrustChange(trade, civ);

    -- Project future states
    let civAfterTrade = applytTadeEffects(civ, trade.offering);
    let yearOne = projectCiv(civAfterTrade, 1);
    let yearTen = projectCiv(civAfterTrade, 10);
    let yearHundred = projectCiv(civAfterTrade, 100);

    -- Check for warnings
    let destab = stabilityDelta < -20 || civ.stability + stabilityDelta < 20;
    let extinct = yearHundred.survivalProbability < 0.5;
    let philConflict = philRisk > 0.3;

    {
        civStabilityDelta: stabilityDelta,
        civTechDelta: techDelta,
        civPhilosophyRisk: philRisk,
        trustChange: trustDelta,
        yearOneProjection: yearOne,
        yearTenProjection: yearTen,
        yearHundredProjection: yearHundred,
        destabilizationRisk: destab,
        extinctionRisk: extinct,
        philosophyConflict: philConflict
    }
}

-- Project civilization state N years into future
pure func projectCiv(civ: Civilization, years: int) -> CivStateProjection {
    -- Simple projection model
    let stabilityDecay = if civ.stability < 30 then years * 2 else 0;
    let techGrowth = (civ.technology * years) / 100;

    let survivalProb = intToFloat(civ.stability + civ.sustainability) / 200.0;
    let adjustedSurvival = survivalProb - (intToFloat(years) * 0.001);  -- Long-term entropy

    let expectedState = if adjustedSurvival < 0.3 then
        Extinct(years)
    else if civ.technology + techGrowth > 95 then
        Transcended(years)
    else if civ.stability - stabilityDecay < 20 then
        Declining(years)
    else
        Thriving(civ.population);

    {
        survivalProbability: max(0.0, adjustedSurvival),
        expectedState: expectedState,
        keyFactors: generateFactors(civ, years)
    }
}
```

### Trade UI Rendering

```ailang
pure func renderTradeUI(state: TradeState, civ: Civilization) -> [DrawCmd] {
    -- Background
    let bgCmds = [Rect(0.0, 0.0, 1280.0, 720.0, 0, 50)];

    -- Main trade panel
    let panelCmds = [Panel(100.0, 100.0, 1080.0, 520.0, 1, 7, 51)];

    -- Title
    let titleCmds = [Text("Trade with " ++ civ.name, 540.0, 120.0, 10, 52)];

    -- Two columns: Offer and Request
    let offerCmds = renderTradeColumn("You Offer", state.offering, 150.0, true, state);
    let requestCmds = renderTradeColumn("You Request", state.requesting, 690.0, false, state);

    -- Center: Balance and impact
    let centerCmds = renderTradeCenter(state);

    -- Bottom: Buttons
    let buttonCmds = renderTradeButtons(state);

    concat(concat(bgCmds, panelCmds),
           concat(titleCmds,
                  concat(offerCmds,
                         concat(requestCmds,
                                concat(centerCmds, buttonCmds)))))
}

pure func renderTradeColumn(title: string, items: [TradeItem], x: float, isOffer: bool, state: TradeState) -> [DrawCmd] {
    let header = [
        Panel(x, 170.0, 400.0, 350.0, 2, 7, 52),
        Text(title, x + 150.0, 185.0, 8, 53)
    ];

    let itemCmds = mapWithIndex(\item, i.
        renderTradeItem(item, x + 20.0, 220.0 + intToFloat(i) * 50.0, isOffer, state), items);

    -- Drop zone indicator
    let dropZone = if state.draggingItem != None then
        [Rect(x + 10.0, 470.0, 380.0, 40.0, 3, 52)]
    else [];

    concat(header, concat(flatten(itemCmds), dropZone))
}

pure func renderTradeItem(item: TradeItem, x: float, y: float, canRemove: bool, state: TradeState) -> [DrawCmd] {
    let name = tradeItemName(item);
    let value = tradeItemDisplayValue(item);
    let isHovered = state.hoveredItem == Some(item);

    [
        Panel(x, y, 360.0, 40.0, if isHovered then 3 else 2, 7, 53),
        Text(name, x + 10.0, y + 12.0, 7, 54),
        Text(value, x + 280.0, y + 12.0, 6, 54),
        if canRemove then Button(x + 330.0, y + 5.0, 25.0, 25.0, "X", 4, 54) else Noop
    ]
}

pure func renderTradeCenter(state: TradeState) -> [DrawCmd] {
    let x = 560.0;
    let y = 280.0;

    [
        -- Balance meter
        Panel(x - 50.0, y, 160.0, 100.0, 1, 7, 52),
        Text("Balance", x, y + 10.0, 7, 53),
        renderBalanceMeter(state.balance, x, y + 40.0),

        -- Accept probability
        Text("Accept Chance", x, y + 80.0, 6, 53),
        Text(formatPercent(state.acceptProbability), x, y + 95.0, 8, 53),

        -- Impact preview button
        Button(x - 40.0, y + 120.0, 140.0, 35.0, "View Impact", 3, 53)
    ]
}

pure func renderBalanceMeter(balance: TradeBalance, x: float, y: float) -> DrawCmd {
    -- Visual meter showing fairness
    -- Green = fair, Yellow = unbalanced, Red = exploitative
    let color = if balance.fairness > 0.7 then 2      -- Green
                else if balance.fairness > 0.4 then 3  -- Yellow
                else 4;                                -- Red
    Rect(x - 30.0, y, 120.0 * balance.fairness, 20.0, color, 53)
}
```

### Trade Impact Preview

```
┌─────────────────────────────────────────────────────────────────┐
│                    TRADE IMPACT PREVIEW                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  IMMEDIATE EFFECTS:                                              │
│  • Stability: -15 (Destabilizing technology)                    │
│  • Technology: +25 (Significant advancement)                     │
│  • Trust: +10 (Fair trade)                                       │
│                                                                  │
│  ════════════════════════════════════════════════════════════   │
│                                                                  │
│  PROJECTED FUTURES:                                              │
│                                                                  │
│  In 1 year:    85% survival  |  State: Thriving → Unstable      │
│  In 10 years:  72% survival  |  State: Recovering                │
│  In 100 years: 61% survival  |  State: Transformed               │
│                                                                  │
│  ────────────────────────────────────────────────────────────   │
│                                                                  │
│  ⚠ WARNING: Fusion technology may accelerate expansion          │
│     beyond sustainable levels.                                   │
│                                                                  │
│  ⚠ WARNING: This trade significantly destabilizes their         │
│     current philosophy. Civil conflict possible.                 │
│                                                                  │
│            [ Back ]                    [ Proceed ]               │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Go/Engine Integration

### Drag and Drop System

```go
// engine/input/trade.go

type DragState struct {
    IsDragging bool
    Item       TradeItem
    StartX     float64
    StartY     float64
    CurrentX   float64
    CurrentY   float64
}

func CaptureTradeInput(state TradeState) FrameInput {
    var input FrameInput

    mx, my := ebiten.CursorPosition()
    input.MouseX, input.MouseY = float64(mx), float64(my)

    // Start drag
    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        item := findItemUnderCursor(state, input.MouseX, input.MouseY)
        if item != nil {
            input.StartDrag = true
            input.DragItem = item
        }
    }

    // Continue drag
    if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && state.DragState.IsDragging {
        input.Dragging = true
        input.DragX = input.MouseX
        input.DragY = input.MouseY
    }

    // End drag
    if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && state.DragState.IsDragging {
        input.EndDrag = true
        input.DropZone = findDropZone(input.MouseX, input.MouseY)
    }

    return input
}
```

### Trade Renderer

```go
// engine/render/trade.go

func (r *TradeRenderer) Render(screen *ebiten.Image, state TradeState, civ Civilization) {
    // Background dim
    r.drawDimmer(screen, 0.8)

    // Main panel
    r.panels.Draw(screen, 100, 100, 1080, 520)

    // Title
    r.fonts.DrawTitle(screen, "Trade with "+civ.Name, 640, 130)

    // Columns
    r.drawOfferColumn(screen, state)
    r.drawRequestColumn(screen, state)

    // Center panel
    r.drawBalance(screen, state)

    // Drag item (if dragging)
    if state.DragState.IsDragging {
        r.drawDraggedItem(screen, state.DragState)
    }

    // Buttons
    r.drawTradeButtons(screen, state)
}
```

---

## Implementation Plan

### Phase 1: Civilization Types

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/civilization.go` | Civilization, Species types |
| 1.2 | `sim_gen/civilization.go` | Philosophy type |
| 1.3 | `sim_gen/civilization.go` | Historical events |
| 1.4 | Test | Types compile |

### Phase 2: Detail Screen Basic

| Task | File | Description |
|------|------|-------------|
| 2.1 | `sim_gen/civdetail.go` | CivDetailState type |
| 2.2 | `engine/render/civdetail.go` | Panel layout |
| 2.3 | `engine/render/civdetail.go` | Tab bar |
| 2.4 | `engine/render/civdetail.go` | Overview tab |
| 2.5 | Test | See civ details |

### Phase 3: Detail Screen Tabs

| Task | File | Description |
|------|------|-------------|
| 3.1 | `engine/render/civdetail.go` | Philosophy tab |
| 3.2 | `engine/render/civdetail.go` | History tab |
| 3.3 | `engine/render/civdetail.go` | Relationships tab |
| 3.4 | Test | All tabs work |

### Phase 4: Trade Types

| Task | File | Description |
|------|------|-------------|
| 4.1 | `sim_gen/trade.go` | TradeState, TradeItem types |
| 4.2 | `sim_gen/trade.go` | TradeImpact type |
| 4.3 | `sim_gen/funcs.go` | Value calculation |
| 4.4 | Test | Trade calculations |

### Phase 5: Trade UI

| Task | File | Description |
|------|------|-------------|
| 5.1 | `engine/render/trade.go` | Trade panel layout |
| 5.2 | `engine/input/trade.go` | Drag and drop |
| 5.3 | `engine/render/trade.go` | Item rendering |
| 5.4 | Test | Drag items between columns |

### Phase 6: Trade Impact

| Task | File | Description |
|------|------|-------------|
| 6.1 | `sim_gen/funcs.go` | Impact calculation |
| 6.2 | `sim_gen/funcs.go` | Projection system |
| 6.3 | `engine/render/trade.go` | Impact preview |
| 6.4 | Test | See trade consequences |

### Phase 7: Trade Execution

| Task | File | Description |
|------|------|-------------|
| 7.1 | `sim_gen/funcs.go` | Accept probability |
| 7.2 | `sim_gen/funcs.go` | Trade execution |
| 7.3 | `sim_gen/funcs.go` | World state update |
| 7.4 | Test | Complete trades |

---

## Testing Strategy

### Manual Testing

```bash
make run-mock
# 1. Navigate to galaxy map
# 2. Click civilization star
# 3. See detail screen
# 4. Switch between tabs
# 5. Click Trade tab
# 6. Drag items to offer
# 7. See balance update
# 8. View impact preview
# 9. Complete trade
# 10. Verify changes persist
```

### Automated Testing

```go
func TestCivilizationProjection(t *testing.T)
func TestTradeValueCalculation(t *testing.T)
func TestTradeBalanceFairness(t *testing.T)
func TestImpactCalculation(t *testing.T)
func TestTradeExecution(t *testing.T)
```

---

## Success Criteria

### Civilization Detail
- [ ] All tabs render correctly
- [ ] Stats display accurately
- [ ] History scrolls
- [ ] Philosophy explained

### Trade System
- [ ] Drag and drop works
- [ ] Balance updates in real-time
- [ ] Impact preview shows consequences
- [ ] Warnings appear appropriately

### Integration
- [ ] Trades affect civilization stats
- [ ] History records trades
- [ ] Trust level changes
- [ ] Long-term effects occur

---

## Future Extensions

| Feature | Description |
|---------|-------------|
| Counter-offers | Civ proposes alternatives |
| Multi-party | Trade involving multiple civs |
| Long-term contracts | Ongoing agreements |
| Trade routes | Automated exchanges |
| Embargo | Refuse trade with certain civs |
| Black market | High-risk unofficial trades |
