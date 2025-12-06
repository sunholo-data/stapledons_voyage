# Narrative Orchestrator

## Status
- **Status:** Planned
- **Sprint:** Vision Integration - Sprint 4
- **Priority:** P2 (Backend system, invisible to player)
- **Source:** [Interview: AI Integration](../../../docs/vision/interview-log.md#2025-12-06-session-ai-integration-archive--orchestrator)

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Choices Are Final | ⚪ N/A | Orchestrator respects player choice |
| Game Doesn't Judge | ✅ Strong | Shapes tension, not morality |
| Time Has Emotional Weight | ✅ Strong | Pacing emotional beats |
| Ship Is Home | ✅ Strong | Internal drama shaping |
| Grounded Strangeness | ⚪ N/A | Meta-system |
| We Are Not Built For This | ✅ Strong | Creates psychological pressure |

## Feature Overview

The Narrative Orchestrator (M-NARRATOR) is a **behind-the-scenes DM**:

- Invisible to player - they never know it exists
- Shapes tension, pacing, and thematic arcs
- Selects events from pools to create coherent stories
- Monitored via developer logs for tuning

**Key Decision:** Pure backend. Player experiences emergent-feeling narrative without seeing machinery.

## Design Decisions

From [design-decisions.md](../../../docs/vision/design-decisions.md):

| Decision | Summary |
|----------|---------|
| Narrative Orchestrator: Behind the Scenes | Player never sees it, devs monitor via logs |

From [open-questions.md](../../../docs/vision/open-questions.md):
- Arc types TBC (not yet decided)
- Event families TBC (starting points identified)

## Core Responsibilities

### 1. Tension Management

Monitor and adjust dramatic tension:

| Metric | Purpose |
|--------|---------|
| **Overall tension** | Is the story engaging? |
| **Faction tension** | Internal conflict level |
| **External stakes** | Universe-level drama |
| **Personal stakes** | Individual crew drama |

**Goal:** Maintain engaging tension without constant crisis.

### 2. Pacing

Ensure rhythm of gameplay:

| Beat Type | Frequency | Purpose |
|-----------|-----------|---------|
| **Quiet** | Regular | Rest, relationship building |
| **Rising** | Building | Anticipation, preparation |
| **Crisis** | Periodic | High stakes decisions |
| **Resolution** | After crisis | Consequences, processing |

### 3. Thematic Coherence

Ensure events connect to themes:

- Time dilation consequences
- Human frailty in cosmic context
- Relationship vs. mission tension
- Archive reliability questions
- Earth salvation progress/failure

### 4. Arc Selection

Shape longer narrative arcs across the journey.

## Arc Types (TBC)

These are starting points, not final designs:

| Arc Type | Duration | Focus |
|----------|----------|-------|
| **Faction Rise** | 10-20 years | One faction gains/loses power |
| **Archive Crisis** | 5-10 years | Archive reliability questioned |
| **Exploration Mystery** | Variable | Strange discovery unfolds |
| **Relationship Arc** | 5-15 years | Key crew relationship evolves |
| **Generation Shift** | 20-30 years | Power transition between generations |
| **Earth Progress** | Game-long | Salvation mission progress |

**Note:** Arc types need design work. These are illustrative.

## Event Taxonomy

### Event Families (Starting Points)

From [input/ai-the-archive.md](../input/ai-the-archive.md):

| Family | Examples | Frequency |
|--------|----------|-----------|
| **Discovery** | New star system, alien contact, anomaly | Journey-based |
| **Internal** | Faction conflict, relationship crisis, mutiny warning | Tension-based |
| **Archive** | Memory degradation, spire reading, repair opportunity | Health-based |
| **Resource** | Mass budget crisis, population pressure | State-based |
| **Personal** | Crew milestone, death, birth, relationship | Time-based |
| **External** | Alien response, cosmic event, time dilation consequence | Journey-based |
| **Mystery** | Spire clue, recursion hint, universe-hopper trace | Rare |

**Note:** Event families need prioritization for MVP.

## Event Selection Logic

### Candidacy

Events become candidates when:
- Preconditions met (state requirements)
- Not in cooldown (recent events excluded)
- Arc appropriate (fits current arc)
- Tension appropriate (matches desired tension level)

### Weighting

Candidate events weighted by:

| Factor | Weight Influence |
|--------|------------------|
| **Arc fit** | High - events should advance current arc |
| **Tension need** | Medium - balance drama level |
| **Time since similar** | Medium - variety matters |
| **Theme relevance** | Low - background consideration |
| **Randomness** | Low - prevents predictability |

### Selection

From weighted candidates:
1. Filter by hard requirements
2. Weight by soft preferences
3. Add randomness factor
4. Select with weighted random

## Tension Model

### Tension Sources

| Source | Contribution |
|--------|--------------|
| Faction conflict | 0-30 |
| Resource pressure | 0-20 |
| External threats | 0-30 |
| Personal crises | 0-15 |
| Archive problems | 0-15 |
| Journey stress | 0-20 |

### Tension Targets

| Phase | Target Range | Notes |
|-------|--------------|-------|
| **Early game** | 20-40 | Learning, establishing |
| **Mid game** | 40-60 | Building stakes |
| **Late game** | 50-80 | High drama, resolution |
| **Post-crisis** | 20-40 | Recovery period |

### Tension Adjustment

If tension too low:
- Introduce complications
- Escalate existing conflicts
- Surface hidden problems

If tension too high:
- Provide resolution opportunities
- Create quieter events
- Delay new complications

## Developer Monitoring

### Log Output

```
[ORCHESTRATOR] Tick 4752
  Current arc: FactionRise(Progressives)
  Tension: 47/100 (target: 45-55)
  Last event: PersonalCrisis(Chen, RelationshipStrain)

  Candidates evaluated: 12
  Selected: InternalEvent(FactionMeeting, Progressives)
  Reason: Arc advancement, tension maintenance

  Arc progress: 65% (transition at 80%)
  Next arc candidates: [ArchiveCrisis, ExplorationMystery]
```

### Tuning Parameters

Developers can adjust:
- Tension targets per phase
- Event family cooldowns
- Arc transition thresholds
- Randomness factors

### Analytics

Track over playthroughs:
- Average tension curves
- Event distribution
- Arc completion rates
- Player engagement correlations

## AILANG Types

```ailang
type ArcType =
    | FactionRise(faction_id: int)
    | ArchiveCrisis
    | ExplorationMystery(mystery_id: int)
    | RelationshipArc(person_a: int, person_b: int)
    | GenerationShift
    | EarthProgress

type EventFamily =
    | Discovery
    | Internal
    | Archive
    | Resource
    | Personal
    | External
    | Mystery

type TensionSource =
    | FactionConflict(amount: int)
    | ResourcePressure(amount: int)
    | ExternalThreat(amount: int)
    | PersonalCrisis(amount: int)
    | ArchiveProblem(amount: int)
    | JourneyStress(amount: int)

type OrchestratorState = {
    current_arc: Option(ArcType),
    arc_progress: float,
    tension: int,
    tension_target: (int, int),  -- min, max
    recent_events: [EventRecord],
    cooldowns: Map(EventFamily, int)
}

type EventCandidate = {
    event: GameEvent,
    family: EventFamily,
    weight: float,
    arc_fit: float,
    tension_impact: int
}

type OrchestratorDecision = {
    tick: int,
    candidates: [EventCandidate],
    selected: Option(GameEvent),
    reason: string
}
```

## Integration with Other Systems

### Society Simulation
- Reads faction state, tension levels
- Triggers faction events
- Respects society constraints

### Archive System
- Reads Archive health
- Triggers Archive events
- Coordinates spire revelations

### Journey System
- Reads journey state
- Triggers discovery events
- Coordinates external encounters

### Player Choices
- **Never overrides** player decisions
- Reacts to player choices
- Shapes context around choices

## Constraints

### Must Not
- Override player agency
- Create impossible situations
- Force specific outcomes
- Be detectable by player

### Must
- Respect game state
- Create coherent narratives
- Maintain playable tension
- Log decisions for debugging

## Testing Scenarios

1. **Tension Recovery:** High tension, observe orchestrator create calm
2. **Arc Progression:** Monitor arc advancement over decades
3. **Event Variety:** Verify no event family dominates
4. **Player Choice Respect:** Make choices, verify orchestrator adapts
5. **Log Analysis:** Verify logs are useful for tuning

## Success Criteria

- [ ] Player cannot detect orchestrator's presence
- [ ] Narratives feel emergent, not scripted
- [ ] Tension is engaging but not exhausting
- [ ] Arcs create coherent story threads
- [ ] Developer logs enable effective tuning
- [ ] Player choices are never overridden
