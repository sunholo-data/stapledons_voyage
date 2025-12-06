# Endgame Legacy Visualization

**Version:** 0.8.0
**Status:** Planned
**Priority:** P0 (Game Climax)
**Complexity:** High
**AILANG Workarounds:** Large data aggregation, counterfactual simulation
**Depends On:** All previous systems

## Related Documents

- [UI Modes Architecture](../v0_5_0/ui-modes.md) - Mode framework
- [Galaxy Map](../v0_5_2/galaxy-map.md) - Network visualization
- [Journey System](../v0_6_0/journey-system.md) - Time tracking
- [Game Vision](../../../docs/game-vision.md) - Legacy as climax

## Problem Statement

The endgame is where Stapledon's Voyage delivers its emotional payload. After 100 subjective years of choices, players need to see:

- What their choices did to the galaxy
- Which civilizations thrived or died because of them
- How philosophies spread or vanished
- What their legacy is at Year 1,000,000
- What would have happened if they'd stayed home

**Current State:**
- No endgame sequence
- No legacy calculation
- No visualization
- No counterfactual simulation

**What's Needed:**
- Fast-forward simulation to Year 1,000,000
- Comprehensive legacy visualization
- Before/after network comparison
- Counterfactual "what if" simulations
- AI-generated epilogue

---

## Design Overview

### Endgame Philosophy

The endgame should feel like **reading the final chapter of an epic**:

- **Scale shift** - From personal to cosmic
- **Consequences manifest** - Every choice echoed forward
- **Bittersweet revelation** - Some things you saved, some you doomed
- **Your name in history** - Or forgotten to time

### Victory Conditions

From the Game Vision document:

| Victory | Description | Scoring |
|---------|-------------|---------|
| **The Shepherd** | Civilizations alive at Year 100 because of you | Count surviving civs you contacted |
| **The Gardener** | Philosophical diversity you preserved | Count unique philosophies × their populations |
| **The Unifier** | Contact network connectivity | Graph connectivity metrics |
| **The Witness** | Distance traveled, time experienced | Light-years × years observed |
| **The Founder** | Lasting institutions you started | Count surviving player-founded entities |
| **The Prometheus** | FTL/near-light spread traced to you | Count civs with drive tech from you |

---

## Detailed Specification

### Legacy State

```ailang
module sim/legacy

type LegacyState = {
    -- Fast-forward results
    finalYear: int,                      -- 1,000,000
    simulationComplete: bool,

    -- Current viewing
    activeSection: LegacySection,
    sectionProgress: float,              -- For reveal animations

    -- Computed legacy data
    networkBefore: NetworkSnapshot,
    networkAfter: NetworkSnapshot,
    civilizationFates: [CivFate],
    philosophyEvolution: PhilosophyTree,
    playerLineage: Lineage,
    counterfactuals: [Counterfactual],

    -- Scores
    victoryScores: [VictoryScore],
    selectedVictory: VictoryType,

    -- Generated content
    epilogueText: string,
    keyMoments: [KeyMoment]
}

type LegacySection =
    | SectionIntro                       -- "Your journey has ended..."
    | SectionFastForward                 -- Watching time accelerate
    | SectionNetwork                     -- Before/after galaxy
    | SectionFates                       -- Civilization outcomes
    | SectionPhilosophy                  -- Idea evolution
    | SectionLineage                     -- Your descendants
    | SectionVictory                     -- Score against chosen goal
    | SectionCounterfactual              -- What if...
    | SectionEpilogue                    -- Final words
    | SectionCredits                     -- Return to menu

type CivFate = {
    civID: CivilizationID,
    civName: string,
    initialState: CivState,
    finalState: FinalState,
    yearOfFate: int,
    causeOfFate: string,
    yourInfluence: InfluenceLevel,
    keyInteractions: [string]
}

type FinalState =
    | StillThriving
    | Transcended(int)                   -- Year of transcendence
    | Extinct(int)                       -- Year of extinction
    | Transformed(CivilizationID)        -- Merged/evolved into
    | Colonized(int)                     -- Number of worlds
    | Stagnant                           -- Survived but unchanging

type InfluenceLevel =
    | DirectCause                        -- You clearly caused this
    | MajorInfluence                     -- Your actions significantly affected
    | MinorInfluence                     -- Some connection to your choices
    | NoInfluence                        -- Would have happened anyway
    | Unknown                            -- Can't determine
```

### Network Comparison

```ailang
type NetworkSnapshot = {
    year: int,
    nodes: [NetworkNode],
    edges: [NetworkEdge],
    clusters: [Cluster],
    connectivity: float,                 -- 0-1
    diversity: float                     -- Philosophy diversity
}

type NetworkNode = {
    civID: CivilizationID,
    name: string,
    x: float,
    y: float,
    state: CivState,
    size: float,                         -- Based on population/influence
    color: int                           -- Based on philosophy
}

type NetworkEdge = {
    from: CivilizationID,
    to: CivilizationID,
    strength: float,                     -- Thickness
    edgeType: EdgeType,
    createdBy: Maybe(CrewID)            -- You or your descendants
}

type Cluster = {
    name: string,
    members: [CivilizationID],
    dominantPhilosophy: PhilosophyID,
    bounds: Rect
}

-- Generate comparison data
pure func generateNetworkComparison(gameStart: Galaxy, gameEnd: Galaxy) -> (NetworkSnapshot, NetworkSnapshot) {
    let before = snapshotNetwork(gameStart, gameStart.currentYear);
    let after = snapshotNetwork(gameEnd, gameEnd.currentYear);
    (before, after)
}

pure func snapshotNetwork(galaxy: Galaxy, year: int) -> NetworkSnapshot {
    let nodes = map(\s. starToNode(s), filter(\s. s.civilization != None, galaxy.stars));
    let edges = galaxy.edges;
    let clusters = identifyClusters(nodes, edges);
    let connectivity = calculateConnectivity(nodes, edges);
    let diversity = calculatePhilosophyDiversity(nodes);

    { year: year, nodes: nodes, edges: edges, clusters: clusters,
      connectivity: connectivity, diversity: diversity }
}
```

### Philosophy Tree

```ailang
type PhilosophyTree = {
    rootPhilosophies: [PhilosophyID],
    evolutions: [PhilosophyEvolution],
    extinctions: [PhilosophyExtinction],
    synthesises: [PhilosophySynthesis],
    currentDiversity: int
}

type PhilosophyEvolution = {
    original: PhilosophyID,
    evolved: PhilosophyID,
    year: int,
    catalyst: EvolutionCatalyst
}

type EvolutionCatalyst =
    | ContactWith(CivilizationID)
    | PlayerIntroduction
    | InternalDevelopment
    | CrisisResponse(string)

type PhilosophyExtinction = {
    philosophy: PhilosophyID,
    year: int,
    lastHolder: CivilizationID,
    cause: string,
    couldYouHavePrevented: bool
}

type PhilosophySynthesis = {
    parent1: PhilosophyID,
    parent2: PhilosophyID,
    child: PhilosophyID,
    year: int,
    synthesizer: CivilizationID
}
```

### Lineage Tracking

```ailang
type Lineage = {
    originalCrew: [CrewID],
    generations: int,
    finalDescendants: [Descendant],
    notableMembers: [NotableMember],
    lineageTree: LineageNode,
    legacyStatus: LineageLegacy
}

type Descendant = {
    id: CrewID,
    name: string,
    generation: int,
    birthYear: int,
    deathYear: Maybe(int),
    achievements: [string]
}

type NotableMember = {
    crewID: CrewID,
    name: string,
    title: string,
    achievement: string,
    year: int
}

type LineageNode = {
    member: CrewID,
    children: [LineageNode]
}

type LineageLegacy =
    | LineageExtinct(int)                -- Year line ended
    | LineageContinues                   -- Still alive at endgame
    | LineageTransformed                 -- Merged with alien species
    | LineageLegendary                   -- Became galactic founders
```

### Counterfactual Simulation

```ailang
type Counterfactual = {
    name: string,
    description: string,
    divergencePoint: CounterfactualType,
    alternateOutcome: AlternateOutcome
}

type CounterfactualType =
    | StayedHome                         -- Never left Earth
    | DifferentFirstContact              -- Met different civ first
    | NoTrade(CivilizationID, int)       -- Didn't trade at year
    | DifferentRoute                     -- Took different path
    | FasterTravel                       -- Used more aggressive velocities
    | SlowerTravel                       -- More conservative speeds

type AlternateOutcome = {
    survivingCivs: int,
    extinctCivs: int,
    philosophyDiversity: float,
    networkConnectivity: float,
    summary: string
}

-- Generate counterfactual "what if you stayed home"
pure func generateStayedHome(galaxy: Galaxy, playerHistory: [PlayerAction]) -> Counterfactual {
    -- Simulate galaxy without any player intervention
    let alternateGalaxy = simulateWithoutPlayer(galaxy);
    let outcome = assessAlternateOutcome(alternateGalaxy);

    {
        name: "If You Had Stayed Home",
        description: "What would have happened without your interference?",
        divergencePoint: StayedHome,
        alternateOutcome: outcome
    }
}

-- Compare specific trade decision
pure func generateTradeCounterfactual(civID: CivilizationID, year: int, galaxy: Galaxy) -> Counterfactual {
    let trade = findTrade(galaxy.history, civID, year);
    let alternateGalaxy = simulateWithoutTrade(galaxy, trade);
    let outcome = assessAlternateOutcome(alternateGalaxy);

    {
        name: "If You Hadn't Traded with " ++ getCivName(civID),
        description: "Year " ++ intToString(year) ++ ": " ++ describeTrade(trade),
        divergencePoint: NoTrade(civID, year),
        alternateOutcome: outcome
    }
}
```

### Victory Scoring

```ailang
type VictoryScore = {
    victoryType: VictoryType,
    score: int,
    maxPossible: int,
    rank: Rank,
    breakdown: [ScoreComponent]
}

type VictoryType = Shepherd | Gardener | Unifier | Witness | Founder | Prometheus

type Rank = Legendary | Excellent | Good | Average | Poor | Failure

type ScoreComponent = {
    name: string,
    value: int,
    description: string
}

pure func calculateVictoryScores(legacy: LegacyState) -> [VictoryScore] {
    [
        calculateShepherd(legacy),
        calculateGardener(legacy),
        calculateUnifier(legacy),
        calculateWitness(legacy),
        calculateFounder(legacy),
        calculatePrometheus(legacy)
    ]
}

pure func calculateShepherd(legacy: LegacyState) -> VictoryScore {
    let savedCivs = filter(\f.
        f.finalState == StillThriving &&
        f.yourInfluence == DirectCause || f.yourInfluence == MajorInfluence,
        legacy.civilizationFates);

    let score = length(savedCivs) * 100;
    let maxPossible = length(legacy.civilizationFates) * 100;

    {
        victoryType: Shepherd,
        score: score,
        maxPossible: maxPossible,
        rank: scoreToRank(score, maxPossible),
        breakdown: [
            { name: "Civilizations Saved", value: length(savedCivs), description: "Civs that survived due to your influence" }
        ]
    }
}

pure func calculateWitness(legacy: LegacyState) -> VictoryScore {
    let distanceTraveled = totalDistanceTraveled(legacy.journeyHistory);
    let yearsObserved = totalYearsObserved(legacy.journeyHistory);
    let uniqueContacts = countUniqueContacts(legacy.contactHistory);

    let score = floatToInt(distanceTraveled * 10.0) + yearsObserved + uniqueContacts * 50;

    {
        victoryType: Witness,
        score: score,
        maxPossible: 10000,
        rank: scoreToRank(score, 10000),
        breakdown: [
            { name: "Distance Traveled", value: floatToInt(distanceTraveled), description: "Light-years journeyed" },
            { name: "Time Observed", value: yearsObserved, description: "Subjective years of observation" },
            { name: "First Contacts", value: uniqueContacts, description: "Unique civilizations met" }
        ]
    }
}
```

---

## Visualization Rendering

### Section: Fast Forward

```ailang
pure func renderFastForward(state: LegacyState) -> [DrawCmd] {
    let currentYear = interpolateYear(state.sectionProgress);

    [
        -- Background: stars streaking
        renderStarStreak(state.sectionProgress),

        -- Center: year counter
        Text("Year " ++ formatLargeNumber(currentYear), 540.0, 300.0, 14, 10),

        -- Side events (civilizations rising/falling)
        renderFastForwardEvents(currentYear, state.civilizationFates),

        -- Progress bar
        Rect(200.0, 600.0, 880.0 * state.sectionProgress, 10.0, 2, 10)
    ]
}

pure func interpolateYear(progress: float) -> int {
    -- Logarithmic scale: slow at start, fast at end
    let startYear = 3000;
    let endYear = 1000000;
    startYear + floatToInt(intToFloat(endYear - startYear) * progress * progress)
}
```

### Section: Network Comparison

```ailang
pure func renderNetworkComparison(state: LegacyState) -> [DrawCmd] {
    let before = state.networkBefore;
    let after = state.networkAfter;

    -- Split screen: before on left, after on right
    let leftCmds = renderNetworkGraph(before, 50.0, 100.0, 580.0, 500.0, "Year " ++ intToString(before.year));
    let rightCmds = renderNetworkGraph(after, 650.0, 100.0, 580.0, 500.0, "Year " ++ intToString(after.year));

    -- Center divider with delta stats
    let deltaCmds = renderNetworkDelta(before, after);

    -- Legend
    let legendCmds = renderNetworkLegend();

    concat(concat(leftCmds, rightCmds), concat(deltaCmds, legendCmds))
}

pure func renderNetworkGraph(snapshot: NetworkSnapshot, x: float, y: float, w: float, h: float, title: string) -> [DrawCmd] {
    let titleCmd = Text(title, x + w / 2.0, y - 20.0, 8, 30);
    let borderCmd = Panel(x, y, w, h, 1, 7, 30);

    -- Draw edges first
    let edgeCmds = map(\e. renderNetworkEdge(e, snapshot.nodes, x, y, w, h), snapshot.edges);

    -- Draw nodes on top
    let nodeCmds = map(\n. renderNetworkNode(n, x, y, w, h), snapshot.nodes);

    titleCmd :: borderCmd :: concat(flatten(edgeCmds), flatten(nodeCmds))
}

pure func renderNetworkDelta(before: NetworkSnapshot, after: NetworkSnapshot) -> [DrawCmd] {
    let x = 600.0;
    let y = 320.0;

    let civDelta = length(after.nodes) - length(before.nodes);
    let connDelta = after.connectivity - before.connectivity;
    let divDelta = after.diversity - before.diversity;

    [
        Panel(x, y, 80.0, 150.0, 2, 7, 35),
        Text("Δ Civs: " ++ formatDelta(civDelta), x + 10.0, y + 20.0, 6, 36),
        Text("Δ Conn: " ++ formatFloatDelta(connDelta), x + 10.0, y + 50.0, 6, 36),
        Text("Δ Div: " ++ formatFloatDelta(divDelta), x + 10.0, y + 80.0, 6, 36)
    ]
}
```

### Section: Civilization Fates

```ailang
pure func renderCivFates(state: LegacyState) -> [DrawCmd] {
    -- Timeline visualization
    let timelineCmds = renderFateTimeline(state.civilizationFates);

    -- Selected fate detail
    let detailCmds = match state.selectedFate {
        Some(fate) => renderFateDetail(fate),
        None => []
    };

    -- Summary stats
    let statsCmds = renderFateSummary(state.civilizationFates);

    concat(concat(timelineCmds, detailCmds), statsCmds)
}

pure func renderFateTimeline(fates: [CivFate]) -> [DrawCmd] {
    -- Horizontal bars showing each civ's lifespan
    let sortedFates = sortBy(\f. f.initialYear, fates);

    mapWithIndex(\fate, i.
        let y = 100.0 + intToFloat(i) * 30.0;
        renderFateBar(fate, y),
        sortedFates)
}

pure func renderFateBar(fate: CivFate, y: float) -> [DrawCmd] {
    let startX = yearToX(fate.initialYear);
    let endX = yearToX(fate.yearOfFate);
    let color = fateToColor(fate.finalState);
    let width = endX - startX;

    [
        -- Name label
        Text(fate.civName, 50.0, y, 5, 30),
        -- Lifespan bar
        Rect(startX, y, width, 20.0, color, 30),
        -- Influence indicator
        renderInfluenceMarker(fate.yourInfluence, endX, y)
    ]
}
```

### Section: Epilogue

```ailang
pure func renderEpilogue(state: LegacyState) -> [DrawCmd] {
    [
        -- Starfield background
        renderStarfield(),

        -- Epilogue text panel
        Panel(200.0, 150.0, 880.0, 420.0, 1, 7, 40),

        -- Scrolling text
        TextWrapped(state.epilogueText, 230.0, 180.0, 820.0, 7, 41),

        -- "The End" or player's legacy title
        Text(state.legacyTitle, 540.0, 600.0, 10, 42)
    ]
}

-- Generate epilogue based on outcomes
pure func generateEpilogue(legacy: LegacyState) -> string {
    let intro = "A million years have passed since your final journey.\n\n";

    let civilizationSection = match countThriving(legacy.civilizationFates) {
        0 => "The galaxy is silent now. Every civilization you knew has passed into history. " ++
             "But their echoes remain in the patterns of stars, in the ruins on a thousand worlds.\n\n",
        n if n < 5 => "A handful of civilizations still persist, carrying forward the torch of consciousness. " ++
                      "Some bear the marks of your influence; others evolved on paths you never touched.\n\n",
        _ => "The galaxy thrives. Dozens of civilizations span the stars, connected by bonds of " ++
             "trade, philosophy, and shared history. Your name appears in their oldest records.\n\n"
    };

    let playerSection = match legacy.lineage.legacyStatus {
        LineageExtinct(_) => "Your lineage ended long ago. But the ideas you carried, the connections you forged—" ++
                            "these outlived any single family line.\n\n",
        LineageContinues => "Your descendants still exist, scattered across the galaxy. They carry fragments of " ++
                           "Earth's memory, stories of the first traveler who dared the void.\n\n",
        LineageLegendary => "They call your descendants the Starborn now. Your family's journey became " ++
                           "the template for all who followed.\n\n",
        _ => ""
    };

    let conclusion = "In the end, what matters is not whether you saved everyone, or united everything, " ++
                    "or left a legacy that spans eons. What matters is that you tried. " ++
                    "You faced the impossible loneliness of relativistic travel, and you chose to connect anyway.\n\n" ++
                    "The universe is vast and time is deep. But for a hundred years, you made a difference.\n\n" ++
                    "That is enough. That will always be enough.";

    intro ++ civilizationSection ++ playerSection ++ conclusion
}
```

---

## Go/Engine Integration

### Legacy Renderer

```go
// engine/render/legacy.go

type LegacyRenderer struct {
    fonts         *FontSet
    graphRenderer *GraphRenderer
    starfield     *StarfieldRenderer
}

func (r *LegacyRenderer) Render(screen *ebiten.Image, state LegacyState) {
    switch state.ActiveSection {
    case SectionIntro:
        r.renderIntro(screen, state)
    case SectionFastForward:
        r.renderFastForward(screen, state)
    case SectionNetwork:
        r.renderNetwork(screen, state)
    case SectionFates:
        r.renderFates(screen, state)
    case SectionPhilosophy:
        r.renderPhilosophy(screen, state)
    case SectionLineage:
        r.renderLineage(screen, state)
    case SectionVictory:
        r.renderVictory(screen, state)
    case SectionCounterfactual:
        r.renderCounterfactual(screen, state)
    case SectionEpilogue:
        r.renderEpilogue(screen, state)
    }

    // Navigation hint
    r.drawNavigationHint(screen, state)
}

func (r *LegacyRenderer) renderFastForward(screen *ebiten.Image, state LegacyState) {
    // Dramatic star streak effect
    r.starfield.DrawStreaking(screen, state.SectionProgress)

    // Giant year counter
    year := interpolateYear(state.SectionProgress)
    yearStr := formatLargeNumber(year)
    r.fonts.DrawGiant(screen, "Year "+yearStr, 640, 360)

    // Events flashing by
    r.drawFastForwardEvents(screen, state)
}
```

---

## Implementation Plan

### Phase 1: Fast Forward Simulation

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/legacy.go` | LegacyState type |
| 1.2 | `sim_gen/simulation.go` | Accelerated simulation |
| 1.3 | `engine/render/legacy.go` | Fast-forward visuals |
| 1.4 | Test | See time pass dramatically |

### Phase 2: Network Visualization

| Task | File | Description |
|------|------|-------------|
| 2.1 | `sim_gen/legacy.go` | NetworkSnapshot type |
| 2.2 | `engine/render/legacy.go` | Graph rendering |
| 2.3 | `engine/render/legacy.go` | Comparison layout |
| 2.4 | Test | See before/after graphs |

### Phase 3: Civilization Fates

| Task | File | Description |
|------|------|-------------|
| 3.1 | `sim_gen/legacy.go` | CivFate calculation |
| 3.2 | `engine/render/legacy.go` | Timeline rendering |
| 3.3 | `engine/render/legacy.go` | Fate detail panel |
| 3.4 | Test | See what happened to civs |

### Phase 4: Victory Scoring

| Task | File | Description |
|------|------|-------------|
| 4.1 | `sim_gen/legacy.go` | Score calculations |
| 4.2 | `engine/render/legacy.go` | Score display |
| 4.3 | Test | See victory scores |

### Phase 5: Counterfactuals

| Task | File | Description |
|------|------|-------------|
| 5.1 | `sim_gen/legacy.go` | Alternate simulation |
| 5.2 | `engine/render/legacy.go` | What-if display |
| 5.3 | Test | Compare outcomes |

### Phase 6: Epilogue

| Task | File | Description |
|------|------|-------------|
| 6.1 | `sim_gen/legacy.go` | Epilogue generation |
| 6.2 | `engine/render/legacy.go` | Epilogue display |
| 6.3 | Test | Read final text |

---

## Success Criteria

### Visualization
- [ ] Fast-forward is dramatic and readable
- [ ] Network comparison clear
- [ ] Civilization fates understandable
- [ ] Your influence visible

### Scoring
- [ ] All six victory types scored
- [ ] Breakdown makes sense
- [ ] Rank feels appropriate

### Emotional Impact
- [ ] Counterfactuals are meaningful
- [ ] Epilogue resonates
- [ ] Closure achieved

---

## Future Extensions

| Feature | Description |
|---------|-------------|
| AI epilogue | LLM-generated personalized ending |
| Achievement system | Specific accomplishments |
| Gallery mode | Review key moments |
| New game+ | Carry knowledge forward |
| Share legacy | Export summary to share |
