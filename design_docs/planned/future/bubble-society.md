# Bubble Society Simulation

## Status
- **Status:** Planned
- **Sprint:** Vision Integration - Sprint 3
- **Priority:** P1 (Core internal pressure mechanic)
- **Source:** [Interview: Game Loop Origin](../../../docs/vision/interview-log.md#2025-12-06-session-game-loop-origin--bubble-constraint)

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Choices Are Final | ✅ Strong | Society evolves irreversibly |
| Game Doesn't Judge | ✅ Strong | Outcomes are emergent, not scored |
| Time Has Emotional Weight | ✅ Strong | Generations pass, people change |
| Ship Is Home | ✅ Strong | Society IS home |
| Grounded Strangeness | ⚪ N/A | Social sim, not physics |
| We Are Not Built For This | ✅ Strong | Psychological fragility is theme |

## Feature Overview

The bubble is not just a ship - it's a **living micro-civilization**:

- Starts with ~100 people
- Multi-generational: births, deaths, succession
- Autonomous: player influences, doesn't control
- Factions form around competing values
- Journey decisions create stress that reshapes society

**Key Principle:** The captain influences; the society lives.

## Design Decisions

From [design-decisions.md](../../../docs/vision/design-decisions.md):

| Decision | Summary |
|----------|---------|
| Bubble Society as Living Sim | Autonomous with births, deaths, factions |
| Internal Tension Model | Long-term collapse risk + short-term crises |
| Game End Conditions | Death or Mutiny |
| Mutiny Warnings Visible | Fair but tragic - you see tension building |
| Bubble Ship Continues After Death | Your legacy shapes their future |
| Captain Succession Exists | Society elects new captain after your death |
| Population Dynamics | Growth and decline based on circumstances |

## Population System

### Starting State
- ~100 founding crew members
- Mixed ages, skills, personalities
- Each with OCEAN profile
- Initial faction seeds (shared experiences, values)

### Growth
- **Births:** Require mass, morale, and social stability
- **Rate:** ~2-4% per year under good conditions
- **Constraints:** Mass budget limits sustainable population

### Decline
- **Deaths:** Age, accidents, crises
- **Emigration:** N/A (bubble is sealed)
- **Crisis casualties:** Some events kill crew

### Generational Turnover
Over 100 years (captain's lifespan):
- Generation 1: Founders (remember Earth)
- Generation 2: Ship-born (Earth is stories)
- Generation 3: Ship-native (Earth is myth)

Each generation has different relationship with Earth, Archive, captain authority.

## Faction System

### Faction Formation

Factions emerge from:
- **Shared values:** Similar OCEAN profiles cluster
- **Shared experiences:** Survived crisis together
- **Ideological alignment:** Response to journey decisions
- **Resource competition:** Who gets the mass?

### Example Factions

| Faction | Core Value | Archive Stance | Journey Stance |
|---------|------------|----------------|----------------|
| **Progressives** | Expansion, risk | Trust Archive | Push further |
| **Preservers** | Stability, safety | Skeptical | Stay conservative |
| **Earthers** | Save Earth priority | Neutral | Mission-focused |
| **Wanderers** | Exploration | Use Archive | Seek new civs |
| **Isolationists** | Self-sufficiency | Distrust | Avoid contact |

### Faction Dynamics

- Factions grow/shrink based on events
- Cross-faction relationships exist
- Major decisions shift faction power
- Faction imbalance increases tension

## Mutiny System

### Tension Accumulation

| Factor | Tension Impact |
|--------|----------------|
| Bad journey outcome | +5-20 |
| Crew death | +10-30 |
| Resource shortage | +5-15 |
| Ignored advice | +5-10 |
| Archive failure | +10-20 |
| Against majority faction | +15-25 |
| Successful journey | -10-20 |
| Good outcome | -5-15 |
| Resource surplus | -5-10 |
| Listened to advice | -3-5 |

### Warning Signs

Players see tension building through:
- Crew dialogue shifts ("Captain, people are worried...")
- Faction meeting frequency increases
- Direct confrontations from faction leaders
- Archive reports unrest metrics (if trusted)

### Mutiny Trigger

When tension exceeds threshold AND majority faction opposes captain:
1. Confrontation event
2. Player can try to defuse (often fails)
3. Mutiny vote
4. **Game Over** if deposed

**Key Decision:** Mutiny is game over, not a setback. Your story ends.

## Society States

### Stability Spectrum

| State | Description | Risk |
|-------|-------------|------|
| **Thriving** | High morale, cooperation, growth | Low |
| **Stable** | Functional, minor tensions | Low |
| **Stressed** | Visible tension, cautious | Medium |
| **Fractured** | Faction conflict, declining morale | High |
| **Crisis** | Near-collapse, mutiny imminent | Critical |

### Meaning Crisis

Long-term existential risk:
- **Trigger:** Extended isolation, repeated failures, generational drift
- **Symptoms:** Nihilism, apathy, self-destructive behavior
- **Counter:** Successful contact, Earth progress, cultural rituals

## Cultural Evolution

### Rituals
Society develops rituals over time:
- **Founding Day:** Commemorate departure
- **Arrival Celebrations:** When reaching destinations
- **Memorial Days:** For lost crew
- **Garden Ceremonies:** Earth remembrance

Rituals emerge from player decisions and events. They provide stability and meaning.

### Values Drift

Over generations, society values shift:
- First gen: Earth-centric, mission-focused
- Second gen: Ship-centric, journey-focused
- Third gen: Universe-centric, exploration-focused

This creates tension between generations.

## Governance

### Captain's Authority

The captain (player) has:
- **Final say** on journey decisions
- **Influence** on resource allocation
- **No control** over faction dynamics
- **Reputation** that rises and falls

### Succession

After captain's death:
1. Society elects new captain from candidates
2. Player's legacy influences who wins
3. Gameplay ends, but projection shows future

### Limits of Authority

Captain cannot:
- Force faction dissolution
- Prevent births/relationships
- Mandate belief systems
- Override mutiny vote

## Open Questions

From [open-questions.md](../../../docs/vision/open-questions.md):

1. **How many generations in 100 years?** (2-3 typical)
2. **Do children inherit OCEAN tendencies?** (Partial, with drift)
3. **How do factions form and evolve?** (Emergent from events/values)
4. **What triggers meaning-crisis?** (Isolation, failure, drift)
5. **How does Archive interact with society?** (See archive-crew-trust.md)
6. **UI visibility?** (Consequences visible, mechanics hidden)

## AILANG Types

```ailang
type Person = {
    id: int,
    name: string,
    age: int,
    generation: int,
    ocean: OCEAN,
    faction: Option(FactionId),
    relationships: [Relationship],
    status: PersonStatus
}

type PersonStatus =
    | Alive
    | Dead(cause: string)
    | Incapacitated(reason: string)

type Faction = {
    id: FactionId,
    name: string,
    core_value: string,
    members: [int],         -- Person IDs
    power: float,           -- 0.0-1.0
    captain_support: float  -- -1.0 to 1.0
}

type SocietyState =
    | Thriving
    | Stable
    | Stressed
    | Fractured
    | Crisis

type BubbleSociety = {
    population: [Person],
    factions: [Faction],
    state: SocietyState,
    tension: float,
    rituals: [Ritual],
    generation_count: int,
    captain_support: float
}

type MutinyCheck = {
    tension_level: float,
    majority_faction_opposes: bool,
    trigger_threshold: float,
    result: MutinyResult
}

type MutinyResult =
    | NoRisk
    | WarningGiven
    | Confrontation
    | MutinyVote(success: bool)
```

## Engine Integration

### Simulation
- Run society sim each game tick
- OCEAN drift over time
- Faction power recalculation
- Tension adjustment

### Events
- Society state triggers events
- Faction events based on power shifts
- Milestone events (generations, anniversaries)

### UI
- Society overview panel
- Faction power display (abstract)
- Tension indicator
- Key relationship summaries

### Dialogue
- Faction-appropriate dialogue options
- Crew reactions reflect society state
- Archive reports on society (if trusted)

## Testing Scenarios

1. **Stable Society:** Make safe choices, observe thriving society
2. **Faction Conflict:** Favor one faction repeatedly, observe tension
3. **Near-Mutiny:** Push tension high, observe warnings
4. **Actual Mutiny:** Trigger mutiny, verify game over
5. **Generational Shift:** Play 50 years, observe value drift
6. **Meaning Crisis:** Long isolation, observe psychological decline

## Success Criteria

- [ ] Society feels alive and autonomous
- [ ] Player influence is meaningful but not controlling
- [ ] Factions emerge naturally from gameplay
- [ ] Tension builds visibly before mutiny
- [ ] Mutiny is fair (warnings given) but final
- [ ] Generational differences are noticeable
- [ ] Cultural evolution reflects journey history
