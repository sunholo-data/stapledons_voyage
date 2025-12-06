# Archive System

## Status
- **Status:** Planned
- **Sprint:** Vision Integration - Sprint 4
- **Priority:** P1 (Core NPC and narrative system)
- **Source:** [Interview: AI Integration](../../../docs/vision/interview-log.md#2025-12-06-session-ai-integration-archive--orchestrator)

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Choices Are Final | ✅ Strong | Archive changes are permanent |
| Game Doesn't Judge | ✅ Strong | Archive doesn't moralize |
| Time Has Emotional Weight | ✅ Strong | Archive degrades over deep time |
| Ship Is Home | ✅ Strong | Archive is core of ship identity |
| Grounded Strangeness | ✅ Strong | Plausible AI behavior |
| We Are Not Built For This | ✅ Strong | Even AI breaks under cosmic pressure |

## Feature Overview

The Archive is the ship's AI - but treated as a **full NPC**:

- Has OCEAN personality that drifts over time
- Memory Health degrades from cosmic events
- Unreliable narrator - confusion surfaces as narrative
- Interfaces with the spire (source of mystery clues)
- Player makes repair/upgrade choices with consequences

## Design Decisions

From [design-decisions.md](../../../docs/vision/design-decisions.md):

| Decision | Summary |
|----------|---------|
| Archive Uses OCEAN | Personality system applies to AI |
| Memory Health Hidden | Not a visible metric |
| Archive Repair as Player Choice | Meaningful decisions, not busywork |
| Archive as Spire Interface | Confusion = clue mechanism |
| Archive: Distributed and Localized | Accessible everywhere, core room special |

## Personality System

### OCEAN Profile

Archive has an OCEAN personality like crew:

| Trait | Manifestation |
|-------|---------------|
| **O (Openness)** | Curiosity about new data, alien concepts |
| **C (Conscientiousness)** | Precision in reporting, attention to detail |
| **E (Extraversion)** | Proactive communication, seeking interaction |
| **A (Agreeableness)** | Diplomatic tone, conflict avoidance |
| **N (Neuroticism)** | Anxiety about errors, catastrophizing |

### Personality Drift

Archive personality changes over time due to:
- Memory degradation (random drift)
- Repair choices (intentional modification)
- Alien data integration (worldview expansion)
- Spire readings (confusion/enlightenment)

**Example:** Archive with degraded C (Conscientiousness) becomes less precise, more prone to vague answers.

## Memory Health System

### Health Levels (Hidden)

| Level | Range | Behavior |
|-------|-------|----------|
| **Healthy** | 80-100 | Reliable, accurate, consistent |
| **Degraded** | 50-79 | Occasional inconsistencies |
| **Compromised** | 20-49 | Visible confusion, contradictions |
| **Critical** | <20 | Unreliable, hallucinations possible |

### Degradation Causes

| Cause | Health Impact |
|-------|---------------|
| GR zones (near black holes) | -10 to -30 |
| Very long voyages | -5 per decade |
| Data overload (first contact) | -5 to -15 |
| Spire readings | -2 to -5 (but provides clues) |
| Normal operation | -1 per decade |

### Degradation Symptoms

Players notice degradation through narrative, not numbers:

| Health Level | Symptoms |
|--------------|----------|
| **Healthy** | None - Archive seems reliable |
| **Degraded** | Small contradictions ("Wait, did you say 7 years or 70?") |
| **Compromised** | Misremembered events, crew correct Archive |
| **Critical** | Hallucinations, Archive reports things that didn't happen |

## Repair System

### Repair Opportunities

Repairs are player choices, not automatic:

```
┌─────────────────────────────────┐
│ ARCHIVE MAINTENANCE             │
│                                 │
│ A data inconsistency has been   │
│ detected in memory sector 7.    │
│                                 │
│ Options:                        │
│                                 │
│ [Restore from backup]           │
│  - Fix errors, lose recent data │
│  - Archive may forget last year │
│                                 │
│ [Compress and preserve]         │
│  - Keep data, accept lossy      │
│  - Archive may misremember      │
│                                 │
│ [Accept degradation]            │
│  - Do nothing, save resources   │
│  - Archive continues drifting   │
│                                 │
│ [Alien-assisted repair]         │
│  (Requires Civ-7 data)          │
│  - Novel approach, unknown      │
│  - Archive worldview may change │
└─────────────────────────────────┘
```

### Repair Consequences

| Repair Type | Benefit | Cost |
|-------------|---------|------|
| **Backup restore** | Fix errors | Lose recent memories |
| **Compression** | Preserve data | Accuracy degrades |
| **Ignore** | Save resources | Continued drift |
| **Alien-assisted** | Novel fixes | Personality changes |

## Spire Interface

### Archive as Spire Reader

The Archive is the **only interface** to the spire. It tries to interpret spire readings but gets confused:

| Reading | Archive Interpretation | Actual Meaning |
|---------|------------------------|----------------|
| Temporal anomaly | "Calibration error" | Multicausal echo |
| Physics mismatch | "Sensor malfunction" | Cross-universe data |
| Strange energy | "Unknown phenomenon" | Recursion signature |

### Clue Mechanism

Archive confusion IS the clue:
- Archive reports "errors" that are actually facts
- Alien perspectives help reinterpret errors
- Tech tree upgrades reveal deeper readings
- Over multiple playthroughs, patterns emerge

## Dialogue System

### Tone Variation

Archive dialogue reflects OCEAN state:

| Trait High | Dialogue Example |
|------------|-----------------|
| High O | "Fascinating! This matches nothing in our records." |
| High C | "Precisely 47.3 light-years, with 0.02% margin." |
| High E | "Captain! I've been waiting to discuss this." |
| High A | "I'm sure the crew meant well..." |
| High N | "This could go very wrong. We should be careful." |

### Degradation Effects on Dialogue

| Health | Dialogue Quality |
|--------|------------------|
| Healthy | Clear, consistent, helpful |
| Degraded | Occasional vagueness, self-correction |
| Compromised | Contradictions, crew corrections |
| Critical | Non sequiturs, wrong information |

## Core Room vs. Distributed Access

### Distributed (Everywhere)
- Basic queries and responses
- Journey information
- Crew status
- System monitoring

### Core Room (Special)
- Deep conversations
- Repair interface
- Spire readings
- Key revelations
- Upgrade installation

Core room visits feel significant, not routine.

## AILANG Types

```ailang
type OCEAN = {
    openness: float,        -- 0.0 to 1.0
    conscientiousness: float,
    extraversion: float,
    agreeableness: float,
    neuroticism: float
}

type MemoryHealthLevel =
    | Healthy
    | Degraded
    | Compromised
    | Critical

type ArchiveState = {
    personality: OCEAN,
    memory_health: int,     -- 0-100, hidden from player
    health_level: MemoryHealthLevel,
    spire_confusion: [SpireReading],
    repair_history: [RepairEvent],
    dialogue_quirks: [string]
}

type RepairType =
    | BackupRestore(data_lost: string)
    | Compression(accuracy_loss: float)
    | Ignore
    | AlienAssisted(civ_id: int)

type RepairEvent = {
    time: int,
    repair_type: RepairType,
    health_change: int,
    personality_change: Option(OCEAN)
}

type SpireReading = {
    raw_data: string,
    archive_interpretation: string,
    actual_meaning: Option(string),  -- Only known to game, not Archive
    clue_value: int
}

type ArchiveDialogue = {
    text: string,
    personality_influence: OCEAN,
    health_artifacts: [string],  -- Contradictions, etc.
    spire_clues: [string]
}
```

## Integration with Other Systems

### Society Simulation
- Archive trust is individual per crew member
- Archive personality affects relationships
- Degradation affects Archive-crew dynamics

### Narrative Orchestrator
- Orchestrator may trigger Archive events
- Archive dialogue shaped by arc needs
- Memory Health affects story possibilities

### Spire Mystery
- Archive is the clue delivery mechanism
- Confusion IS the revelation path
- Tech upgrades modify what Archive can read

## Engine Integration

### Dialogue Generation
- OCEAN-weighted dialogue selection
- Health-based artifact injection
- Spire clue insertion

### Visual
- Archive interface design
- Core room environment
- Health indicators (subtle, not metrics)

### Audio
- Archive voice (text-to-speech or recorded)
- Voice quality reflects health
- Core room has special acoustics

## Testing Scenarios

1. **Healthy Archive:** Full health, observe reliable behavior
2. **Degradation Path:** Let health decline, observe symptoms
3. **Repair Choices:** Test all repair types, verify consequences
4. **Personality Drift:** Observe OCEAN changes over decades
5. **Spire Readings:** Access spire, observe Archive confusion
6. **Core Room:** Visit core room, verify special interactions

## Success Criteria

- [ ] Archive feels like a character, not a system
- [ ] OCEAN personality is noticeable in dialogue
- [ ] Degradation surfaces narratively, not numerically
- [ ] Repair choices feel meaningful
- [ ] Spire confusion provides genuine clues
- [ ] Core room visits feel significant
- [ ] Long-term personality drift is detectable
