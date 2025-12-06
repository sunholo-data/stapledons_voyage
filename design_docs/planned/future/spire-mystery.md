# Spire Mystery & Tech Tree

## Status
- **Status:** Planned
- **Sprint:** Vision Integration - Sprint 4
- **Priority:** P3 (Late-game revelation system)
- **Source:** [Interview: Bubble Ship Design](../../../docs/vision/interview-log.md#2025-12-06-session-bubble-ship-design-integration)

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Choices Are Final | ⚪ N/A | Mystery, not choice |
| Game Doesn't Judge | ✅ Strong | Discovery is player's to interpret |
| Time Has Emotional Weight | ⚪ N/A | Revelation, not time |
| Ship Is Home | ✅ Strong | Spire is heart of ship |
| Grounded Strangeness | ✅ Strong | Physics mystery |
| We Are Not Built For This | ✅ Strong | Beyond human understanding |

## Feature Overview

The Higgs Generator Spire is the **source of mystery**:

- May be constant across all universes in recursion loop
- Archive interfaces with it but gets "confused"
- Tech tree progression reveals deeper clues
- Never fully explained - preserves mystery

## Design Decisions

From [design-decisions.md](../../../docs/vision/design-decisions.md):

| Decision | Summary |
|----------|---------|
| The Spire as Universal Constant | Same across all universes, explains uniqueness |
| Archive as Spire Interface | Archive confusion = clue mechanism |

From [open-questions.md](../../../docs/vision/open-questions.md):

| Question | Status |
|----------|--------|
| Spire tech tree reveals | Options outlined, TBD |
| How explicit revelations become | Never fully explicit |

## The Mystery

### What the Spire IS

The spire:
- Creates and maintains the Higgs bubble
- Cannot be replicated (no one else has this tech)
- Predates the ship itself (was... always there?)
- Produces readings that don't match this universe's physics

### Why It's Constant

Across the recursion loop:
- The spire is the same structure in every universe
- This explains why it can't be replicated - it's not FROM any universe
- Archive confusion comes from readings that reference multiple realities
- This is the deepest clue to "you are not the first"

### The Revelation Path

| Stage | What Player Knows | How Revealed |
|-------|-------------------|--------------|
| **Early** | "The spire powers the ship" | Tutorial |
| **Mid** | "Archive has errors reading spire" | Archive dialogue |
| **Late** | "Readings don't match physics" | Tech tree upgrade |
| **Endgame** | "This might be from... outside" | Alien perspective |
| **NG+** | Patterns across playthroughs | Player inference |

## Tech Tree Integration

### Spire-Related Upgrades

| Tier | Upgrade | Clue Revealed |
|------|---------|---------------|
| **1** | Sensor Calibration | "Calibration improves elsewhere, not spire" |
| **2** | Deep Scan | "Readings suggest temporal anomaly" |
| **3** | Archive Integration | "AI can now sense... but misunderstands" |
| **4** | Alien Interpretive Framework | "What Archive calls errors are... facts" |
| **5** | Physical Access (partial) | "The geometry is wrong. Impossible." |

### Upgrade Costs

Spire upgrades cost significant resources:
- Mass from budget
- Archive processing capacity
- Crew focus (takes specialists)
- Time to implement

### Gating

Higher tiers require:
- Lower tier completion
- Specific alien contact (for frameworks)
- Archive above certain health
- Society stability (for physical access)

## Clue Taxonomy

### Archive Confusion

Archive reports "errors" that are clues:

| Error Type | Actual Meaning |
|------------|----------------|
| "Temporal calibration failure" | Readings from multiple timelines |
| "Physics constant mismatch" | Different universe's physics |
| "Energy signature unknown" | Pre-universe energy |
| "Self-referential error" | Structure references itself across universes |
| "Causality violation warning" | Effect precedes cause (cross-universal) |

### Alien Perspectives

Different civilizations interpret spire differently:

| Civ Type | Interpretation | Clue Value |
|----------|----------------|------------|
| **Scientific** | "Anomaly, need more data" | Low |
| **Mystical** | "Sacred, do not question" | Low |
| **Ancient** | "We have legends of travelers" | High |
| **Universe-hoppers** | "We recognize this" | Maximum |

### Physical Clues

If player reaches physical access:
- Impossible angles
- Surfaces that shouldn't connect
- Materials that don't exist in this universe
- Faint echoes of other places

## Discovery Experience

### First Confusion (Year 10-30)

Archive: "Captain, I'm detecting calibration errors in spire readings. Likely sensor degradation. I'll compensate."

Player: Thinks it's Archive problem.

### Growing Mystery (Year 30-50)

Archive: "The spire readings are... inconsistent. The errors don't follow expected patterns. It's as if the spire predates its own construction."

Player: Starts wondering.

### Alien Input (Year 50-70)

Alien: "Your Archive is not malfunctioning. Those readings are accurate. We do not understand them either, but they are not errors."

Player: "Wait, what?"

### Revelation (Year 70-100 or NG+)

Archive (upgraded): "I have reprocessed the spire data using the [Alien] framework. What I interpreted as errors are... readings from elsewhere. Somewhere that exists in the same way across different... I don't have words for this."

Player: Understands the recursion.

## Never Fully Explicit

**Key Design Principle:** The mystery is preserved.

Even at maximum revelation:
- Archive can't fully explain
- Aliens don't fully understand
- Physical access shows impossibility
- Player fills in gaps

This serves:
- Replayability (discover more each time)
- Wonder (some things are beyond understanding)
- Pillar 6 (we are not built for this)

## AILANG Types

```ailang
type SpireAccessLevel =
    | None
    | BasicScans
    | DeepScans
    | ArchiveIntegration
    | AlienFramework
    | PhysicalAccess

type SpireClue = {
    id: int,
    content: string,
    archive_interpretation: string,
    actual_meaning: string,  -- For game logic, never shown directly
    access_required: SpireAccessLevel,
    discovered: bool
}

type SpireMystery = {
    access_level: SpireAccessLevel,
    clues_discovered: [SpireClue],
    alien_frameworks: [CivId],
    archive_confusion_events: int,
    revelation_progress: float  -- 0.0 to 1.0
}

type TechTreeNode = {
    id: int,
    name: string,
    tier: int,
    cost_mass: float,
    cost_time: int,
    prerequisites: [int],
    spire_clue: Option(SpireClue),
    unlocked: bool,
    installed: bool
}
```

## Integration with Other Systems

### Archive System
- Archive health affects spire readings
- Confusion events tracked
- Upgrades modify Archive interpretation

### Proto-tech System
- Spire upgrades use proto-tech mechanics
- Alien frameworks require contact
- Mass costs apply

### Recursion System
- Spire IS the recursion mechanism
- Clues connect to NG+ understanding
- Physical access may hint at BH entry point

## Visual Design

### Spire Appearance
- Looks wrong - angles don't quite add up
- Materials shimmer in impossible ways
- Color seems to shift in peripheral vision
- Gets stranger at higher access levels

### Archive Interface
- Readings show with "error" annotations
- As access increases, "errors" become "anomalies" then "data"
- Visual representation of the incomprehensible

## Testing Scenarios

1. **Early Game:** Observe initial Archive confusion about spire
2. **Tech Progression:** Unlock upgrades, verify clues reveal
3. **Alien Framework:** Acquire perspective, see reinterpretation
4. **Physical Access:** Reach highest tier, experience impossibility
5. **NG+ Patterns:** Multiple playthroughs, verify accumulating understanding

## Success Criteria

- [ ] Spire feels mysterious from the start
- [ ] Archive confusion is intriguing, not frustrating
- [ ] Tech tree progression reveals meaningful clues
- [ ] Alien perspectives add genuine insight
- [ ] Mystery is preserved even at max revelation
- [ ] Multiple playthroughs deepen understanding
- [ ] Visual design supports impossibility feeling
