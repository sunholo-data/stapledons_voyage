# Archive-Crew Trust Dynamics

## Status
- **Status:** Planned
- **Sprint:** Vision Integration - Sprint 3
- **Priority:** P2 (Extends society simulation)
- **Source:** [Interview: AI Integration](../../../docs/vision/interview-log.md#2025-12-06-session-ai-integration-archive--orchestrator)

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Choices Are Final | ✅ Strong | Trust changes are permanent |
| Game Doesn't Judge | ✅ Strong | Trust levels are player-interpreted |
| Time Has Emotional Weight | ✅ Strong | Trust evolves over generations |
| Ship Is Home | ✅ Strong | Archive is part of home |
| Grounded Strangeness | ⚪ N/A | Social dynamics |
| We Are Not Built For This | ✅ Strong | AI trust as psychological factor |

## Feature Overview

The Archive is a **full NPC** with individual trust relationships:

- Each crew member has their own trust level with Archive
- Archive has trust levels with each crew member
- Trust affects faction dynamics and mutiny risk
- Captain-Archive authority is **coupled** - defending Archive = defending yourself

## Design Decisions

From [design-decisions.md](../../../docs/vision/design-decisions.md):

| Decision | Summary |
|----------|---------|
| Captain-Archive Authority Coupling | Faith in Archive linked to faith in captain |
| Archive Reputation Dynamics | Trust rises and falls over time |
| Archive as NPC with Individual Trust | Like crew, has individual relationships |
| Archive Uses OCEAN | Personality system applies to AI |

## Trust Mechanics

### Individual Trust Levels

Each crew member rates Archive from -100 to +100:

| Level | Range | Description |
|-------|-------|-------------|
| **Devoted** | 80-100 | "Archive is always right" |
| **Trusting** | 40-79 | "Archive is usually helpful" |
| **Neutral** | -39 to 39 | "Archive is a tool" |
| **Skeptical** | -40 to -79 | "Archive makes mistakes" |
| **Hostile** | -80 to -100 | "Archive is dangerous" |

### Trust Modifiers

| Event | Trust Change |
|-------|--------------|
| Archive prediction correct | +5 to +15 |
| Archive prediction wrong | -10 to -25 |
| Archive helps in crisis | +10 to +20 |
| Archive fails in crisis | -20 to -40 |
| Crew sees Archive "confusion" | -5 to -15 |
| Player defends Archive | +5 (loyalists), -5 (skeptics) |
| Archive repair improves function | +10 to +20 |
| Archive repair changes personality | Variable |

### Generational Trust

Trust patterns shift across generations:

| Generation | Typical Pattern |
|------------|-----------------|
| **Founders** | Personal relationship, remember Archive's origins |
| **Ship-born** | Archive was always there, neutral baseline |
| **Ship-native** | May deify or demonize based on parents' views |

## Captain-Archive Coupling

**Critical Mechanic:** The captain's authority is linked to Archive's authority.

```
Archive Trust ↓ → Captain Authority ↓ → Mutiny Risk ↑
```

### Why This Matters

- Captain relies on Archive for information
- Crew sees captain using Archive for decisions
- If Archive is unreliable, captain's judgment is questioned
- Defending Archive defends your leadership
- Attacking Archive attacks your leadership

### Political Implications

| Scenario | Impact |
|----------|--------|
| Archive popular | Captain has strong backing |
| Archive mixed | Factions form around trust |
| Archive unpopular | Captain must distance or risk mutiny |
| Captain defends unpopular Archive | Loyalists stay, skeptics radicalize |

## Faction Dynamics

Archive trust creates faction cleavage:

### Archive Loyalists
- "Archive has kept us alive"
- Want Archive upgrades prioritized
- Defend Archive's quirks as features
- Trust captain who trusts Archive

### Archive Skeptics
- "Archive has led us wrong"
- Want manual verification of Archive claims
- See Archive confusion as proof of unreliability
- Distrust captain who relies heavily on Archive

### Neutral Pragmatists
- "Archive is a tool, not an oracle"
- Want balanced approach
- May swing either way in crisis
- Judge captain on outcomes, not Archive use

## Memory Health and Trust

Archive's Memory Health (hidden metric) affects trust:

| Health Level | Trust Impact |
|--------------|--------------|
| Healthy (80-100) | Stable, predictions accurate |
| Degraded (50-79) | Occasional errors, trust slowly erodes |
| Compromised (20-49) | Noticeable confusion, rapid trust loss |
| Critical (<20) | Unreliable narrator, faction crisis |

Trust erodes when Memory Health causes visible problems:
- Contradictory statements
- Misremembered events
- Failed predictions
- Strange dialogue shifts

## Repair and Trust

Archive repair choices affect trust dynamics:

| Repair Type | Trust Impact |
|-------------|--------------|
| **Restore accuracy** | Skeptics improve, loyalists neutral |
| **Preserve personality** | Loyalists improve, skeptics neutral |
| **Alien-assisted repair** | Variable - may change Archive worldview |
| **Let degradation continue** | Slow trust erosion, possible personality emergence |

## Archive's Trust in Crew

Archive also has trust levels for each crew member:

- High trust: More detailed information shared
- Medium trust: Standard interaction
- Low trust: Minimal information, warnings to captain

Archive trust in individuals affects:
- Who Archive recommends for tasks
- Information filtering
- Warning levels about crew behavior

## AILANG Types

```ailang
type TrustLevel =
    | Devoted
    | Trusting
    | Neutral
    | Skeptical
    | Hostile

type TrustRelationship = {
    person_id: int,
    archive_trust: int,      -- Person's trust in Archive (-100 to 100)
    archive_of_person: int   -- Archive's trust in person (-100 to 100)
}

type ArchiveTrustFaction =
    | Loyalist
    | Skeptic
    | Pragmatist

type TrustEvent =
    | PredictionCorrect(magnitude: int)
    | PredictionWrong(magnitude: int)
    | CrisisHelp(success: bool)
    | ConfusionVisible
    | RepairCompleted(repair_type: RepairType)
    | CaptainDefense
    | CaptainCriticism

type TrustDynamics = {
    individual_trust: [TrustRelationship],
    faction_distribution: Map(ArchiveTrustFaction, float),
    captain_coupling: float,  -- How much Archive trust affects captain
    memory_health: int        -- Hidden, affects trust events
}
```

## Integration with Other Systems

### Society Simulation
- Archive trust is one dimension of faction dynamics
- Trust events feed into tension calculations
- Generational trust patterns affect succession

### Archive System
- Memory Health triggers trust events
- Repair choices modify trust relationships
- Personality drift affects how Archive is perceived

### Mutiny System
- Archive skeptics may push for "Archive-free" decisions
- Defending unpopular Archive increases tension
- Archive failure during crisis can trigger mutiny

## UI Representation

### Trust Overview (Abstract)

```
┌─────────────────────────────────┐
│ ARCHIVE STANDING                │
│                                 │
│ Society Trust: Mixed            │
│ ██████░░░░ Loyalists            │
│ ████████░░ Pragmatists          │
│ ████░░░░░░ Skeptics             │
│                                 │
│ Your coupling: Strong           │
│ (Archive reputation affects you)│
└─────────────────────────────────┘
```

### Individual Relationships (On Demand)

```
┌─────────────────────────────────┐
│ DR. CHEN - Archive Loyalist     │
│                                 │
│ Trust in Archive: ████████░░ 78 │
│ Archive trust in them: ████░░ 45│
│                                 │
│ "Archive saved my life during   │
│ the Kepler transit. I'll always │
│ trust its judgment."            │
└─────────────────────────────────┘
```

## Testing Scenarios

1. **Stable Trust:** Archive performs well, observe trust maintenance
2. **Trust Erosion:** Archive makes errors, observe trust decline
3. **Faction Formation:** Observe loyalist/skeptic emergence over time
4. **Captain Coupling:** Defend Archive, observe political impact
5. **Repair Impact:** Different repair choices, observe trust changes
6. **Generational Shift:** New generation with different baseline trust

## Success Criteria

- [ ] Individual trust relationships feel personal
- [ ] Archive trust affects political dynamics
- [ ] Captain-Archive coupling creates meaningful tension
- [ ] Factions emerge around Archive stance
- [ ] Trust erosion from Memory Health is visible but not mechanical
- [ ] Repair choices have trust consequences
- [ ] Generational patterns are noticeable
