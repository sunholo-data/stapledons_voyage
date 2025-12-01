# Exploration Modes: Planet Surface & Ruins

**Version:** 0.7.0
**Status:** Planned
**Priority:** P1 (Environmental Storytelling)
**Complexity:** Medium
**AILANG Workarounds:** Tile-based movement, artifact discovery
**Depends On:** v0.5.0 UI Modes, v0.6.0 Journey System

## Related Documents

- [UI Modes Architecture](../v0_5_0/ui-modes.md) - Mode framework
- [Ship Exploration](../v0_5_1/ship-exploration.md) - Similar movement system
- [Civilization Detail](../v0_6_1/civilization-detail.md) - Planet context
- [Game Vision](../../../docs/game-vision.md) - Environmental storytelling

## Problem Statement

Players need to experience civilizations directly, not just through data screens:

- Landing on thriving worlds to meet locals
- Exploring ruins of extinct civilizations
- Discovering artifacts and ancient logs
- Environmental storytelling across deep time

**Current State:**
- No planetary exploration
- No surface rendering
- No artifact system
- No ruins mechanics

**What's Needed:**
- Planet Surface mode for living civilizations
- Ruins mode for extinct civilizations
- Artifact discovery and collection
- Ancient log reading system
- Environmental variation by civ state

---

## Part 1: Planet Surface Mode

### Surface State

```ailang
module sim/surface

type PlanetSurfaceState = {
    planetID: PlanetID,
    zoneID: LandingZoneID,
    playerPos: Coord,
    playerFacing: Direction,
    visibleArea: Rect,

    -- Zone data
    zoneTiles: [Tile],
    zoneWidth: int,
    zoneHeight: int,
    zoneType: ZoneType,

    -- Entities
    localNPCs: [LocalNPC],
    interactables: [SurfaceInteractable],

    -- Civilization context
    civState: CivState,
    civID: CivilizationID,

    -- Interaction
    hoveredEntity: Maybe(EntityID),
    activeDialogue: Maybe(DialogueState),

    -- Collection
    discoveredItems: [ItemID]
}

type ZoneType =
    | SpaceportZone         -- Where you land
    | CityZone              -- Urban area
    | NaturalZone           -- Wilderness
    | IndustrialZone        -- Factories, tech
    | SacredZone            -- Religious/cultural
    | ResidentialZone       -- Where they live
    | RuinsZone             -- Abandoned area

type LocalNPC = {
    id: LocalNPCID,
    pos: Coord,
    species: Species,
    role: LocalRole,
    dialogueTree: Maybe(ConversationID),
    disposition: Disposition,
    sprite: SpriteID
}

type LocalRole =
    | Official              -- Government representative
    | Scholar               -- Shares knowledge
    | Merchant              -- Offers trade
    | Citizen               -- Everyday person
    | Elder                 -- Wisdom, history
    | Child                 -- Curiosity, hope
    | Guard                 -- Security
    | Outcast               -- Alternative views

type Disposition = Friendly | Neutral | Wary | Hostile | Curious
```

### Zone Generation

```ailang
-- Generate landing zone based on civilization state
pure func generateZone(planet: Planet, civ: Civilization, zoneType: ZoneType) -> Zone {
    let baseTiles = generateBaseTerrain(planet.climate, zoneType);
    let structures = generateStructures(civ, zoneType);
    let npcs = generateLocalNPCs(civ, zoneType);
    let interactables = generateInteractables(civ, zoneType);

    {
        tiles: overlayStructures(baseTiles, structures),
        width: 32,
        height: 32,
        npcs: npcs,
        interactables: interactables,
        ambience: determineAmbience(civ.currentState)
    }
}

-- Structure density varies by civ state
pure func generateStructures(civ: Civilization, zoneType: ZoneType) -> [Structure] {
    let density = match civ.currentState {
        Thriving(_) => 0.7,
        Declining(_) => 0.5,
        PreContact => 0.3,
        _ => 0.1
    };

    let structureTypes = match zoneType {
        SpaceportZone => [Hangar, ControlTower, Terminal, FuelDepot],
        CityZone => [Building, Monument, Plaza, Garden],
        IndustrialZone => [Factory, Warehouse, Silo, Pipe],
        SacredZone => [Temple, Shrine, Altar, Meditation],
        _ => [BasicStructure]
    };

    distributeStructures(structureTypes, density, 32, 32)
}

-- NPC generation based on zone and civ state
pure func generateLocalNPCs(civ: Civilization, zoneType: ZoneType) -> [LocalNPC] {
    let count = match civ.currentState {
        Thriving(_) => 8,
        Declining(_) => 4,
        PreContact => 3,
        _ => 0
    };

    let roles = zoneToRoles(zoneType);
    generateNPCsWithRoles(civ.species, roles, count)
}
```

### Surface Interaction

```ailang
-- Process surface mode input
pure func processSurfaceInput(state: PlanetSurfaceState, input: FrameInput) -> PlanetSurfaceState {
    -- Handle active dialogue first
    match state.activeDialogue {
        Some(dialogue) => processDialogueInSurface(state, dialogue, input),
        None => {
            let afterMove = processSurfaceMovement(state, input);
            let afterHover = processSurfaceHover(afterMove, input);
            let afterInteract = processSurfaceInteract(afterHover, input);
            afterInteract
        }
    }
}

-- Movement on planetary surface
pure func processSurfaceMovement(state: PlanetSurfaceState, input: FrameInput) -> PlanetSurfaceState {
    let dir = inputToDirection(input);
    match dir {
        None => state,
        Some(d) => {
            let newPos = moveInDirection(state.playerPos, d);
            if isSurfaceWalkable(newPos, state.zoneTiles, state.zoneWidth) then
                { state | playerPos: newPos, playerFacing: d }
            else
                { state | playerFacing: d }
        }
    }
}

-- Interact with surface entity
pure func processSurfaceInteract(state: PlanetSurfaceState, input: FrameInput) -> PlanetSurfaceState {
    if input.interactPressed then
        match state.hoveredEntity {
            None => state,
            Some(entityID) => {
                let entity = findSurfaceEntity(state, entityID);
                match entity {
                    NPCEntity(npc) => {
                        match npc.dialogueTree {
                            Some(convID) => {
                                let dialogue = initLocalDialogue(npc, convID);
                                { state | activeDialogue: Some(dialogue) }
                            },
                            None => state  -- NPC has nothing to say
                        }
                    },
                    InteractableEntity(obj) => processSurfaceObject(state, obj),
                    _ => state
                }
            }
        }
    else
        state
}

-- Process interaction with surface object
pure func processSurfaceObject(state: PlanetSurfaceState, obj: SurfaceInteractable) -> PlanetSurfaceState {
    match obj.interactType {
        Artifact(artID) => {
            -- Collect artifact
            { state | discoveredItems: artID :: state.discoveredItems }
        },
        InfoTerminal(content) => {
            -- Display info panel
            { state | activePanel: Some(InfoPanel(content)) }
        },
        TransportPad(destZone) => {
            -- Move to different zone
            transitionToZone(state, destZone)
        },
        ReturnToShip => {
            -- Signal mode transition
            { state | requestReturn: true }
        }
    }
}
```

### Surface Rendering

```ailang
pure func renderPlanetSurface(state: PlanetSurfaceState) -> [DrawCmd] {
    -- Sky/background based on planet type
    let bgCmds = renderSurfaceBackground(state);

    -- Tiles
    let tileCmds = renderSurfaceTiles(state);

    -- Structures (sorted by Y for depth)
    let structCmds = renderStructures(state);

    -- NPCs
    let npcCmds = renderLocalNPCs(state);

    -- Player
    let playerCmds = renderSurfacePlayer(state);

    -- Interaction highlights
    let highlightCmds = renderSurfaceHighlights(state);

    -- UI overlay
    let uiCmds = renderSurfaceUI(state);

    -- Dialogue overlay if active
    let dialogueCmds = match state.activeDialogue {
        Some(d) => renderDialogue(d),
        None => []
    };

    concatAll([bgCmds, tileCmds, structCmds, npcCmds, playerCmds, highlightCmds, uiCmds, dialogueCmds])
}
```

---

## Part 2: Ruins Mode

### Ruins State

```ailang
type RuinsState = {
    siteID: RuinsSiteID,
    siteName: string,
    extinctCiv: CivilizationID,
    extinctionYear: int,
    extinctionCause: string,

    -- Zone data
    playerPos: Coord,
    zoneTiles: [Tile],
    zoneWidth: int,
    zoneHeight: int,

    -- Discovery
    artifacts: [RuinsArtifact],
    logs: [AncientLog],
    discoveredArtifacts: [ArtifactID],
    decipheredLogs: [LogID],

    -- Exploration progress
    explorationProgress: float,
    areasRevealed: [AreaID],

    -- Active UI
    activePanel: Maybe(RuinsPanel),
    activeLog: Maybe(AncientLog)
}

type RuinsArtifact = {
    id: ArtifactID,
    pos: Coord,
    name: string,
    description: string,
    discovered: bool,
    value: int,
    category: ArtifactCategory
}

type ArtifactCategory =
    | Technology             -- Devices, tools
    | Art                    -- Cultural expression
    | Record                 -- Data storage
    | Religious              -- Sacred objects
    | Personal               -- Individual items
    | Scientific             -- Research equipment

type AncientLog = {
    id: LogID,
    pos: Coord,              -- Where found
    originalDate: int,       -- When written
    author: string,
    title: string,
    content: string,
    deciphered: bool,
    difficulty: int,         -- Decipherment difficulty
    emotionalTone: Emotion,
    category: LogCategory
}

type LogCategory =
    | Personal               -- Diary, letters
    | Official               -- Government records
    | Scientific             -- Research notes
    | Religious              -- Sacred texts
    | Warning                -- Messages to future
    | Art                    -- Poetry, stories
```

### Ruins Generation

```ailang
-- Generate ruins based on extinct civilization
pure func generateRuins(civ: Civilization, extinctionYear: int, cause: string) -> RuinsSite {
    let decay = calculateDecay(civ.gameYear - extinctionYear);

    let tiles = generateRuinsTerrain(civ, decay);
    let artifacts = generateRuinsArtifacts(civ, decay);
    let logs = generateAncientLogs(civ, extinctionYear, cause);

    {
        id: generateSiteID(),
        name: civ.name ++ " Ruins",
        extinctCiv: civ.id,
        extinctionYear: extinctionYear,
        extinctionCause: cause,
        tiles: tiles,
        artifacts: artifacts,
        logs: logs,
        totalArea: 4096,
        exploredArea: 0
    }
}

-- Decay affects what remains
pure func calculateDecay(years: int) -> float {
    -- 0.0 = pristine, 1.0 = completely eroded
    let base = intToFloat(years) / 10000.0;
    min(1.0, base)
}

-- Generate terrain with decay
pure func generateRuinsTerrain(civ: Civilization, decay: float) -> [Tile] {
    let structureDensity = civ.technology / 100;
    let remainingDensity = structureDensity * (1.0 - decay);

    -- More advanced civs leave more durable ruins
    -- Decay reduces what's left
    generateDecayedTerrain(remainingDensity, civ.species.physiology)
}

-- Generate artifacts based on civ characteristics
pure func generateRuinsArtifacts(civ: Civilization, decay: float) -> [RuinsArtifact] {
    let baseCount = 10 + civ.technology / 10;
    let survivingCount = floatToInt(intToFloat(baseCount) * (1.0 - decay * 0.8));

    let techArtifacts = if civ.technology > 50 then
        generateTechArtifacts(civ, survivingCount / 3)
    else [];

    let artArtifacts = generateArtArtifacts(civ, survivingCount / 3);
    let personalArtifacts = generatePersonalArtifacts(civ, survivingCount / 3);

    concat(techArtifacts, concat(artArtifacts, personalArtifacts))
}

-- Generate logs that tell the story
pure func generateAncientLogs(civ: Civilization, extinctionYear: int, cause: string) -> [AncientLog] {
    [
        -- Everyday life before
        {
            id: 1,
            pos: randomPos(),
            originalDate: extinctionYear - 100,
            author: "Unknown Citizen",
            title: "A Day in the City",
            content: generateDailyLifeLog(civ),
            deciphered: false,
            difficulty: 20,
            emotionalTone: Happy,
            category: Personal
        },
        -- Growing concerns
        {
            id: 2,
            pos: randomPos(),
            originalDate: extinctionYear - 20,
            author: "Council Member",
            title: "Official Warning",
            content: generateWarningLog(civ, cause),
            deciphered: false,
            difficulty: 40,
            emotionalTone: Fearful,
            category: Official
        },
        -- The end
        {
            id: 3,
            pos: randomPos(),
            originalDate: extinctionYear - 1,
            author: "Last Witness",
            title: "To Those Who Come After",
            content: generateFinalLog(civ, cause),
            deciphered: false,
            difficulty: 60,
            emotionalTone: Grieving,
            category: Warning
        }
    ]
}
```

### Log Content Generation

```ailang
-- Generate log content based on civilization traits
pure func generateDailyLifeLog(civ: Civilization) -> string {
    let intro = match civ.species.physiology {
        Humanoid => "Today I walked through the market district...",
        Crystalline => "The harmonic vibrations of morning resonated through the lattice...",
        Gaseous => "The currents carried me to the Gathering...",
        Collective => "The hive-mind consensus formed around the day's tasks...",
        _ => "Another cycle began..."
    };

    let middle = match civ.philosophy.name {
        "The Long View" => "We discussed the millennial plans, as always. Patience is our greatest virtue.",
        "The Burning" => "The pace of progress never slows. Every day brings new achievements.",
        "Sacred Mortality" => "We honored those who have passed, knowing our own time will come.",
        _ => "Life continued as it always has."
    };

    let conclusion = "I wonder what tomorrow will bring. I wonder if we will remember these simple days.";

    intro ++ " " ++ middle ++ " " ++ conclusion
}

pure func generateWarningLog(civ: Civilization, cause: string) -> string {
    let intro = "To all citizens of " ++ civ.name ++ ":";

    let warning = match cause {
        "Internal Collapse" => "Our stability metrics have fallen to critical levels. The divisions among us grow wider each day.",
        "External Threat" => "The threat we face may be beyond our capacity to resist. We must prepare for the worst.",
        "Resource Depletion" => "Our world can no longer sustain us. The calculations are undeniable.",
        "Philosophical Crisis" => "The foundations of our society are crumbling. We no longer agree on what matters.",
        "Technological Catastrophe" => "Our greatest achievement has become our greatest danger. We built too much, too fast.",
        _ => "A crisis approaches. We must act, or we will not survive."
    };

    intro ++ " " ++ warning
}

pure func generateFinalLog(civ: Civilization, cause: string) -> string {
    "To whoever finds this:\n\n" ++
    "We were the " ++ civ.name ++ ". We asked ourselves: '" ++ civ.philosophy.coreQuestion ++ "'\n\n" ++
    "We thought we had found the answer. Perhaps we did. Perhaps the answer simply wasn't enough.\n\n" ++
    "Remember us. Learn from us. Do not repeat our mistakes.\n\n" ++
    "We believed in something. That should count for something.\n\n" ++
    "Farewell."
}
```

### Ruins Exploration

```ailang
-- Process ruins exploration
pure func processRuinsInput(state: RuinsState, input: FrameInput) -> RuinsState {
    match state.activeLog {
        Some(log) => {
            -- Reading a log
            if input.closePressed then
                { state | activeLog: None }
            else
                state
        },
        None => {
            let afterMove = processRuinsMovement(state, input);
            let afterDiscover = checkDiscoveries(afterMove);
            let afterInteract = processRuinsInteract(afterDiscover, input);
            afterInteract
        }
    }
}

-- Check for new discoveries at current position
pure func checkDiscoveries(state: RuinsState) -> RuinsState {
    -- Check artifacts
    let nearbyArtifacts = filter(\a. distance(a.pos, state.playerPos) < 2, state.artifacts);
    let newlyDiscovered = filter(\a. not(a.discovered), nearbyArtifacts);

    if length(newlyDiscovered) > 0 then
        let first = head(newlyDiscovered);
        { state |
            discoveredArtifacts: first.id :: state.discoveredArtifacts,
            activePanel: Some(ArtifactDiscovery(first)) }
    else
        -- Check logs
        let nearbyLogs = filter(\l. distance(l.pos, state.playerPos) < 2, state.logs);
        let newLogs = filter(\l. not(l.deciphered), nearbyLogs);

        if length(newLogs) > 0 then
            let first = head(newLogs);
            { state | activeLog: Some(first) }
        else
            state
}

-- Interact with ruins object
pure func processRuinsInteract(state: RuinsState, input: FrameInput) -> RuinsState {
    if input.interactPressed then
        match state.activePanel {
            Some(ArtifactDiscovery(artifact)) => {
                -- Collect artifact
                { state |
                    activePanel: None,
                    explorationProgress: state.explorationProgress + 0.1 }
            },
            _ => state
        }
    else
        state
}
```

### Ruins Rendering

```ailang
pure func renderRuins(state: RuinsState) -> [DrawCmd] {
    -- Atmosphere: muted, somber colors
    let bgCmds = renderRuinsBackground(state);

    -- Decayed tiles
    let tileCmds = renderRuinsTiles(state);

    -- Standing structures (broken)
    let structCmds = renderRuinsStructures(state);

    -- Artifact indicators (glowing points)
    let artifactCmds = renderArtifactIndicators(state);

    -- Player
    let playerCmds = renderRuinsPlayer(state);

    -- Discovery panel if active
    let panelCmds = match state.activePanel {
        Some(panel) => renderRuinsPanel(panel),
        None => []
    };

    -- Log reading overlay if active
    let logCmds = match state.activeLog {
        Some(log) => renderLogReading(log),
        None => []
    };

    -- Exploration progress HUD
    let hudCmds = renderRuinsHUD(state);

    concatAll([bgCmds, tileCmds, structCmds, artifactCmds, playerCmds, panelCmds, logCmds, hudCmds])
}

pure func renderLogReading(log: AncientLog) -> [DrawCmd] {
    [
        -- Dim background
        Rect(0.0, 0.0, 1280.0, 720.0, 0, 50),

        -- Parchment-style panel
        Panel(200.0, 100.0, 880.0, 520.0, 1, 7, 51),

        -- Header
        Text(log.title, 320.0, 130.0, 10, 52),
        Text("Written by " ++ log.author, 320.0, 165.0, 6, 52),
        Text("Year " ++ intToString(log.originalDate), 750.0, 165.0, 6, 52),

        -- Divider
        Rect(250.0, 190.0, 780.0, 2.0, 7, 52),

        -- Content (wrapped)
        TextWrapped(log.content, 250.0, 210.0, 780.0, 7, 52),

        -- Close hint
        Text("[Press ESC to close]", 540.0, 590.0, 5, 52)
    ]
}
```

---

## Go/Engine Integration

### Surface Renderer

```go
// engine/render/surface.go

type SurfaceRenderer struct {
    tilesets    map[ClimateType]*ebiten.Image
    npcSprites  *SpriteSheet
    structSprites *SpriteSheet
}

func (r *SurfaceRenderer) Render(screen *ebiten.Image, state PlanetSurfaceState) {
    // Background gradient based on sky
    r.drawSkyGradient(screen, state.Climate)

    // Terrain tiles
    r.drawTerrain(screen, state)

    // Structures (Y-sorted)
    r.drawStructures(screen, state)

    // NPCs
    for _, npc := range state.LocalNPCs {
        r.drawLocalNPC(screen, npc, state)
    }

    // Player
    r.drawSurfacePlayer(screen, state)

    // Interaction prompts
    if state.HoveredEntity != nil {
        r.drawInteractionPrompt(screen, state.HoveredEntity)
    }

    // UI
    r.drawSurfaceUI(screen, state)
}
```

### Ruins Renderer

```go
// engine/render/ruins.go

type RuinsRenderer struct {
    decayedTiles *ebiten.Image
    artifactGlow *ebiten.Image
    fonts        *FontSet
}

func (r *RuinsRenderer) Render(screen *ebiten.Image, state RuinsState) {
    // Somber background
    r.drawRuinsAtmosphere(screen, state.Decay)

    // Decayed terrain
    r.drawDecayedTerrain(screen, state)

    // Ruins structures
    r.drawRuins(screen, state)

    // Artifact glows (undiscovered only)
    for _, artifact := range state.Artifacts {
        if !artifact.Discovered {
            r.drawArtifactGlow(screen, artifact)
        }
    }

    // Player
    r.drawRuinsPlayer(screen, state)

    // Discovery overlay
    if state.ActivePanel != nil {
        r.drawDiscoveryPanel(screen, state.ActivePanel)
    }

    // Log reading
    if state.ActiveLog != nil {
        r.drawLogReading(screen, state.ActiveLog)
    }

    // Progress HUD
    r.drawExplorationProgress(screen, state)
}

func (r *RuinsRenderer) drawRuinsAtmosphere(screen *ebiten.Image, decay float64) {
    // Desaturated, misty atmosphere
    overlay := ebiten.NewImage(screenWidth, screenHeight)

    // Grayish overlay
    grayLevel := uint8(30 + decay*30)
    overlay.Fill(color.RGBA{grayLevel, grayLevel, grayLevel + 10, 80})

    screen.DrawImage(overlay, nil)
}
```

---

## Implementation Plan

### Phase 1: Surface Basics

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/surface.go` | PlanetSurfaceState type |
| 1.2 | `sim_gen/surface.go` | Zone generation |
| 1.3 | `engine/render/surface.go` | Basic terrain rendering |
| 1.4 | Test | See planet surface |

### Phase 2: Surface Interaction

| Task | File | Description |
|------|------|-------------|
| 2.1 | `sim_gen/funcs.go` | Surface movement |
| 2.2 | `sim_gen/surface.go` | Local NPC types |
| 2.3 | `sim_gen/funcs.go` | NPC interaction |
| 2.4 | Test | Walk around, talk to NPCs |

### Phase 3: Ruins Basics

| Task | File | Description |
|------|------|-------------|
| 3.1 | `sim_gen/ruins.go` | RuinsState type |
| 3.2 | `sim_gen/ruins.go` | Ruins generation |
| 3.3 | `engine/render/ruins.go` | Decayed terrain |
| 3.4 | Test | See ruins render |

### Phase 4: Artifact System

| Task | File | Description |
|------|------|-------------|
| 4.1 | `sim_gen/ruins.go` | Artifact types |
| 4.2 | `sim_gen/funcs.go` | Discovery mechanic |
| 4.3 | `engine/render/ruins.go` | Artifact glow |
| 4.4 | `engine/render/ruins.go` | Discovery panel |
| 4.5 | Test | Find and collect artifacts |

### Phase 5: Ancient Logs

| Task | File | Description |
|------|------|-------------|
| 5.1 | `sim_gen/ruins.go` | AncientLog type |
| 5.2 | `sim_gen/ruins.go` | Log content generation |
| 5.3 | `engine/render/ruins.go` | Log reading UI |
| 5.4 | Test | Read ancient logs |

### Phase 6: Integration

| Task | File | Description |
|------|------|-------------|
| 6.1 | `sim_gen/funcs.go` | Mode transitions |
| 6.2 | `sim_gen/funcs.go` | Inventory integration |
| 6.3 | Test | Full exploration flow |

---

## Success Criteria

### Planet Surface
- [ ] Surface renders based on civ state
- [ ] Player can move around zone
- [ ] NPCs present and interactive
- [ ] Can return to ship

### Ruins
- [ ] Ruins render with decay
- [ ] Artifacts discoverable
- [ ] Logs readable
- [ ] Emotional storytelling effective

### Integration
- [ ] Artifacts add to inventory
- [ ] Logs add to logbook
- [ ] Exploration progress tracks
- [ ] Mode transitions smooth

---

## Future Extensions

| Feature | Description |
|---------|-------------|
| Multiple zones | Explore entire planet |
| Combat | Hostile encounters |
| Puzzles | Artifact unlocking |
| Excavation | Dig for buried items |
| AI logs | Generated content |
| Photo mode | Capture moments |
