# Journey System

**Version:** 0.6.0
**Status:** Planned
**Priority:** P0 (Core Gameplay Identity)
**Complexity:** Very High
**AILANG Workarounds:** Float math precision, crew projection recursion
**Depends On:** v0.5.2 Galaxy Map, v0.5.3 Dialogue System

## Related Documents

- [UI Modes Architecture](../v0_5_0/ui-modes.md) - Mode framework
- [Galaxy Map](../v0_5_2/galaxy-map.md) - Journey origin
- [Game Vision](../../../docs/game-vision.md) - Time dilation core mechanic

## Problem Statement

The journey system is the **heart of Stapledon's Voyage**. When players commit to a journey:
- Time passes differently for them vs. the galaxy
- They can't turn back without paying the time cost again
- Civilizations evolve, die, or transcend while they travel
- Their crew ages, forms relationships, has children, dies

This is where the game's central theme manifests: the **loneliness of relativistic travel**.

**Current State:**
- No journey planning UI
- No time dilation calculation
- No crew projection system
- No journey event simulation
- No "during transit" gameplay

**What's Needed:**
- Journey Planning UI with dilation calculator
- The irreversible Commit button
- Journey Events mode (life during transit)
- Crew lifecycle simulation
- Arrival sequence

---

## Design Overview

### Journey Philosophy

Journeys should feel **weighty and consequential**:

- **Planning is deliberate** - See exactly what you're giving up
- **Commitment is irreversible** - The Commit button is a point of no return
- **Transit is alive** - Things happen during the voyage
- **Arrival is bittersweet** - Time has passed, things have changed

### The Time Dilation Experience

```
You travel 10 light-years at 0.99c

Your experience:          The galaxy's experience:
- 1.4 years pass          - 10.1 years pass
- You age 1.4 years       - Everyone else ages 10 years
- 1 crew member dies      - Civilizations evolve
- 2 children born         - Wars are fought
- Relationships deepen    - Technologies advance
                          - Some go extinct
```

### Journey Phases

```
1. PLANNING         → Select destination, choose velocity
2. COMMIT           → Irreversible decision point
3. DEPARTURE        → Leave current star
4. TRANSIT          → Journey Events mode
5. APPROACH         → See destination changes
6. ARRIVAL          → New galaxy map state
```

---

## Part 1: Journey Planning

### Planning State

```ailang
module sim/journey

type JourneyPlanState = {
    destination: StarID,
    destName: string,
    destCiv: Maybe(CivilizationID),
    distance: float,                    -- Light-years

    -- Current slider/selection
    selectedVelocity: float,            -- 0.9 to 0.999999 c
    velocityIndex: int,                 -- Which preset

    -- Calculated values
    lorentzFactor: float,               -- Gamma
    subjectiveYears: float,             -- Time you experience
    objectiveYears: float,              -- Time galaxy experiences
    arrivalYear: int,                   -- Galaxy year on arrival

    -- Crew projections
    crewProjection: CrewProjection,

    -- Player's remaining time
    playerAge: int,
    playerAgeOnArrival: int,
    yearsRemaining: float,              -- Of your 100-year career
    yearsRemainingAfter: float,

    -- UI state
    showAdvancedOptions: bool,
    crewVoteResult: Maybe(VoteResult),
    warnings: [string],
    confirmed: bool                     -- First confirmation step
}

type CrewProjection = {
    startingCrew: int,
    expectedDeaths: int,
    expectedBirths: int,
    endingCrew: int,
    crewDetails: [CrewProjectionDetail],
    relationshipChanges: [RelationshipProjection]
}

type CrewProjectionDetail = {
    crewID: CrewID,
    name: string,
    currentAge: int,
    arrivalAge: int,
    survivalProbability: float,
    willDie: bool,
    causeIfDies: Maybe(DeathCause)
}

type RelationshipProjection = {
    crew1: CrewID,
    crew2: CrewID,
    currentRelation: RelationType,
    projectedRelation: RelationType,
    probability: float
}

type VoteResult = {
    votesFor: int,
    votesAgainst: int,
    abstentions: int,
    concerns: [string],
    strongOpposers: [CrewID]
}
```

### Velocity Presets

```ailang
type VelocityPreset = {
    name: string,
    velocity: float,
    description: string
}

-- Velocity presets for UI
pure func velocityPresets() -> [VelocityPreset] {
    [
        { name: "Cautious",     velocity: 0.9,      description: "Slow but safe. Time dilation: 2.3x" },
        { name: "Standard",     velocity: 0.95,     description: "Balanced approach. Time dilation: 3.2x" },
        { name: "Swift",        velocity: 0.99,     description: "Fast travel. Time dilation: 7.1x" },
        { name: "Rapid",        velocity: 0.999,    description: "Very fast. Time dilation: 22.4x" },
        { name: "Express",      velocity: 0.9999,   description: "Near maximum. Time dilation: 70.7x" },
        { name: "Extreme",      velocity: 0.99999,  description: "Dangerous speed. Time dilation: 223.6x" },
        { name: "Maximum",      velocity: 0.999999, description: "Theoretical limit. Time dilation: 707.1x" }
    ]
}
```

### Time Dilation Calculation

```ailang
-- Lorentz factor: gamma = 1 / sqrt(1 - v^2/c^2)
-- Where v is velocity as fraction of c

pure func lorentzFactor(velocity: float) -> float {
    1.0 / sqrt(1.0 - velocity * velocity)
}

-- Calculate journey time
pure func calculateJourneyTimes(distance: float, velocity: float) -> (float, float) {
    let gamma = lorentzFactor(velocity);
    let objectiveTime = distance / velocity;  -- Years in galaxy frame
    let subjectiveTime = objectiveTime / gamma;  -- Years in ship frame
    (subjectiveTime, objectiveTime)
}

-- Full journey calculation
pure func calculateJourney(galaxy: Galaxy, destID: StarID, velocity: float, crew: [Crew], currentYear: int, playerAge: int) -> JourneyPlanState {
    let dest = findStar(galaxy.stars, destID);
    let current = findStar(galaxy.stars, galaxy.currentPosition);
    let distance = starDistance(current, dest);

    let gamma = lorentzFactor(velocity);
    let (subjective, objective) = calculateJourneyTimes(distance, velocity);

    let arrivalYear = currentYear + floatToInt(objective);
    let playerArrival = playerAge + floatToInt(subjective);

    let projection = projectCrew(crew, subjective);
    let warnings = generateWarnings(projection, subjective, dest, galaxy);

    {
        destination: destID,
        destName: dest.name,
        destCiv: dest.civilization,
        distance: distance,
        selectedVelocity: velocity,
        velocityIndex: velocityToIndex(velocity),
        lorentzFactor: gamma,
        subjectiveYears: subjective,
        objectiveYears: objective,
        arrivalYear: arrivalYear,
        crewProjection: projection,
        playerAge: playerAge,
        playerAgeOnArrival: playerArrival,
        yearsRemaining: 100.0 - intToFloat(playerAge),
        yearsRemainingAfter: 100.0 - intToFloat(playerArrival),
        showAdvancedOptions: false,
        crewVoteResult: None,
        warnings: warnings,
        confirmed: false
    }
}
```

### Crew Projection

```ailang
-- Project crew survival and events during journey
pure func projectCrew(crew: [Crew], years: float) -> CrewProjection {
    let details = map(\c. projectCrewMember(c, years), crew);
    let deaths = length(filter(\d. d.willDie, details));
    let births = estimateBirths(crew, years);
    let relationships = projectRelationships(crew, years);

    {
        startingCrew: length(crew),
        expectedDeaths: deaths,
        expectedBirths: births,
        endingCrew: length(crew) - deaths + births,
        crewDetails: details,
        relationshipChanges: relationships
    }
}

-- Project individual crew member
pure func projectCrewMember(c: Crew, years: float) -> CrewProjectionDetail {
    let arrivalAge = c.age + floatToInt(years);
    let (survives, cause) = calculateSurvival(c, arrivalAge);

    {
        crewID: c.id,
        name: c.name,
        currentAge: c.age,
        arrivalAge: arrivalAge,
        survivalProbability: if survives then 0.95 else 0.3,
        willDie: not(survives),
        causeIfDies: if survives then None else Some(cause)
    }
}

-- Simple survival model (before RNG)
pure func calculateSurvival(c: Crew, arrivalAge: int) -> (bool, DeathCause) {
    -- Deterministic: die if arrival age > 85
    -- With RNG: more nuanced probability curves
    if arrivalAge > 85 then
        (false, OldAge)
    else if arrivalAge > 75 && c.health < 50 then
        (false, Illness)
    else
        (true, OldAge)  -- Cause only used if dies
}

-- Estimate births based on crew composition
pure func estimateBirths(crew: [Crew], years: float) -> int {
    let fertilePairs = countFertilePairs(crew);
    let yearsInt = floatToInt(years);
    -- Rough estimate: 0.1 births per fertile pair per year
    (fertilePairs * yearsInt) / 10
}
```

### Warning Generation

```ailang
pure func generateWarnings(projection: CrewProjection, years: float, dest: Star, galaxy: Galaxy) -> [string] {
    let warnings = [];

    -- Crew warnings
    let warnings = if projection.expectedDeaths > 0 then
        concat(warnings, [formatDeathWarning(projection.expectedDeaths)])
    else warnings;

    let warnings = if projection.endingCrew < 10 then
        concat(warnings, ["WARNING: Crew will be critically low on arrival."])
    else warnings;

    -- Time warnings
    let warnings = if years > 20.0 then
        concat(warnings, ["This journey will consume a significant portion of your career."])
    else warnings;

    let warnings = if years > 50.0 then
        concat(warnings, ["CAUTION: Over half your remaining life will be spent in transit."])
    else warnings;

    -- Destination warnings
    let warnings = match dest.lastKnownState {
        Declining(_) => concat(warnings, ["The civilization at this destination was declining when last observed."]),
        PreContact => concat(warnings, ["This civilization has not yet developed space travel."]),
        _ => warnings
    };

    -- Time since last visit
    let warnings = match dest.lastVisitYear {
        Some(year) => {
            let yearsSince = galaxy.currentYear - year;
            if yearsSince > 1000 then
                concat(warnings, ["Over 1,000 years have passed since this location was last visited."])
            else warnings
        },
        None => concat(warnings, ["This star has never been visited."])
    };

    warnings
}
```

---

## Part 2: The Commit Decision

### Commit Sequence

The Commit button triggers a multi-step confirmation:

```ailang
type CommitPhase =
    | NotStarted
    | FirstConfirm         -- "Are you sure?"
    | CrewVoting           -- Crew expresses opinion
    | FinalConfirm         -- "This is irreversible"
    | Committed            -- Point of no return

pure func processCommitInput(state: JourneyPlanState, input: FrameInput) -> JourneyPlanState {
    if input.commitPressed then
        match state.commitPhase {
            NotStarted => { state | commitPhase: FirstConfirm },
            FirstConfirm => {
                let vote = conductCrewVote(state.crew, state);
                { state | commitPhase: CrewVoting, crewVoteResult: Some(vote) }
            },
            CrewVoting => { state | commitPhase: FinalConfirm },
            FinalConfirm => { state | commitPhase: Committed },
            Committed => state
        }
    else if input.cancelPressed then
        { state | commitPhase: NotStarted }
    else
        state
}
```

### Crew Voting

```ailang
pure func conductCrewVote(crew: [Crew], plan: JourneyPlanState) -> VoteResult {
    let votes = map(\c. crewVotes(c, plan), crew);
    let votesFor = length(filter(\v. v == For, votes));
    let votesAgainst = length(filter(\v. v == Against, votes));
    let abstentions = length(filter(\v. v == Abstain, votes));

    let opposers = filterMap(\c. if crewVotes(c, plan) == Against then Some(c.id) else None, crew);

    let concerns = collectConcerns(crew, plan);

    {
        votesFor: votesFor,
        votesAgainst: votesAgainst,
        abstentions: abstentions,
        concerns: concerns,
        strongOpposers: opposers
    }
}

pure func crewVotes(c: Crew, plan: JourneyPlanState) -> Vote {
    -- Crew vote based on personality and projection
    let detail = findCrewDetail(plan.crewProjection.crewDetails, c.id);

    -- Will they die? Strong opposition
    if detail.willDie then Against
    -- Leaving loved ones? Opposition
    else if hasLovedOneAtDestination(c, plan) then Against
    -- Adventurous? Support
    else if hasTrait(c, Adventurous) then For
    -- Cautious and long journey? Opposition
    else if hasTrait(c, Cautious) && plan.subjectiveYears > 10.0 then Against
    -- Default: abstain
    else Abstain
}
```

### Commit Dialogue

The final confirmation shows everything clearly:

```
┌─────────────────────────────────────────────────────────────────┐
│                    COMMIT TO JOURNEY                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Destination: Tau Ceti III                                       │
│  Distance: 12.3 light-years                                      │
│  Velocity: 0.99c                                                 │
│                                                                  │
│  ════════════════════════════════════════════════════════════   │
│                                                                  │
│  Time for you:        1.7 years                                  │
│  Time for galaxy:     12.4 years                                 │
│                                                                  │
│  You will age:        1.7 years  (47 → 49)                       │
│  Career remaining:    51.3 years → 49.6 years                    │
│                                                                  │
│  ────────────────────────────────────────────────────────────   │
│                                                                  │
│  CREW IMPACT:                                                    │
│  • 2 crew members will likely die during transit                 │
│  • 1 child expected to be born                                   │
│  • Chen strongly opposes this journey                            │
│                                                                  │
│  ════════════════════════════════════════════════════════════   │
│                                                                  │
│  ⚠ THIS DECISION CANNOT BE UNDONE                               │
│                                                                  │
│  When you arrive, 12.4 years will have passed.                   │
│  The civilization you're visiting will have changed.             │
│  Any civilization you leave behind will have evolved.            │
│                                                                  │
│            [ CANCEL ]              [ COMMIT ]                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Part 3: Journey Events Mode

### During Transit

Once committed, the game enters Journey Events mode - a hybrid experience:

```ailang
type JourneyEventState = {
    journeyID: JourneyID,
    origin: StarID,
    destination: StarID,

    -- Progress
    yearsElapsed: float,
    yearsTotal: float,
    progressPercent: float,

    -- Current ship state
    shipState: ShipState,
    currentEvent: Maybe(JourneyEvent),

    -- Voyage log
    voyageLog: [VoyageLogEntry],

    -- Time controls
    timeSpeed: TimeSpeed,
    paused: bool
}

type TimeSpeed = Slow | Normal | Fast | VeryFast

type JourneyEvent =
    | CrewDeath(CrewID, DeathCause)
    | CrewBirth(CrewID, CrewID, string)  -- parents, child name
    | SystemFailure(SystemID, Severity)
    | RelationshipEvent(CrewID, CrewID, RelationshipChange)
    | PhilosophicalDebate(Topic, [CrewID])
    | Discovery(DiscoveryType)
    | CrewMilestone(CrewID, Milestone)
    | ShipEvent(ShipEventType)

type VoyageLogEntry = {
    year: float,                        -- Subjective year of journey
    title: string,
    description: string,
    category: LogCategory,
    involvedCrew: [CrewID]
}
```

### Event Generation

```ailang
-- Generate events during journey (deterministic until RNG)
pure func generateJourneyEvents(journey: JourneyEventState, crew: [Crew], tick: int) -> [JourneyEvent] {
    let year = journey.yearsElapsed;

    -- Check for scheduled deaths
    let deaths = checkScheduledDeaths(crew, year);

    -- Check for births
    let births = checkScheduledBirths(crew, year);

    -- Check for relationship milestones
    let relationships = checkRelationshipEvents(crew, year, tick);

    -- System events (deterministic pattern)
    let systems = checkSystemEvents(journey.shipState, year, tick);

    -- Debates (triggered by year milestones)
    let debates = if isYearMilestone(year) then
        [generateDebate(crew, year)]
    else [];

    concat(concat(concat(deaths, births), relationships), concat(systems, debates))
}

-- Milestone years trigger special events
pure func isYearMilestone(year: float) -> bool {
    let y = floatToInt(year * 10.0);
    y % 50 == 0  -- Every 5 subjective years
}

-- Death check (deterministic: age-based)
pure func checkScheduledDeaths(crew: [Crew], year: float) -> [JourneyEvent] {
    filterMap(\c.
        if shouldDieThisYear(c, year) then
            Some(CrewDeath(c.id, determineDeathCause(c)))
        else None,
        crew)
}

pure func shouldDieThisYear(c: Crew, year: float) -> bool {
    let ageNow = c.age + floatToInt(year);
    -- Deterministic: each crew has a "death year" based on starting health
    let deathYear = 70 + (c.health / 5);  -- Health 0-100 maps to 70-90
    ageNow >= deathYear
}
```

### Journey Event Display

```ailang
-- Render journey event popup
pure func renderJourneyEvent(event: JourneyEvent, state: JourneyEventState) -> [DrawCmd] {
    match event {
        CrewDeath(crewID, cause) => renderDeathEvent(crewID, cause),
        CrewBirth(parent1, parent2, name) => renderBirthEvent(parent1, parent2, name),
        RelationshipEvent(c1, c2, change) => renderRelationshipEvent(c1, c2, change),
        PhilosophicalDebate(topic, crew) => renderDebateEvent(topic, crew),
        SystemFailure(sysID, severity) => renderSystemFailure(sysID, severity),
        _ => []
    }
}

pure func renderDeathEvent(crewID: CrewID, cause: DeathCause) -> [DrawCmd] {
    let crew = findCrew(crewID);
    [
        -- Dim background
        Rect(0.0, 0.0, 1280.0, 720.0, 0, 50),
        -- Memorial panel
        Panel(340.0, 160.0, 600.0, 400.0, 1, 7, 51),
        -- Portrait (dimmed)
        Portrait(390.0, 200.0, 200.0, 250.0, crew.portrait, Grieving, 52),
        -- Text
        Text(crew.name, 620.0, 220.0, 8, 52),
        Text(formatYears(crew.birthYear) ++ " - " ++ formatYears(currentYear), 620.0, 260.0, 7, 52),
        Text(causeToString(cause), 620.0, 300.0, 6, 52),
        TextWrapped(crew.epitaph, 390.0, 480.0, 520.0, 6, 52),
        -- Continue button
        Button(490.0, 520.0, 300.0, 40.0, "Continue Journey", 3, 52)
    ]
}
```

### Journey Progress Display

```ailang
-- Main journey HUD
pure func renderJourneyHUD(state: JourneyEventState) -> [DrawCmd] {
    [
        -- Progress bar
        Panel(50.0, 20.0, 400.0, 60.0, 1, 7, 30),
        Rect(60.0, 40.0, 380.0 * state.progressPercent, 30.0, 2, 30),  -- Fill
        Text("Journey Progress", 60.0, 28.0, 8, 31),
        Text(formatProgress(state), 240.0, 48.0, 8, 31),

        -- Time displays
        Panel(50.0, 90.0, 200.0, 80.0, 1, 7, 30),
        Text("Shipboard Time", 60.0, 100.0, 7, 31),
        Text(formatYears(state.yearsElapsed), 60.0, 130.0, 8, 31),

        Panel(260.0, 90.0, 200.0, 80.0, 1, 7, 30),
        Text("Galaxy Time", 270.0, 100.0, 7, 31),
        Text(formatYears(state.galaxyYearsElapsed), 270.0, 130.0, 8, 31),

        -- Crew count
        Panel(50.0, 180.0, 150.0, 60.0, 1, 7, 30),
        Text("Crew: " ++ intToString(state.crewCount), 60.0, 200.0, 8, 31),

        -- Time controls
        renderTimeControls(state)
    ]
}
```

---

## Part 4: Arrival Sequence

### Arrival State

```ailang
type ArrivalState = {
    destination: StarID,
    journeySummary: JourneySummary,
    civilizationChanges: [CivChange],
    revealPhase: RevealPhase
}

type JourneySummary = {
    yearsElapsed: float,
    galaxyYearsElapsed: float,
    crewDeaths: [(CrewID, string)],     -- Who died and how
    crewBirths: [(CrewID, string)],     -- Who was born
    significantEvents: [string],
    shipDamage: [SystemID]
}

type CivChange = {
    civID: CivilizationID,
    oldState: CivState,
    newState: CivState,
    majorEvents: [string]
}

type RevealPhase =
    | JourneySummaryPhase
    | ShipStatusPhase
    | DestinationRevealPhase
    | GalaxyChangesPhase
    | CompletePhase
```

### Arrival Sequence

```ailang
-- Process arrival sequence
pure func processArrivalSequence(state: ArrivalState, input: FrameInput) -> ArrivalState {
    if input.clicked || input.skipPressed then
        advanceRevealPhase(state)
    else
        state
}

pure func advanceRevealPhase(state: ArrivalState) -> ArrivalState {
    let nextPhase = match state.revealPhase {
        JourneySummaryPhase => ShipStatusPhase,
        ShipStatusPhase => DestinationRevealPhase,
        DestinationRevealPhase => GalaxyChangesPhase,
        GalaxyChangesPhase => CompletePhase,
        CompletePhase => CompletePhase
    };
    { state | revealPhase: nextPhase }
}
```

### Arrival Reveal UI

```
PHASE 1: Journey Summary
┌─────────────────────────────────────────────────────────────────┐
│                    JOURNEY COMPLETE                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  You traveled 12.3 light-years at 0.99c                         │
│                                                                  │
│  For you:   1.7 years passed                                     │
│  For them:  12.4 years passed                                    │
│                                                                  │
│  ────────────────────────────────────────────────────────────   │
│                                                                  │
│  WHAT HAPPENED ABOARD:                                           │
│                                                                  │
│  • Dr. Sarah Chen passed away peacefully (Year 0.8)             │
│  • Marcus and Elena welcomed daughter Maya (Year 1.2)           │
│  • A philosophical debate reshaped crew values                   │
│  • Life support system required emergency repair                 │
│                                                                  │
│                        [ Continue ]                              │
└─────────────────────────────────────────────────────────────────┘

PHASE 2: Destination Reveal
┌─────────────────────────────────────────────────────────────────┐
│                    ARRIVAL: TAU CETI III                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  When you left (Year 3,247):                                     │
│  • Population: 2.3 billion                                       │
│  • Status: Thriving technological society                        │
│  • Last contact: Trade agreement signed                          │
│                                                                  │
│  ════════════════════════════════════════════════════════════   │
│                                                                  │
│  Now (Year 3,259):                                               │
│  • Population: 3.1 billion                                       │
│  • Status: Expanding into nearby systems                         │
│  • Major event: Achieved fusion power (Year 3,252)              │
│                                                                  │
│  They remember you. They've been waiting.                        │
│                                                                  │
│                        [ Continue ]                              │
└─────────────────────────────────────────────────────────────────┘
```

---

## Go/Engine Integration

### Journey Planning Renderer

```go
// engine/render/journey.go

type JourneyPlanRenderer struct {
    fonts    *FontSet
    panels   *NineSlice
    slider   *SliderWidget
}

func (r *JourneyPlanRenderer) Render(screen *ebiten.Image, state JourneyPlanState) {
    // Main panel
    r.panels.Draw(screen, 100, 50, 1080, 620)

    // Title
    r.fonts.DrawTitle(screen, "PLAN JOURNEY TO "+state.DestName, 640, 80)

    // Distance display
    r.drawDistanceInfo(screen, state)

    // Velocity slider
    r.drawVelocitySlider(screen, state)

    // Time calculation display
    r.drawTimeCalculation(screen, state)

    // Crew projection
    r.drawCrewProjection(screen, state)

    // Warnings
    r.drawWarnings(screen, state.Warnings)

    // Commit button
    r.drawCommitButton(screen, state)
}

func (r *JourneyPlanRenderer) drawTimeCalculation(screen *ebiten.Image, state JourneyPlanState) {
    x, y := 600.0, 200.0

    // Subjective time
    r.fonts.DrawLabel(screen, "Your time:", x, y)
    r.fonts.DrawValue(screen, fmt.Sprintf("%.1f years", state.SubjectiveYears), x+150, y)

    // Objective time
    r.fonts.DrawLabel(screen, "Galaxy time:", x, y+30)
    r.fonts.DrawValue(screen, fmt.Sprintf("%.1f years", state.ObjectiveYears), x+150, y+30)

    // Time dilation factor
    r.fonts.DrawLabel(screen, "Dilation:", x, y+60)
    r.fonts.DrawValue(screen, fmt.Sprintf("%.1fx", state.LorentzFactor), x+150, y+60)

    // Arrival year
    r.fonts.DrawLabel(screen, "Arrival year:", x, y+90)
    r.fonts.DrawValue(screen, fmt.Sprintf("%d", state.ArrivalYear), x+150, y+90)
}
```

### Journey Event Input

```go
// engine/input/journey.go

func CaptureJourneyEventInput() FrameInput {
    var input FrameInput

    // Time controls
    if inpututil.IsKeyJustPressed(ebiten.Key1) {
        input.SetTimeSpeed = Slow
    } else if inpututil.IsKeyJustPressed(ebiten.Key2) {
        input.SetTimeSpeed = Normal
    } else if inpututil.IsKeyJustPressed(ebiten.Key3) {
        input.SetTimeSpeed = Fast
    } else if inpututil.IsKeyJustPressed(ebiten.Key4) {
        input.SetTimeSpeed = VeryFast
    }

    // Pause
    if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
        input.TogglePause = true
    }

    // Click to advance event
    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        input.Clicked = true
    }

    // Enter ship mode
    if inpututil.IsKeyJustPressed(ebiten.KeyS) {
        input.OpenShip = true
    }

    return input
}
```

---

## Implementation Plan

### Phase 1: Journey Planning UI

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/journey.go` | JourneyPlanState type |
| 1.2 | `sim_gen/journey.go` | Time dilation calculations |
| 1.3 | `engine/render/journey.go` | Planning panel layout |
| 1.4 | `engine/render/journey.go` | Velocity slider |
| 1.5 | Test | See journey plan UI |

### Phase 2: Crew Projection

| Task | File | Description |
|------|------|-------------|
| 2.1 | `sim_gen/journey.go` | CrewProjection type |
| 2.2 | `sim_gen/funcs.go` | Survival calculation |
| 2.3 | `sim_gen/funcs.go` | Birth estimation |
| 2.4 | `engine/render/journey.go` | Projection display |
| 2.5 | Test | See crew projections |

### Phase 3: Commit Sequence

| Task | File | Description |
|------|------|-------------|
| 3.1 | `sim_gen/journey.go` | CommitPhase enum |
| 3.2 | `sim_gen/funcs.go` | Crew voting logic |
| 3.3 | `engine/render/journey.go` | Confirm dialogs |
| 3.4 | Test | Full commit flow |

### Phase 4: Journey Events Mode

| Task | File | Description |
|------|------|-------------|
| 4.1 | `sim_gen/journey.go` | JourneyEventState type |
| 4.2 | `sim_gen/funcs.go` | Event generation |
| 4.3 | `engine/render/journey.go` | Event popups |
| 4.4 | `engine/render/journey.go` | Journey HUD |
| 4.5 | Test | See events during transit |

### Phase 5: Arrival Sequence

| Task | File | Description |
|------|------|-------------|
| 5.1 | `sim_gen/journey.go` | ArrivalState type |
| 5.2 | `sim_gen/funcs.go` | Civ change calculation |
| 5.3 | `engine/render/journey.go` | Reveal sequence |
| 5.4 | Test | Complete journey cycle |

---

## Testing Strategy

### Manual Testing

```bash
make run-mock
# 1. Navigate to galaxy map
# 2. Right-click star → journey planning
# 3. Adjust velocity slider
# 4. Verify time calculations correct
# 5. Click Commit → see confirmations
# 6. Final commit → enter journey mode
# 7. Watch events unfold
# 8. Speed up time
# 9. Arrival sequence plays
# 10. See changes
```

### Automated Testing

```go
func TestLorentzFactor(t *testing.T) {
    // At 0.99c, gamma ≈ 7.09
    gamma := lorentzFactor(0.99)
    assert.InDelta(t, 7.09, gamma, 0.01)
}

func TestJourneyTimeCalculation(t *testing.T)
func TestCrewProjectionSurvival(t *testing.T)
func TestCommitSequence(t *testing.T)
func TestEventGeneration(t *testing.T)
func TestArrivalSequence(t *testing.T)
```

---

## AILANG Constraints

| Limitation | Impact | Workaround |
|------------|--------|------------|
| Float precision | Large time values | Keep years, not seconds |
| No RNG | Events deterministic | Seed from world state |
| Recursion depth | Long event lists | Process in chunks |
| No mutable state | Event accumulation | Fold pattern |

---

## Success Criteria

### Journey Planning
- [ ] Distance calculation correct
- [ ] Time dilation calculation correct
- [ ] Velocity slider works
- [ ] Crew projection displays

### Commit Flow
- [ ] Multi-step confirmation
- [ ] Crew vote shows
- [ ] Warnings display
- [ ] Final commit works

### Journey Events
- [ ] Events generate during transit
- [ ] Deaths/births occur
- [ ] Time controls work
- [ ] Voyage log populates

### Arrival
- [ ] Journey summary shows
- [ ] Civ changes revealed
- [ ] Galaxy state updated
- [ ] Return to normal play

---

## Future Extensions

| Feature | Description |
|---------|-------------|
| Abort journey | Emergency return (with time cost) |
| In-transit trading | Encounter other ships |
| Dream sequences | Narrative interludes |
| Ship customization | Refit during transit |
| Generational play | Control descendants |
