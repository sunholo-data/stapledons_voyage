# Supporting UI Systems

**Version:** 0.9.0
**Status:** Planned
**Priority:** P2 (Quality of Life)
**Complexity:** Medium
**AILANG Workarounds:** List filtering, data aggregation
**Depends On:** All core UI modes

## Related Documents

- [UI Modes Architecture](../v0_5_0/ui-modes.md) - Mode framework
- [Ship Exploration](../v0_5_1/ship-exploration.md) - Ship context
- [Journey System](../v0_6_0/journey-system.md) - Time tracking

## Overview

Supporting UIs enhance the core experience without being essential to gameplay:

1. **Logbook/Chronicle** - Timeline of everything
2. **Crew Sociogram** - Relationship visualization
3. **Technology Inventory** - What you have
4. **Philosophy Browser** - Ideas encountered
5. **Time Comparison** - Subjective vs objective

---

## 1. Logbook / Chronicle UI

### Purpose

The logbook is the player's memory - a searchable, filterable timeline of everything that happened during the voyage.

### State

```ailang
type LogbookState = {
    entries: [LogEntry],
    filter: LogFilter,
    searchQuery: string,
    searchResults: [LogEntryID],
    selectedEntry: Maybe(LogEntryID),
    scrollPosition: int,
    sortOrder: SortOrder
}

type LogEntry = {
    id: LogEntryID,
    objectiveYear: int,              -- Galaxy year
    subjectiveYear: float,           -- Player's experienced time
    category: LogCategory,
    title: string,
    summary: string,
    details: string,
    linkedEntities: [EntityRef],
    importance: Importance
}

type LogCategory =
    | CatJourney                     -- Departures, arrivals
    | CatContact                     -- First contacts, meetings
    | CatTrade                       -- Exchanges
    | CatCrew                        -- Births, deaths, relationships
    | CatCivilization                -- Civ changes observed
    | CatDiscovery                   -- Artifacts, knowledge
    | CatShip                        -- System events
    | CatPersonal                    -- Player milestones

type LogFilter = {
    categories: [LogCategory],       -- Empty = all
    yearRange: Maybe((int, int)),
    importance: Maybe(Importance),
    involvedEntity: Maybe(EntityRef)
}

type Importance = Critical | Major | Minor | Routine

type EntityRef =
    | CrewRef(CrewID)
    | CivRef(CivilizationID)
    | StarRef(StarID)
    | ArtifactRef(ArtifactID)
```

### Automatic Entry Generation

```ailang
-- Generate log entries automatically from game events
pure func generateLogEntry(event: GameEvent, world: World) -> LogEntry {
    match event {
        JourneyComplete(from, to, subjYears) => {
            id: nextLogID(),
            objectiveYear: world.gameYear,
            subjectiveYear: world.subjectiveYear,
            category: CatJourney,
            title: "Arrival at " ++ getStarName(to),
            summary: "Completed journey from " ++ getStarName(from),
            details: "Traveled " ++ formatFloat(distance(from, to)) ++ " light-years. " ++
                     formatFloat(subjYears) ++ " years passed aboard ship.",
            linkedEntities: [StarRef(from), StarRef(to)],
            importance: Major
        },
        CrewDeath(crewID, cause) => {
            let crew = findCrew(crewID);
            {
                id: nextLogID(),
                objectiveYear: world.gameYear,
                subjectiveYear: world.subjectiveYear,
                category: CatCrew,
                title: crew.name ++ " has died",
                summary: deathCauseToSummary(cause),
                details: generateObituary(crew, cause),
                linkedEntities: [CrewRef(crewID)],
                importance: Critical
            }
        },
        FirstContact(civID) => {
            let civ = findCiv(civID);
            {
                id: nextLogID(),
                objectiveYear: world.gameYear,
                subjectiveYear: world.subjectiveYear,
                category: CatContact,
                title: "First Contact: " ++ civ.name,
                summary: "Established contact with " ++ civ.species.name,
                details: generateFirstContactDetails(civ),
                linkedEntities: [CivRef(civID)],
                importance: Critical
            }
        },
        _ => defaultLogEntry(event, world)
    }
}
```

### Rendering

```ailang
pure func renderLogbook(state: LogbookState) -> [DrawCmd] {
    -- Background panel
    let bgCmds = [Panel(100.0, 50.0, 1080.0, 620.0, 1, 7, 40)];

    -- Header with search and filters
    let headerCmds = renderLogbookHeader(state);

    -- Entry list (left side)
    let listCmds = renderEntryList(state);

    -- Entry detail (right side)
    let detailCmds = match state.selectedEntry {
        Some(id) => renderEntryDetail(findEntry(state.entries, id)),
        None => [Text("Select an entry to view details", 800.0, 350.0, 6, 41)]
    };

    -- Timeline visualization (bottom)
    let timelineCmds = renderMiniTimeline(state);

    concat(bgCmds, concat(headerCmds, concat(listCmds, concat(detailCmds, timelineCmds))))
}

pure func renderEntryList(state: LogbookState) -> [DrawCmd] {
    let filteredEntries = applyFilters(state.entries, state.filter, state.searchQuery);
    let visibleEntries = take(20, drop(state.scrollPosition, filteredEntries));

    let entryCmds = mapWithIndex(\entry, i.
        let y = 130.0 + intToFloat(i) * 25.0;
        let isSelected = state.selectedEntry == Some(entry.id);
        renderEntryListItem(entry, y, isSelected),
        visibleEntries);

    let scrollbarCmds = renderScrollbar(state.scrollPosition, length(filteredEntries), 20);

    concat(flatten(entryCmds), scrollbarCmds)
}

pure func renderEntryListItem(entry: LogEntry, y: float, selected: bool) -> [DrawCmd] {
    let bgColor = if selected then 3 else 2;
    let categoryIcon = categoryToIcon(entry.category);
    let importanceColor = importanceToColor(entry.importance);

    [
        Rect(120.0, y, 400.0, 22.0, bgColor, 41),
        Text(categoryIcon, 125.0, y + 4.0, 5, 42),
        Rect(145.0, y + 2.0, 4.0, 18.0, importanceColor, 42),  -- Importance bar
        Text(truncate(entry.title, 40), 155.0, y + 4.0, 5, 42),
        Text(formatYear(entry.objectiveYear), 470.0, y + 4.0, 5, 42)
    ]
}
```

---

## 2. Crew Sociogram

### Purpose

Visual map of all crew relationships - bonds, conflicts, romance, family ties.

### State

```ailang
type SociogramState = {
    crew: [Crew],
    relationships: [Relationship],
    selectedCrew: Maybe(CrewID),
    hoveredCrew: Maybe(CrewID),
    filter: RelationFilter,
    layout: SociogramLayout,
    zoomLevel: float,
    panOffset: Vec2
}

type Relationship = {
    crew1: CrewID,
    crew2: CrewID,
    relationType: RelationType,
    strength: int,                   -- -100 to 100
    since: int                       -- Year established
}

type RelationType =
    | Family(FamilyType)
    | Romantic
    | Friendship
    | Professional
    | Conflict
    | Mentorship

type FamilyType = Parent | Child | Sibling | Spouse | Grandparent | Grandchild

type RelationFilter =
    | FilterAll
    | FilterPositive
    | FilterNegative
    | FilterFamily
    | FilterRomantic
    | FilterProfessional

type SociogramLayout =
    | ForceDirected                  -- Physics-based positioning
    | Hierarchical                   -- Family tree style
    | Circular                       -- Ring layout
    | Timeline                       -- Ordered by birth year
```

### Layout Calculation

```ailang
-- Force-directed layout (simplified)
pure func calculateForceLayout(crew: [Crew], relationships: [Relationship]) -> [(CrewID, Vec2)] {
    let initial = initialCircleLayout(crew);
    -- Apply multiple iterations of force calculation
    let iteration1 = applyForces(initial, relationships);
    let iteration2 = applyForces(iteration1, relationships);
    let iteration3 = applyForces(iteration2, relationships);
    iteration3
}

pure func applyForces(positions: [(CrewID, Vec2)], relationships: [Relationship]) -> [(CrewID, Vec2)] {
    map(\(id, pos).
        let forces = calculateCrewForces(id, pos, positions, relationships);
        let newPos = addVec(pos, scaleVec(forces, 0.1));
        (id, clampToScreen(newPos)),
        positions)
}

pure func calculateCrewForces(id: CrewID, pos: Vec2, allPositions: [(CrewID, Vec2)], rels: [Relationship]) -> Vec2 {
    -- Repulsion from all other crew
    let repulsion = foldl(\acc, (otherId, otherPos).
        if id == otherId then acc
        else addVec(acc, repulsionForce(pos, otherPos)),
        Vec2(0.0, 0.0), allPositions);

    -- Attraction to related crew
    let myRels = filter(\r. r.crew1 == id || r.crew2 == id, rels);
    let attraction = foldl(\acc, rel.
        let otherId = if rel.crew1 == id then rel.crew2 else rel.crew1;
        let otherPos = findPos(allPositions, otherId);
        addVec(acc, attractionForce(pos, otherPos, rel.strength)),
        Vec2(0.0, 0.0), myRels);

    addVec(repulsion, attraction)
}
```

### Rendering

```ailang
pure func renderSociogram(state: SociogramState) -> [DrawCmd] {
    let bgCmds = [Panel(50.0, 50.0, 1180.0, 620.0, 1, 7, 40)];

    -- Title and controls
    let headerCmds = [
        Text("Crew Relationships", 540.0, 70.0, 9, 41),
        renderFilterButtons(state.filter)
    ];

    -- Calculate positions
    let positions = calculateForceLayout(state.crew, state.relationships);

    -- Draw edges first (relationships)
    let edgeCmds = renderRelationshipEdges(state.relationships, positions, state.filter);

    -- Draw nodes (crew members)
    let nodeCmds = renderCrewNodes(state.crew, positions, state);

    -- Selected crew detail panel
    let detailCmds = match state.selectedCrew {
        Some(id) => renderCrewRelationshipDetail(id, state),
        None => []
    };

    concat(bgCmds, concat(headerCmds, concat(flatten(edgeCmds), concat(flatten(nodeCmds), detailCmds))))
}

pure func renderRelationshipEdges(rels: [Relationship], positions: [(CrewID, Vec2)], filter: RelationFilter) -> [[DrawCmd]] {
    let filtered = filterRelationships(rels, filter);
    map(\rel.
        let pos1 = findPos(positions, rel.crew1);
        let pos2 = findPos(positions, rel.crew2);
        let color = relationTypeToColor(rel.relationType);
        let thickness = strengthToThickness(rel.strength);
        [Line(pos1.x, pos1.y, pos2.x, pos2.y, color, thickness, 41)],
        filtered)
}

pure func renderCrewNodes(crew: [Crew], positions: [(CrewID, Vec2)], state: SociogramState) -> [[DrawCmd]] {
    map(\c.
        let pos = findPos(positions, c.id);
        let isSelected = state.selectedCrew == Some(c.id);
        let isHovered = state.hoveredCrew == Some(c.id);
        let size = if isSelected then 30.0 else if isHovered then 25.0 else 20.0;
        [
            Circle(pos.x, pos.y, size, if c.alive then 2 else 4, 42),
            if isSelected || isHovered then
                Text(c.name, pos.x, pos.y + size + 5.0, 5, 43)
            else Noop
        ],
        crew)
}
```

---

## 3. Technology Inventory

### Purpose

Shows what technologies the ship has installed and available.

### State

```ailang
type TechInventoryState = {
    installed: [InstalledTech],
    available: [TechID],
    selectedSlot: Maybe(SlotID),
    selectedTech: Maybe(TechID),
    category: TechCategory
}

type InstalledTech = {
    slotID: SlotID,
    tech: Maybe(TechID),
    slotType: SlotType,
    condition: int                   -- 0-100
}

type SlotType =
    | DriveSlot
    | EnergySlot
    | LifeSupportSlot
    | CommsSlot
    | DefenseSlot
    | ArchiveSlot
    | MedicalSlot
    | UtilitySlot

type TechCategory = CatAll | CatDrive | CatEnergy | CatLife | CatComms | CatDefense | CatArchive | CatMedical | CatUtility
```

### Rendering

```ailang
pure func renderTechInventory(state: TechInventoryState) -> [DrawCmd] {
    let bgCmds = [Panel(150.0, 100.0, 980.0, 520.0, 1, 7, 40)];

    -- Ship schematic with slot positions
    let schematicCmds = renderShipSchematic(state.installed);

    -- Available tech list
    let listCmds = renderAvailableTech(state);

    -- Selected tech detail
    let detailCmds = match state.selectedTech {
        Some(id) => renderTechDetail(getTech(id)),
        None => match state.selectedSlot {
            Some(slotID) => renderSlotDetail(findSlot(state.installed, slotID)),
            None => []
        }
    };

    concat(bgCmds, concat(schematicCmds, concat(listCmds, detailCmds)))
}

pure func renderShipSchematic(slots: [InstalledTech]) -> [DrawCmd] {
    -- Top-down ship outline with slot indicators
    let outline = [
        -- Ship body
        Polygon(shipOutlinePoints(), 1, 41),
        Text("SHIP SYSTEMS", 350.0, 120.0, 7, 42)
    ];

    let slotCmds = map(\slot.
        let pos = slotTypeToPosition(slot.slotType);
        let color = if slot.tech == None then 4 else conditionToColor(slot.condition);
        [
            Rect(pos.x, pos.y, 40.0, 40.0, color, 42),
            Text(slotTypeIcon(slot.slotType), pos.x + 12.0, pos.y + 12.0, 6, 43)
        ],
        slots);

    concat(outline, flatten(slotCmds))
}
```

---

## 4. Philosophy Browser

### Purpose

Gallery of all philosophies encountered, their effects, and who follows them.

### State

```ailang
type PhilosophyBrowserState = {
    knownPhilosophies: [Philosophy],
    selectedPhilosophy: Maybe(PhilosophyID),
    viewMode: PhilosophyView,
    compareMode: Maybe(PhilosophyID)
}

type PhilosophyView =
    | ListView
    | GridView
    | TreeView                       -- Show evolution/synthesis
```

### Rendering

```ailang
pure func renderPhilosophyBrowser(state: PhilosophyBrowserState) -> [DrawCmd] {
    let bgCmds = [Panel(100.0, 50.0, 1080.0, 620.0, 1, 7, 40)];

    -- Header
    let headerCmds = [
        Text("Known Philosophies", 540.0, 70.0, 9, 41),
        Text("(" ++ intToString(length(state.knownPhilosophies)) ++ " discovered)", 540.0, 95.0, 6, 41)
    ];

    -- Philosophy cards grid
    let gridCmds = renderPhilosophyGrid(state);

    -- Detail panel (if selected)
    let detailCmds = match state.selectedPhilosophy {
        Some(id) => renderPhilosophyDetail(findPhilosophy(state.knownPhilosophies, id)),
        None => []
    };

    concat(bgCmds, concat(headerCmds, concat(gridCmds, detailCmds)))
}

pure func renderPhilosophyCard(phil: Philosophy, x: float, y: float, selected: bool) -> [DrawCmd] {
    let bgColor = if selected then 3 else 2;
    [
        Panel(x, y, 200.0, 120.0, bgColor, 7, 41),
        Text(phil.name, x + 10.0, y + 10.0, 7, 42),
        TextWrapped(truncate(phil.coreQuestion, 50), x + 10.0, y + 35.0, 180.0, 5, 42),
        renderPhilosophyModifiers(phil.modifiers, x + 10.0, y + 80.0)
    ]
}

pure func renderPhilosophyModifiers(mods: PhilosophyModifiers, x: float, y: float) -> [DrawCmd] {
    let bars = [
        ("S", mods.stabilityBonus),
        ("T", mods.techBonus),
        ("E", mods.expansionBonus),
        ("C", mods.contactBonus)
    ];

    mapWithIndex(\(label, value), i.
        let barX = x + intToFloat(i) * 45.0;
        let barColor = if value >= 0 then 2 else 4;
        let barHeight = abs(value) / 2.0;
        [
            Text(label, barX, y, 4, 43),
            Rect(barX, y + 10.0, 30.0, barHeight, barColor, 43)
        ],
        bars)
}
```

---

## 5. Time Comparison UI

### Purpose

Interactive slider showing the relationship between subjective (ship) time and objective (galaxy) time.

### State

```ailang
type TimeComparisonState = {
    viewMode: TimeViewMode,
    selectedYear: int,               -- Objective year to examine
    showEvents: bool,
    eventFilter: [LogCategory]
}

type TimeViewMode =
    | DualTimeline                   -- Side by side
    | OverlayTimeline                -- Superimposed
    | SliderMode                     -- Interactive scrub
```

### Rendering

```ailang
pure func renderTimeComparison(state: TimeComparisonState) -> [DrawCmd] {
    let bgCmds = [Panel(100.0, 50.0, 1080.0, 620.0, 1, 7, 40)];

    -- Title
    let titleCmds = [Text("Time Comparison", 540.0, 70.0, 9, 41)];

    match state.viewMode {
        DualTimeline => renderDualTimeline(state),
        OverlayTimeline => renderOverlayTimeline(state),
        SliderMode => renderTimeSlider(state)
    }
}

pure func renderDualTimeline(state: TimeComparisonState) -> [DrawCmd] {
    let objectiveY = 150.0;
    let subjectiveY = 450.0;

    -- Objective timeline (galaxy)
    let objCmds = [
        Text("Galaxy Time", 150.0, objectiveY - 20.0, 7, 41),
        Rect(150.0, objectiveY, 980.0, 100.0, 1, 41),
        renderTimelineMarkers(state.events, objectiveY, true)
    ];

    -- Subjective timeline (player)
    let subjCmds = [
        Text("Your Time", 150.0, subjectiveY - 20.0, 7, 41),
        Rect(150.0, subjectiveY, 980.0, 100.0, 2, 41),
        renderTimelineMarkers(state.events, subjectiveY, false)
    ];

    -- Connecting lines showing time dilation
    let connCmds = renderDilationConnections(state);

    concat(objCmds, concat(subjCmds, connCmds))
}

pure func renderDilationConnections(state: TimeComparisonState) -> [DrawCmd] {
    -- Draw lines connecting same events on both timelines
    -- The angle shows how much dilation occurred
    mapWithIndex(\event, i.
        let objX = yearToX(event.objectiveYear);
        let subjX = subjectiveYearToX(event.subjectiveYear);
        let color = if subjX < objX then 2 else 4;  -- Green if time saved, red if lost
        Line(objX, 250.0, subjX, 450.0, color, 1, 42),
        state.events)
}

pure func renderTimeSlider(state: TimeComparisonState) -> [DrawCmd] {
    [
        -- Slider track
        Rect(150.0, 300.0, 980.0, 20.0, 1, 41),

        -- Slider handle
        let handleX = yearToX(state.selectedYear);
        Rect(handleX - 5.0, 290.0, 10.0, 40.0, 3, 42),

        -- Year display
        Text("Year: " ++ formatYear(state.selectedYear), 540.0, 350.0, 8, 42),
        Text("Your age: " ++ formatFloat(yearToSubjectiveAge(state.selectedYear)), 540.0, 380.0, 7, 42),

        -- Events at this year
        renderEventsAtYear(state.selectedYear, state.events)
    ]
}
```

---

## Implementation Plan

### Phase 1: Logbook

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/logbook.go` | LogEntry, LogbookState types |
| 1.2 | `sim_gen/funcs.go` | Auto-entry generation |
| 1.3 | `engine/render/logbook.go` | List and detail rendering |
| 1.4 | Test | Browse log entries |

### Phase 2: Sociogram

| Task | File | Description |
|------|------|-------------|
| 2.1 | `sim_gen/sociogram.go` | Relationship types |
| 2.2 | `sim_gen/funcs.go` | Force layout calculation |
| 2.3 | `engine/render/sociogram.go` | Graph rendering |
| 2.4 | Test | See relationship web |

### Phase 3: Tech Inventory

| Task | File | Description |
|------|------|-------------|
| 3.1 | `sim_gen/inventory.go` | Tech slot types |
| 3.2 | `engine/render/inventory.go` | Ship schematic |
| 3.3 | Test | View installed tech |

### Phase 4: Philosophy Browser

| Task | File | Description |
|------|------|-------------|
| 4.1 | `engine/render/philosophy.go` | Card grid |
| 4.2 | `engine/render/philosophy.go` | Detail panel |
| 4.3 | Test | Browse philosophies |

### Phase 5: Time Comparison

| Task | File | Description |
|------|------|-------------|
| 5.1 | `sim_gen/timecomp.go` | Time mapping |
| 5.2 | `engine/render/timecomp.go` | Timeline rendering |
| 5.3 | Test | Compare timelines |

---

## Success Criteria

### Logbook
- [ ] Entries auto-generated
- [ ] Search and filter work
- [ ] Detail view comprehensive

### Sociogram
- [ ] Layout readable
- [ ] Relationships clear
- [ ] Selection works

### Tech Inventory
- [ ] Slots displayed correctly
- [ ] Install/remove works
- [ ] Condition visible

### Philosophy Browser
- [ ] All discovered shown
- [ ] Details complete
- [ ] Modifiers clear

### Time Comparison
- [ ] Dilation visualized
- [ ] Events synchronized
- [ ] Slider interactive

---

## Future Extensions

| Feature | Description |
|---------|-------------|
| Export logbook | Save as text file |
| Sociogram animation | Watch relationships evolve |
| Tech tree | Show upgrade paths |
| Philosophy synthesis | Predict combinations |
| Time machine | Scrub through history |
